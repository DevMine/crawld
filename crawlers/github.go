// Copyright 2014-2015 The DevMine authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crawlers

import (
	"database/sql"
	"errors"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/DevMine/crawld/config"
)

// apiCallFunc is the default prototype a function that calls the GitHub API
// must have. This is necessary because API calls are wrapped into a function
// that checks if the API call rate limit is reached or not and waits before
// doing the call again if the limit is reached.
type apiCallFunc func(args ...interface{}) (interface{}, error)

// gitHubCrawler implements the Crawler interface.
type gitHubCrawler struct {
	config.CrawlerConfig

	client *github.Client
	db     *sql.DB
}

// ensure that gitHubCrawler implements the Crawler interface
var _ Crawler = (*gitHubCrawler)(nil)

// implement the oauth2.TokenSource interface
type tokenSource struct {
	AccessToken string
}

func (ts *tokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: ts.AccessToken,
	}
	return token, nil
}

// newGitHubCrawler creates a new GitHub crawler.
func newGitHubCrawler(cfg config.CrawlerConfig, db *sql.DB) (*gitHubCrawler, error) {
	if db == nil {
		return nil, errors.New("database session cannot be nil")
	}

	var httpClient *http.Client
	if len(strings.Trim(cfg.OAuthAccessToken, " ")) != 0 {
		ts := &tokenSource{
			AccessToken: cfg.OAuthAccessToken,
		}
		httpClient = oauth2.NewClient(context.TODO(), ts)
	}
	client := github.NewClient(httpClient)

	return &gitHubCrawler{cfg, client, db}, nil
}

// Crawl implements the Crawl() method of the Crawler interface.
func (g *gitHubCrawler) Crawl() {
	if g.UseSearchAPI {
		for _, lang := range g.Languages {
			_ = g.call(true, g.fetchTopRepositories, lang)
		}
	} else {
		_ = g.call(false, g.fetchRepositories)
	}
}

// call shall be used when doing a query on the GitHub API. If the query is
// refused, typically because the rate limit is reached, then this function
// waits for the appropriate time before retrying the query.
// isSearchRequest shall be used to indicate if apiCallFunc calls the search API
// (rate limit for the search API differ from the core API).
func (g *gitHubCrawler) call(isSearchRequest bool, fct apiCallFunc, args ...interface{}) interface{} {
	var ret interface{}
	var err error

	// gotta wait if rate limit is exceeded
	for {
		if ret, err = fct(args...); err != errTooManyCall {
			break
		}

		var reset int64
		limits, _, _ := g.client.RateLimits()
		if isSearchRequest {
			reset = limits.Search.Reset.Unix()
		} else {
			reset = limits.Core.Reset.Unix()
		}
		waitTime := reset - time.Now().Unix() + 1
		glog.Infof("not enough API calls left => waiting for %d minutes and %d seconds",
			waitTime/60, waitTime%60)
		time.Sleep(time.Duration(waitTime) * time.Second)
	}

	return ret
}

// fetchRepositories fetches N GitHub repositories in the given
// language (if provided).
//
// Warning: This method does not use the search API, thus, it uses a lot of API
// calls.
//
// args expects no argument.
//
// TODO add doc => the limit N is global to all languages
func (g *gitHubCrawler) fetchRepositories(args ...interface{}) (interface{}, error) {
	if len(args) != 0 {
		glog.Error("invalid number of arguments")
		return nil, errInvalidArgs
	}

	n := g.Limit

	keepFork := g.Fork
	hasLimit := n > 0

	// GitHub lists repositories 100 per page, regardless of the per_page option...
	opt := &github.RepositoryListAllOptions{}

	sinceID := g.SinceID
ResultsLoop:
	for {
		opt.Since = sinceID
		repos, resp, err := g.client.Repositories.ListAll(opt)
		if err != nil {
			glog.Error(err)
			return nil, g.genAPICallFuncError(resp, err)
		}

		if len(repos) == 0 {
			break
		}

		for _, repo := range repos {
			if repo.ID == nil {
				glog.Error("'repo' has nil ID field")
				continue
			}
			sinceID = *repo.ID

			if n == 0 && hasLimit {
				break ResultsLoop
			}

			if repo.Fork == nil {
				glog.Error("'repo' has nil Fork field")
				continue
			}
			// skip? fork repos
			if *repo.Fork && !keepFork {
				continue
			}

			if ok, err := isLanguageWanted(g.Languages, repo.Language); err != nil {
				glog.Error(err)
				continue
			} else if !ok {
				langs := g.call(false, g.fetchRepositoryLanguages, *repo.Owner.Login, *repo.Name)

				if ok, err := isLanguageWanted(g.Languages, langs); err != nil {
					glog.Error(err)
					continue
				} else if !ok {
					continue
				}
			}

			var fullRepo *github.Repository
			tmpRepo := g.call(false, g.fetchRepository, *repo.Owner.Login, *repo.Name)
			switch tmpRepo.(type) {
			case *github.Repository:
				fullRepo = tmpRepo.(*github.Repository)
				err = verifyRepo(fullRepo)
				if err != nil {
					glog.Error(err)
					continue
				}
			default:
				glog.Error("invalid fetched repository")
				continue
			}

			// skip when an the method fail because the repository is not
			// saved into the DB
			if !g.insertOrUpdateRepo(fullRepo) {
				continue
			}

			n--
		}

		if n <= 0 && hasLimit {
			break
		}
	}
	return nil, nil
}

// fetchTopRepositories fetches top N GitHub repositories in the given
// language (if provided).
//
// Warning: This method uses the search API, thus it cannot fetch more than
// 1000 results.
//
// args expects 1 values:
//   - language: string indicating the programming language to limit the fetch
// Be very careful if you do not specify a limit and/or a programming language.
//
// TODO add doc => the limit N is for language separately
func (g *gitHubCrawler) fetchTopRepositories(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		glog.Error("invalid number of arguments")
		return nil, errInvalidArgs
	}

	n := g.Limit

	var lang string
	switch args[0].(type) {
	case string:
		lang = args[0].(string)
	default:
		glog.Errorf("invalid parameter type (given %v, expected string)", reflect.TypeOf(args[0]))
		return nil, errInvalidParamType
	}

	keepFork := g.Fork
	hasLimit := n > 0

	opt := &github.SearchOptions{Sort: "stars", ListOptions: github.ListOptions{PerPage: 100}}

ResultsLoop:
	for {
		results, resp, err := g.client.Search.Repositories(
			"language:"+lang, opt)
		if err != nil {
			glog.Error(err)
			return nil, g.genAPICallFuncError(resp, err)
		}

		repos := results.Repositories

		for _, repo := range repos {
			if n == 0 && hasLimit {
				break ResultsLoop
			}

			err = verifyRepo(&repo)
			if err != nil {
				glog.Error(err)
				continue
			}

			// skip? fork repos
			if *repo.Fork && !keepFork {
				continue
			}

			// skip when an the method fail because the repository is not
			// saved into the DB
			if !g.insertOrUpdateRepo(&repo) {
				continue
			}

			n--
		}

		if resp.NextPage == 0 || (n <= 0 && hasLimit) {
			break
		}

		opt.Page = resp.NextPage
	}
	return nil, nil
}

// fetchRepositoryLanguages fetches all languages related to a repository
// args expects 2 values:
// - owner: the repository owner
// - rpeo: the repository name
//
// It returns a map of languages (map[string]int, language => num bytes)
func (g *gitHubCrawler) fetchRepositoryLanguages(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		glog.Error("invalid number of arguments")
		return nil, errInvalidArgs
	}

	var owner string
	switch args[0].(type) {
	case string:
		owner = args[0].(string)
	default:
		glog.Errorf("invalid parameter type (given %v, expected string)", reflect.TypeOf(args[0]))
		return nil, errInvalidParamType
	}

	var repo string
	switch args[1].(type) {
	case string:
		repo = args[1].(string)
	default:
		glog.Errorf("invalid parameter type (given %v, expected string)", reflect.TypeOf(args[1]))
		return nil, errInvalidParamType
	}

	langs, resp, err := g.client.Repositories.ListLanguages(owner, repo)
	if err != nil {
		glog.Error(err)
		return nil, g.genAPICallFuncError(resp, err)
	}

	return langs, nil
}

// fetchRepository fetches the information about a specific repository.
//
// args expects 2 values:
// - owner: the repository owner
// - rpeo: the repository name
//
// It returns a github.Repository
func (g *gitHubCrawler) fetchRepository(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		glog.Error("invalid number of arguments")
		return nil, errInvalidArgs
	}

	var owner string
	switch args[0].(type) {
	case string:
		owner = args[0].(string)
	default:
		glog.Errorf("invalid parameter type (given %v, expected string)", reflect.TypeOf(args[0]))
		return nil, errInvalidParamType
	}

	var repo string
	switch args[1].(type) {
	case string:
		repo = args[1].(string)
	default:
		glog.Errorf("invalid parameter type (given %v, expected string)", reflect.TypeOf(args[1]))
		return nil, errInvalidParamType
	}

	ghRepo, resp, err := g.client.Repositories.Get(owner, repo)
	if err != nil {
		glog.Error(err)
		return nil, g.genAPICallFuncError(resp, err)
	}

	return ghRepo, nil
}

// getRepoID returns the repository id of repo in repositories table.
// If repo is not in the table, then 0 is returned. If an error occurs, -1 is returned.
func (g *gitHubCrawler) getRepoID(repo *github.Repository) int {
	if repo == nil {
		glog.Error("'repo' arg given is nil")
		return -1
	}

	var id int
	err := g.db.QueryRow("SELECT repository_id FROM gh_repositories WHERE github_id=$1", repo.ID).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return 0
	case err != nil:
		glog.Error(err)
		return -1
	}
	return id
}

// getGhRepoID returns the github repository id of repo in repositories table.
// If repo is not in the table, then 0 is returned. If an error occurs, -1 is returned.
func (g *gitHubCrawler) getGhRepoID(repo *github.Repository) int {
	if repo == nil {
		glog.Error("'repo' arg given is nil")
		return -1
	}

	var id int
	err := g.db.QueryRow("SELECT id FROM gh_repositories WHERE github_id=$1", repo.ID).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return 0
	case err != nil:
		glog.Error(err)
		return -1
	}
	return id
}

// getGhOrgID returns the github organization id of org in gh_organizations table.
// If org is not in the table, then 0 is returned. If an error occurs, -1 is returned.
func (g *gitHubCrawler) getGhOrgID(org *github.Organization) int {
	if org == nil {
		glog.Error("'org' arg given is nil")
		return -1
	}

	var id int
	err := g.db.QueryRow("SELECT id FROM gh_organizations WHERE github_id=$1", org.ID).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return 0
	case err != nil:
		glog.Error(err)
		return -1
	}
	return id
}

// getGhUserID returns the github user id of user in gh_users table.
// If user not in the table, then 0 is returned. If an error occurs, -1 is returned.
func (g *gitHubCrawler) getGhUserID(user *github.User) int {
	if user == nil {
		glog.Error("'user' arg given is nil")
		return -1
	}

	var id int
	err := g.db.QueryRow("SELECT id FROM gh_users WHERE github_id=$1", user.ID).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return 0
	case err != nil:
		glog.Error(err)
		return -1
	}
	return id
}

// getUserID returns the github user id of user in users table.
// If user not in the table, then 0 is returned. If an error occurs, -1 is returned.
func (g *gitHubCrawler) getUserID(user *github.User) int {
	if user == nil {
		glog.Error("'user' arg given is nil")
		return -1
	}

	var id int
	err := g.db.QueryRow("SELECT user_id FROM gh_users WHERE github_id=$1", user.ID).Scan(&id)
	switch {
	case err == sql.ErrNoRows:
		return 0
	case err != nil:
		glog.Error(err)
		return -1
	}
	return id
}

// insertOrUpdateRepo inserts or updates a repository. It also inserts or
// updates related GitHub repository, users, GitHub users and GitHub
// organization (if any).
func (g *gitHubCrawler) insertOrUpdateRepo(repo *github.Repository) bool {
	if repo == nil {
		glog.Error("'repo' arg given is nil")
		return false
	}
	glog.Infof("insert or update repository: %s", *repo.Name)

	clonePath := strings.ToLower(filepath.Join(*repo.Language, *repo.Owner.Login, *repo.Name))
	repoFields := []string{"name", "primary_language", "clone_url", "clone_path", "vcs"}

	var query string
	if id := g.getRepoID(repo); id > 0 {
		query = genUpdateQuery("repositories", id, repoFields...)
	} else if id == 0 {
		query = genInsQuery("repositories", repoFields...)
	} else {
		return false
	}

	var repoID int64
	err := g.db.QueryRow(query+" RETURNING id", repo.Name, repo.Language, repo.CloneURL, clonePath, "git").Scan(&repoID)
	if err != nil {
		glog.Error(err)
		return false
	}

	if *repo.Owner.Type != "Organization" {
		if !g.insertOrUpdateUser(repo.Owner.Login, repoID, 0) {
			return false
		}
	} else {
		if !g.insertOrUpdateGhOrg(repo.Owner.Login, repoID) {
			return false
		}
	}

	if !g.insertOrUpdateGhRepo(repoID, repo) {
		return false
	}

	return true
}

// insertOrUpdateGhRepo inserts, or updates, a github repository in the
// database.
func (g *gitHubCrawler) insertOrUpdateGhRepo(repoID int64, repo *github.Repository) bool {
	if repo == nil {
		glog.Error("'repo' arg given is nil")
		return false
	}
	glog.Infof("insert or update github repository: %s", *repo.Name)

	var ghOrganizationID *int
	if repo.Organization != nil {
		if repo.Organization.ID == nil {
			glog.Info("organization ID is nil")
		} else {
			ghOrganizationID = repo.Organization.ID
		}
	}

	ghRepoFields := []string{
		"repository_id",
		"full_name",
		"description",
		"homepage",
		"fork",
		"github_id",
		"default_branch",
		"master_branch",
		"html_url",
		"forks_count",
		"open_issues_count",
		"stargazers_count",
		"subscribers_count",
		"watchers_count",
		"size_in_kb",
		"created_at",
		"updated_at",
		"pushed_at",
	}

	var query string
	if id := g.getGhRepoID(repo); id > 0 {
		query = genUpdateQuery("gh_repositories", id, ghRepoFields...)
	} else if id == 0 {
		query = genInsQuery("gh_repositories", ghRepoFields...)
	} else {
		return false
	}

	_, err := g.db.Exec(query,
		repoID,
		repo.FullName,
		repo.Description,
		repo.Homepage,
		repo.Fork,
		repo.ID,
		repo.DefaultBranch,
		repo.MasterBranch,
		repo.HTMLURL,
		repo.ForksCount,
		repo.OpenIssuesCount,
		repo.StargazersCount,
		repo.SubscribersCount,
		repo.WatchersCount,
		repo.Size,
		formatTimestamp(repo.CreatedAt),
		formatTimestamp(repo.UpdatedAt),
		formatTimestamp(repo.PushedAt))

	if err != nil {
		glog.Error(err)
		return false
	}

	if ghOrganizationID != nil {
		if !g.insertOrUpdateGhOrg(repo.Organization.Login, repoID) {
			return false
		}
	}

	return true
}

// insertOrUpdateGhOrg inserts, or updates, a github organization into
// the database.
func (g *gitHubCrawler) insertOrUpdateGhOrg(orgName *string, repoID int64) bool {
	if orgName == nil {
		glog.Error("'orgName' arg given is nil")
		return false
	}
	glog.Infof("insert or update github organization: %s", *orgName)

	tmp := g.call(false, g.fetchOrganization, *orgName)
	var org *github.Organization
	switch tmp.(type) {
	case *github.Organization:
		org = tmp.(*github.Organization)
	default:
		glog.Error("invalid function return type")
		return false
	}

	ghOrgFields := []string{
		"login",
		"github_id",
		"avatar_url",
		"html_url",
		"name",
		"company",
		"blog",
		"location",
		"email",
		"collaborators_count",
		"created_at",
		"updated_at",
	}

	var query string
	if id := g.getGhOrgID(org); id > 0 {
		query = genUpdateQuery("gh_organizations", id, ghOrgFields...)
	} else if id == 0 {
		query = genInsQuery("gh_organizations", ghOrgFields...)
	} else {
		return false
	}

	var orgID int64
	err := g.db.QueryRow(query+" RETURNING id",
		org.Login,
		org.ID,
		org.AvatarURL,
		org.HTMLURL,
		org.Name,
		org.Company,
		org.Blog,
		org.Location,
		org.Email,
		org.Collaborators,
		formatTimestamp(&github.Timestamp{Time: *org.CreatedAt}),
		formatTimestamp(&github.Timestamp{Time: *org.UpdatedAt})).Scan(&orgID)

	if err != nil {
		glog.Error(err)
		return false
	}

	tmp = g.call(false, g.fetchOrganizationMembers, *org.Login)
	var users []github.User
	switch tmp.(type) {
	case []github.User:
		users = tmp.([]github.User)
	default:
		glog.Error("invalid function return type")
	}

	for _, user := range users {
		if !g.insertOrUpdateUser(user.Login, repoID, orgID) {
			return false
		}
	}

	return true
}

// insertOrUpdateUser inserts, or updates, a github user into the database.
func (g *gitHubCrawler) insertOrUpdateUser(username *string, repoID int64, orgID int64) bool {
	if username == nil {
		glog.Error("'username' arg given is nil")
		return false
	}
	glog.Infof("insert or update user: %s", *username)

	if repoID <= 0 {
		glog.Error("trying to insert a user without linked GitHub repository")
		return false
	}

	tmp := g.call(false, g.fetchUser, *username)
	var user *github.User
	switch tmp.(type) {
	case *github.User:
		user = tmp.(*github.User)
	default:
		glog.Error("invalid function return type")
		return false
	}

	userFields := []string{"username", "name", "email"}

	var query string
	if id := g.getUserID(user); id > 0 {
		query = genUpdateQuery("users", id, userFields...)
	} else if id == 0 {
		query = genInsQuery("users", userFields...)
	} else {
		return false
	}

	var userID int64
	err := g.db.QueryRow(query+" RETURNING id", user.Login, user.Name, user.Email).Scan(&userID)
	if err != nil {
		glog.Error(err)
		return false
	}

	if !g.linkUserToRepo(userID, repoID) {
		return false
	}

	if !g.insertOrUpdateGhUser(userID, user, orgID) {
		return false
	}

	return true
}

// insertOrUpdateGhUser inserts, or updates, a github user into the database.
func (g *gitHubCrawler) insertOrUpdateGhUser(userID int64, user *github.User, orgID int64) bool {
	if user == nil {
		glog.Error("'user' arg given is nil")
		return false
	}
	glog.Infof("insert or update github user: %s", *user.Login)

	if userID <= 0 {
		glog.Error("trying to insert a github user but no user ID given")
		return false
	}

	ghUserFields := []string{
		"user_id",
		"github_id",
		"login",
		"bio",
		"blog",
		"company",
		"email",
		"hireable",
		"location",
		"avatar_url",
		"html_url",
		"followers_count",
		"following_count",
		"collaborators_count",
		"created_at",
		"updated_at",
	}

	var query string
	if id := g.getGhUserID(user); id > 0 {
		query = genUpdateQuery("gh_users", id, ghUserFields...)
	} else if id == 0 {
		query = genInsQuery("gh_users", ghUserFields...)
	} else {
		return false
	}

	var ghUserID int64
	err := g.db.QueryRow(query+" RETURNING id",
		userID,
		user.ID,
		user.Login,
		user.Bio,
		user.Blog,
		user.Company,
		user.Email,
		user.Hireable,
		user.Location,
		user.AvatarURL,
		user.HTMLURL,
		user.Followers,
		user.Following,
		user.Collaborators,
		formatTimestamp(user.CreatedAt),
		formatTimestamp(user.UpdatedAt)).Scan(&ghUserID)

	if err != nil {
		glog.Error(err)
		return false
	}

	if orgID != 0 {
		if !g.linkGhUserToGhOrg(ghUserID, orgID) {
			return false
		}
	}

	return true
}

// isUserLinkedToRepo checks whether a user is already linked to the given
// repository.
func (g *gitHubCrawler) isUserLinkedToRepo(userID, repoID int64) bool {
	row := g.db.QueryRow(
		`SELECT COUNT(*) AS total
		 FROM users_repositories
		 WHERE user_id = $1 AND repository_id = $2`, userID, repoID)

	var total int64
	if err := row.Scan(&total); err != nil {
		glog.Error(err)
		return false
	}

	return total > 0
}

// linkUserToRepo creates a many to many relationship between a user and a
// repository.
func (g *gitHubCrawler) linkUserToRepo(userID, repoID int64) bool {
	if g.isUserLinkedToRepo(userID, repoID) {
		return true
	}

	fields := []string{"user_id", "repository_id"}

	query := genInsQuery("users_repositories", fields...)

	_, err := g.db.Exec(query, userID, repoID)
	if err != nil {
		glog.Error(err)
		return false
	}

	return true
}

// isGhUserLinkedToGhOrg checks whether a github user is linked to the given
// github organization or not.
func (g *gitHubCrawler) isGhUserLinkedToGhOrg(ghUserID, orgID int64) bool {
	row := g.db.QueryRow(
		`SELECT COUNT(*) AS total
		 FROM gh_users_organizations
		 WHERE gh_user_id = $1 AND gh_organization_id = $2`, ghUserID, orgID)

	var total int64
	if err := row.Scan(&total); err != nil {
		glog.Error(err)
		return false
	}

	return total > 0
}

// linkGhUserToGhOrg links a github user to the given github organization.
func (g *gitHubCrawler) linkGhUserToGhOrg(ghUserID, orgID int64) bool {
	if g.isGhUserLinkedToGhOrg(ghUserID, orgID) {
		return true
	}

	fields := []string{"gh_user_id", "gh_organization_id"}

	query := genInsQuery("gh_users_organizations", fields...)

	_, err := g.db.Exec(query, ghUserID, orgID)
	if err != nil {
		glog.Error(err)
		return false
	}

	return true
}

// fetchOrganization fetches information about a github organization.
// args expects 1 value:
// - orgName: the organization name
func (g *gitHubCrawler) fetchOrganization(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		glog.Error("invalid number of arguments")
		return nil, errInvalidArgs
	}

	var orgName string
	switch args[0].(type) {
	case string:
		orgName = args[0].(string)
	default:
		glog.Errorf("invalid parameter type (given %v, expected string)", reflect.TypeOf(args[0]))
		return nil, errInvalidParamType
	}

	org, resp, err := g.client.Organizations.Get(orgName)
	if err != nil {
		glog.Error(err)
		return nil, g.genAPICallFuncError(resp, err)
	}

	return org, nil
}

// fetchUser fetches information about a user.
// args expects 1 value:
// - username: the user login name
func (g *gitHubCrawler) fetchUser(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		glog.Error("invalid number of arguments")
		return nil, errInvalidArgs
	}

	var username string
	switch args[0].(type) {
	case string:
		username = args[0].(string)
	default:
		glog.Errorf("invalid parameter type (given %v, expected string)", reflect.TypeOf(args[0]))
		return nil, errInvalidParamType
	}

	user, resp, err := g.client.Users.Get(username)
	if err != nil {
		glog.Error(err)
		return nil, g.genAPICallFuncError(resp, err)
	}

	return user, nil
}

// fetchContributors fetches all the contributors of a GitHub repository.
//
// args expects 2 values:
// - owner: the repository owner
// - repoName:  the repository name
//
// It returns a list of users.
func (g *gitHubCrawler) fetchContributors(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		glog.Error("invalid number of arguments")
		return nil, errInvalidArgs
	}

	var owner string
	switch args[0].(type) {
	case string:
		owner = args[0].(string)
	default:
		glog.Errorf("invalid parameter type (given %v, expected string)", reflect.TypeOf(args[0]))
		return nil, errInvalidParamType
	}

	var repoName string
	switch args[1].(type) {
	case string:
		repoName = args[1].(string)
	default:
		glog.Errorf("invalid parameter type (given %v, expected string)", reflect.TypeOf(args[1]))
		return nil, errInvalidParamType
	}

	users, resp, err := g.client.Repositories.ListContributors(owner, repoName, nil)
	if err != nil {
		glog.Error(err)
		return nil, g.genAPICallFuncError(resp, err)
	}

	return users, nil
}

// fetchOrganizationMembers fetches all the members of a GitHub organization.
//
// args expects 1 values:
// - orgName: the organization name
//
// It returns a list of users.
func (g *gitHubCrawler) fetchOrganizationMembers(args ...interface{}) (interface{}, error) {
	if len(args) != 1 {
		glog.Error("invalid number of arguments")
		return nil, errInvalidArgs
	}

	var orgName string
	switch args[0].(type) {
	case string:
		orgName = args[0].(string)
	default:
		glog.Errorf("invalid parameter type (given %v, expected string)", reflect.TypeOf(args[0]))
		return nil, errInvalidParamType
	}

	users, resp, err := g.client.Organizations.ListMembers(orgName, nil)
	if err != nil {
		glog.Error(err)
		return nil, g.genAPICallFuncError(resp, err)
	}

	return users, nil
}

// genAPICallFuncError creates an error base on the http response.
func (g *gitHubCrawler) genAPICallFuncError(resp *github.Response, err error) error {
	if resp == nil {
		glog.Error("'resp' arg given is nil")
		if err != nil {
			return err
		}
		return errNilArg
	}

	if err == nil || resp.StatusCode != 403 {
		return err
	}

	switch {
	case strings.Contains(err.Error(), "API rate limit exceeded"):
		return errTooManyCall
	case strings.Contains(err.Error(), "access blocked"):
		return errUnavailable
	}

	return err
}
