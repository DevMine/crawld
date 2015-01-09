# crawld: a data crawler and repository fetcher

[![Build Status](https://travis-ci.org/DevMine/crawld.png?branch=master)](https://travis-ci.org/DevMine/crawld)
[![GoDoc](http://godoc.org/github.com/DevMine/crawld?status.svg)](http://godoc.org/github.com/DevMine/crawld)
[![GoWalker](http://img.shields.io/badge/doc-gowalker-blue.svg?style=flat)](https://gowalker.org/github.com/DevMine/crawld)
[![Gobuild Download](http://gobuild.io/badge/github.com/DevMine/crawld/downloads.svg)](http://gobuild.io/github.com/DevMine/crawld)

`crawld` is a data crawler and source code repository fetcher.

Currently, only a [GitHub](https://github.com) crawler is implemented but the
architecture of `crawld` has been designed in a way such that new crawlers (for
instance a [BitBucket](https://bitbucket.org/) crawler) can be added without
hassle, regardless of the source code management system
([git](http://git-scm.com/), [mercurial](http://mercurial.selenic.com/),
[svn](http://subversion.apache.org/), ...).

This crawler focuses on crawling repositories metadata and those of the users
that contributed, or are directly related, to the repositories.

All of the collected metadata is stored into a
[PostgreSQL](http://www.postgresql.org/) database. As `crawld` is designed to
support several platforms, information common across them is stored in two
tables: `users` and `repositories`. For the rest of the information, specific
tables are created (`gh_repositores`, `gh_users` and `gh_organizations` for
now) and relations are established with the `users` and `repositories` tables.

The table below gives information about what is collected. Bear in mind that
some information might be incomplete (for instance, if a user does not provide
any company information).

Repository       | GitHub Repository | User     | GitHub User         | GitHub Organization
-----------------|-------------------|----------|---------------------|--------------------
Name             | GitHub ID         | Username | GitHub ID           | GitHub ID
Primary language | Full name         | Name     | Login               | Login
Clone URL        | Description       | Email    | Bio                 | Avatar URL
                 | Homepage          |          | Blog                | HTML URL
                 | Fork              |          | Company             | Name
                 | Default branch    |          | Email               | Company
                 | Master branch     |          | Hireable            | Blog
                 | HTML URL          |          | Location            | Location
                 | Forks count       |          | Avatar URL          | Email
                 | Open issues count |          | HTML URL            | Collaborators count
                 | Stargazers count  |          | Followers count     | Creation date
                 | Subscribers count |          | Following count     | Update date
                 | Watchers count    |          | Collaborators count |
                 | Size              |          | Creation date       |
                 | Creation date     |          | Update date         |
                 | Update date       |          |                     |
                 | Last push date    |          |                     |

Besides crawling, `crawld` is also able to clone and update repositories, from
their cloning URL stored into the database. Depending on the number of
repositories you have in your database, this may required a fair amount of
storage space.

## Installation

To install `crawld`, run this command in a terminal, assuming
[Go](http://golang.org/) is installed:

    go get github.com/DevMine/crawld

Or you can download a binary for your platform from the DevMine project's
[downloads page](http://devmine.ch/downloads).

You also need to setup a [PostgreSQL](http://www.postgresql.org/) database. Look
at the [README file](https://github.com/DevMine/crawld/blob/master/db/README.md)
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
     will get at most the 1000 most popular languages (in terms of
     stars count) per language listed in "languages".

Once the configuration file has been adjusted, you are ready to run `crawld`.
You need to specify the path to the configuration file with the help of the `-c`
option. Example:

    crawld -c crawld.conf

Some command line options are also available, mainly where to store log files
and whether to disable the data crawlers or repositories fetcher (by default,
the crawlers and the fetcher run in parallel). See `crawld -h` for more
information.
