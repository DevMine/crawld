# crawld: a data crawler and repositories fetcher

[![Build Status](https://travis-ci.org/DevMine/crawld.png?branch=master)](https://travis-ci.org/DevMine/crawld)
[![GoDoc](http://godoc.org/github.com/DevMine/crawld?status.svg)](http://godoc.org/github.com/DevMine/crawld)
[![GoWalker](http://img.shields.io/badge/doc-gowalker-blue.svg?style=flat)](https://gowalker.org/github.com/DevMine/crawld)

`crawld` is a metadata crawler and source code repository fetcher.
Hence, `crawld` comprises two different parts: the crawlers and the fetcher.

## Crawlers

`crawld` focuses on crawling repositories metadata and those of the users
that contributed, or are directly related, to the repositories from code
sharing platforms such as [GitHub](https://github.com).

Only a GitHub crawler is currently implemented.
However, the architecture of `crawld` has been designed in a way such that new
crawlers (for instance a [BitBucket](https://bitbucket.org/) crawler) can be
added without hassle.

All of the collected metadata is stored into a
[PostgreSQL](http://www.postgresql.org/) database. As `crawld` is designed to
be able to crawl several code sharing platforms, common information is stored
in two tables: `users` and `repositories`. For the rest of the information,
specific tables are created (`gh_repositories`, `gh_users` and
`gh_organizations` for now) and relations are established with the `users` and
`repositories` tables.

The table below gives information about what is collected. Bear in mind that
some information might be incomplete (for instance, if a user does not provide
any company information).

<table>
  <thead>
    <tr>
      <th>Repository</th>
      <th>GitHub Repository</th>
      <th>User</th>
      <th>GitHub User</th>
      <th>GitHub Organization</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>Name</td>
      <td>GitHub ID</td>
      <td>Username</td>
      <td>GitHub ID</td>
      <td>GitHub ID</td>
    </tr>
    <tr>
      <td>Primary language</td>
      <td>Full name</td>
      <td>Name</td>
      <td>Login</td>
      <td>Login</td>
    </tr>
    <tr>
      <td>Clone URL</td>
      <td>Description</td>
      <td>Email</td>
      <td>Bio</td>
      <td>Avatar URL</td>
    </tr>
    <tr>
      <td> </td>
      <td>Homepage</td>
      <td> </td>
      <td>Blog</td>
      <td>HTML URL</td>
    </tr>
    <tr>
      <td> </td>
      <td>Fork</td>
      <td> </td>
      <td>Company</td>
      <td>Name</td>
    </tr>
    <tr>
      <td> </td>
      <td>Default branch</td>
      <td> </td>
      <td>Email</td>
      <td>Company</td>
    </tr>
    <tr>
      <td> </td>
      <td>Master branch</td>
      <td> </td>
      <td>Hireable</td>
      <td>Blog</td>
    </tr>
    <tr>
      <td> </td>
      <td>HTML URL</td>
      <td> </td>
      <td>Location</td>
      <td>Location</td>
    </tr>
    <tr>
      <td> </td>
      <td>Forks count</td>
      <td> </td>
      <td>Avatar URL</td>
      <td>Email</td>
    </tr>
    <tr>
      <td> </td>
      <td>Open issues count</td>
      <td> </td>
      <td>HTML URL</td>
      <td>Collaborators count</td>
    </tr>
    <tr>
      <td> </td>
      <td>Stargazers count</td>
      <td> </td>
      <td>Followers count</td>
      <td>Creation date</td>
    </tr>
    <tr>
      <td> </td>
      <td>Subscribers count</td>
      <td> </td>
      <td>Following count</td>
      <td>Update date</td>
    </tr>
    <tr>
      <td> </td>
      <td>Watchers count</td>
      <td> </td>
      <td>Collaborators count</td>
      <td> </td>
    </tr>
    <tr>
      <td> </td>
      <td>Size</td>
      <td> </td>
      <td>Creation date</td>
      <td> </td>
    </tr>
    <tr>
      <td> </td>
      <td>Creation date</td>
      <td> </td>
      <td>Update date</td>
      <td> </td>
    </tr>
    <tr>
      <td> </td>
      <td>Update date</td>
      <td> </td>
      <td> </td>
      <td> </td>
    </tr>
    <tr>
      <td> </td>
      <td>Last push date</td>
      <td> </td>
      <td> </td>
      <td> </td>
    </tr>
  </tbody>
</table>

## Fetcher

Aside from crawling metadata, `crawld` is able to clone and update
repositories, using their clone URL stored into the database.

Cloning and updating can be done regardless of the source code management
system in use ([git](http://git-scm.com/),
[mercurial](http://mercurial.selenic.com/),
[svn](http://subversion.apache.org/), ...), however only a `git` fetcher is
currently implemented.

As source code repositories usually contain a lot of files, `crawld` has an
option that allows storing source code repositories as tar archives which makes
things easier for the file system shall you clone a huge number of
repositories.

## Installation

`crawld` uses `git2go`, a `ligit2` Go binding for its `git` operations. Hence,
`libgit2` needs to be installed on your system unless you statically compile it
with the `git2go` package.

To install `crawld`, run this command in a terminal, assuming
[Go](http://golang.org/) is installed:

    go get github.com/DevMine/crawld

Or you can download a binary for your platform from the DevMine project's
[downloads page](http://devmine.ch/downloads).

You also need to setup a [PostgreSQL](http://www.postgresql.org/) database.
Look at the
[README file](https://github.com/DevMine/crawld/blob/master/db/README.md)
in the `db` sub-folder for details.

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
 * **fetch\_languages**: specify the list of languages the fetcher shall
   restrict to. If left empty, all languages are considered.
 * **tar\_repositories**: a boolean value indicating whether the repositories
   shall be stored as tar archives or not.
 * **tmp\_dir**: specify a temporary working directory. If left empty, the
   default temporary directory will be used. This directory is used on clone and
   update operations when the _tar\_repositories_ option is activated. It is
   recommended to use a ramdisk for increased performance.
 * **tmp\_dir\_file\_size\_limit**: specify the maximum size in GB of an object
   to be temporarily placed in _tmp\_dir_ for processing. Files of size larger
   than this value will not be processed in _tmp\_dir_.
 * **max\_fetcher\_workers**: specify the maximum number of workers to use for
   the fetching task. It defaults to 1 but if your machine has good I/O
   throughput and a good CPU, you probably want to increase this conservative
   value for performance reasons. Note that fetching is I/O and networked bound
   more than CPU bound and hence you probably do not want to increase this
   value too much.
 * **throttler\_wait\_time**: indicates how much time to wait, in seconds,
   before resuming normal operation after throttling.
 * **throttler\_sliding\_window\_size**: represents the size of the sliding
   window used by the error rate throttler. If you have no idea about what that
   means, it is safe to omit it since default value shall be sane.
 * **throttler\_leak\_interval**: specify the time to wait, in milliseconds,
   before taking off a unit from the sliding window. Again, if you have no idea
   about what that means, it is safe to omit it since default value shall be
   sane.
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
   - **since\_id**: corresponds to the repository ID (eg: GitHub repository ID
     in the case of the github crawler) from which to start querying
     repositories. Note that this value is ignored when using the search API.
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
     will get at most the 1000 most popular projects (in terms of
     stars count) per language listed in "languages".

Once the configuration file has been adjusted, you are ready to run `crawld`.
You need to specify the path to the configuration file with the help of the `-c`
option. Example:

    crawld -c crawld.conf

Some command line options are also available, mainly where to store log files
and whether to disable the data crawlers or repositories fetcher (by default,
the crawlers and the fetcher run in parallel). See `crawld -h` for more
information.
