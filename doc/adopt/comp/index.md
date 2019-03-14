# Code search product comparisons

## [Oracle OpenGrok](https://github.com/oracle/opengrok)

OpenGrok was traditionally the most popular open-source code search tool. Originally developed inside Sun Microsystems around 2004, it’s now an Oracle open-source project after Oracle acquired Sun.

**What people tell us they love about OpenGrok:**

* Simple interface
* Support for non-Git repositories
* Easy deployment (for Java shops)

**What has made organizations switch from OpenGrok:**

* Slow and difficult-to-manage indexing process (leading to stale results for users)
* Poor support for searching/browsing multiple commits and branches
* Poor scalability to many repositories and for large repositories
* Inflexible and buggy API (using the new REST API)

## [Hound](https://github.com/etsy/hound)

Hound was created inside Etsy by [Kelly Norton](https://github.com/kellegous) and others and open-sourced in 2015. It appears to not be actively maintained anymore, with the most recent commit (as of this document's publishing) being more than 5 months old, and many open PRs unattended to.

**What people tell us they love about Hound:**

* Very simple deployment
* Very simple interface

**What has made organizations switch from Hound:**

* Apparent unmaintained status
* Poor scalability to many repositories (both performance-wise and UX-wise)
* Limited filtering available for search queries
* No support for code navigation (code intelligence and/or ctags)

## [Livegrep](https://github.com/livegrep/livegrep)

Livegrep is maintained by a Stripe developer and is used inside Stripe. It has a slick demo on the Linux source code at livegrep.com.

**What people tell us they love about Livegrep:**

* Instant, as-you-type search results
* Quick search filters to narrow by path, etc.

**What has made organizations plan to switch from Livegrep:**

* No support for code navigation (code intelligence and/or ctags)
* Difficult to manage and scale (requires building custom infrastructure)
* Although it’s used heavily by a small number of companies, it’s maintained part-time by a single * developer, so future trajectory is limited

## [Atlassian FishEye](https://www.atlassian.com/software/fisheye)

Atlassian released FishEye around 2007, initially as a source browser for companies who (before the advent of GitHub/Bitbucket) mostly lacked web-based source code browsers. You can try it out on [JBoss’ public FishEye instance](https://source.jboss.org/browse) (see [example quick search results page](https://source.jboss.org/qsearch?q=open&t=3&s=2&bucket=ANY_DATE&userFilter=) and [example advanced search results page](https://source.jboss.org/search/Aesh/?head=true&comment=&contents=open&addedText=&deletedText=&filename=&branch=&tag=&fromdate=&todate=&datesortorder=DESCENDING&groupby=file&col=path&col=revision&col=author&col=date&col=csid&refresh=y)).

**What people tell us they love about FishEye:**

* It integrates well with the Atlassian suite of products
* It integrates well with builds in particular (when using Atlassian Bamboo)
* It supports Perforce (as well as Git, Subversion, Mercurial, and CVS)
* Stability and reliability

**What has made organizations switch from FishEye:**

* “It feels like FishEye’s code search is for managers to report on changes, not for developers in their daily workflow”
* Poor integration with GitHub and GitLab
* It is deprioritized by Atlassian, with [only infrequent and minor feature releases](https://confluence.atlassian.com/fisheye/fisheye-releases-960155725.html)

## [Sourcegraph](https://sourcegraph.com/)

(Disclaimer: This document was written by Sourcegraph teammates.)

Sourcegraph was released in Dec 2017 to be the most productive code search and navigation tool for developers. Almost 1,000 companies are using Sourcegraph, with some of the best-known names being Uber, Lyft, Yelp, and several other companies (who we can’t name yet) of similar size and stature. Inside those companies, 40-85% of all developers use Sourcegraph daily.

**What people tell us they love about Sourcegraph:**

* Fast code search with quick reindexing (average time from commit-pushed to reindex at a customer with 30,000 repositories and 3,500 engineers is 30 seconds)
* Clean UI for search that developers can figure out how to use more easily
* Code navigation (go-to-definition and find-references) via precise code intelligence or ctags, depending on the language
* Ease of maintenance and scaling (for admins)

**What has made organizations switch from Sourcegraph:**
No organization with at least 20 daily Sourcegraph users has ever stopped using Sourcegraph, so there are no such companies to speak of.

## Other tools

* **[Google Cloud Source Repositories Code Search](https://cloud.google.com/source-repositories/docs/searching-code)**: This is a hosted Git code search offering that integrates with Google Cloud Source Repositories (Google Cloud’s Git hosting product), but has limited distribution. You can try a public demo at [source.bazel.build](https://source.bazel.build/). 

* **[GitHub code search](https://help.github.com/en/articles/searching-code)**, **[GitLab code search](https://docs.gitlab.com/ee/user/search/advanced_global_search.html)**, **[Bitbucket Server code search](https://confluence.atlassian.com/bitbucketserver/search-for-code-in-bitbucket-server-814204781.html)**, **[Bitbucket Cloud code search](https://confluence.atlassian.com/bitbucket/search-873876782.html)**: This is a great solution for smaller teams who already use these code hosts and who have simple/infrequent code search needs. The main pain points we hear are that they lack support for punctuation or regexps in queries, and the UXs are not optimized for searching across multiple repositories.

* **[Searchcode Server](https://searchcodeserver.com/)**
