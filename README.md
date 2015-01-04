# Crawld: a data crawler and repository fetcher

`crawld` is part of the [DevMine project](http://devmine.github.io/).

It is responsible for collecting source code projects metadata and code
from repositories. Currently, only a GitHub crawler is implemented but
the architecture of the crawler has been thought that new crawlers (eg a
BitBucket crawler) can be added without hassle, regardless of the source
code management system (git, mercurial, svn, ...).

This crawler focuses on crawling repositories metadata and those of the
users that contributed to the repositories.

Typical information gathered for a repository are:

 * Name
 * Primary language
 * Clone URL
 * Clone path (where to clone the repository)
 * Version control system (git, mercurial, ...)

And for a user:

 * Username
 * Name
 * Email

Other information specific to a crawler type are also gathered. Bear in
mind the some information might be incomplete (for instance, if a use
does not provide any company information). Here is the information
gathered by the GitHub crawler for a GitHub user:

 * GitHub ID
 * Login
 * Bio
 * Blog
 * Company
 * Email
 * Hireable
 * Location
 * Avatar URL
 * HTML URL
 * Followers count
 * Following count
 * Collaborators count
 * Creation date
 * Update date

And for a GitHub repository:

 * GitHub ID
 * Full name
 * Description
 * Homepage
 * Fork
 * Default branch
 * Master branch
 * HTML URL
 * Forks count
 * Open issues count
 * Stargazers count
 * Subscribers count
 * Watchers count
 * Size in kb
 * Creation date
 * Update date
 * Last push date

Some information is also gathered regarding GitHub organizations the
GitHub users belong to:

 * GitHub ID
 * Login
 * Avatar URL
 * HTML URL
 * Name
 * Company
 * Blog
 * Location
 * Email
 * Collaborators count
 * Creation date
 * Update date

Besides getting these information, `crawld` can also clone and update
the repositories. Depending on the number of repositories you have in
your database, this may required a fair amount of storage space.

## Installation

This will get you `crawld`:

    go get github.com/DevMine/crawld

Or you can download a binary for your platform from
[gobuild.io](http://gobuild.io/github.com/DevMine/crawld).

You also need to setup a [PostgreSQL](http://www.postgresql.org/)
database. Look at the [README
file](https://github.com/DevMine/crawld/blob/master/db/README.md) in the
`db` sub-folder for details.

## Usage and configuration

Copy `crawld.conf.sample` to `crawld.conf` and edit it according to your
needs. The configuration file has several sections:

 * **database**: allows you to configure access to your PostgreSQL
   database.
   - **hostname**: hostname of the machine.
   - **port**: PostgreSQL port.
   - **username**: PostgreSQL user that has access to the database.
   - **password**: password of the database user.
   - **dbname**: database name.
   - **ssl\_mode**: takes any of these 4 values: "disable",
     "require", "verify-ca", "verify-null". Refer to PostgreSQL
     [documentation](http://www.postgresql.org/docs/9.4/static/libpq-ssl.html)
     for details.
 * **clone\_dir**: specify where you would like to clone the
   repositories.
 * **crawling\_time\_interval**: specify the waiting time between 2
   full crawling periods. This is irrelevant for the crawlers where no
   limit is specified.
 * **fetch\_time\_interval**: specify the waiting time between 2 full
   repositories cloning/updating periods. You shall preferably choose a
   small time period here since the repositories fetcher cannot usually
   keep up with the crawlers and you likely want it to update/clone the
   repositories continuously.
 * **crawlers**: allows you to configure options for the crawlers.
   - **type**: specify crawler type. Currently, only "github" is
     implemented.
   - **languages**: list of programming languages of the repositories
     you are interested into. All languages used in a repository are
     considered and not only the primary language.
   - **limit**: set this value to 0 to not use a limit. Otherwise,
     crawling will stop when "limit" repositories have been fetched.
     Note that the behavior is slightly different whether you use the
     search API or not. When you use the search API, this limit
     correspond to the number of repositories to crawl *per language*
     listed in "languages" . When you do not use the search API, this
     is a global limit, regardless of the language.
   - **fork**: skip fork repositories if set to false.
   - **oauth\_access\_token**: your API token. If not provided,
     `crawld` will work but the number of API call is usually limited
     to a low number. For instance, in the case of the GitHub
     crawler, unauthenticated requests are limited to 60 per hour
     where authenticated requests goes up to 5000 per hour.
   - **use\_search\_api**: specify whether you want to use the search
     API or not. Bear in mind that results returned via the search
     API are usually limited so you probably not want this option set
     to true usually. In the case of the GitHub crawler, when set to
     true, the limit is 1000 results per search. This means that you
     will get at most the 1000 most popular languages (in terms of
     stars count) per language listed in "languages".

Once the configuration file has been adjusted, you are ready to run
`crawld`. You need to specify the path to the configuration file with
the help of the `-c` option. Example:

    crawld -c crawld.conf

Some command line options are also available, mainly where to store log
files and whether to disable the data crawlers or repositories fetcher
(by default, the crawlers and the fetcher run in parallel). See
`crawld -h` for more information.
