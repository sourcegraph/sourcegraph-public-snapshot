# Glossary

If you've encountered a term and aren't sure what it is, start here. This is intended to describe what a thing is, where it comes from, and how it relates to our product or processes.

### [Atom](https://atom.io/)

GitHub's extensible/customizable text editor, which can be adapted to do things like [LSP](#lsp) support.

### [Buildkite](https://buildkite.com/)

Buildkite is the [continuous integration](#ci) tool we use. Our instances are actually connected to [kubernetes](#kubernetes) stuff running on Google Cloud. Buildkite runs builds automatically whenever changes show up in our trees (but only for internal pull requests, not for pull requests from third parties), and produces feedback on whether the pull breaks anything. To access our Buildkite infrastructure, [make a Buildkite account](https://buildkite.com/signup), then ask for access to the [Sourcegraph organization](https://buildkite.com/sourcegraph). Don't create a new organization when the signup process prompts; just give your email address to the person administering it. (Nick Snyder.)

### CI

Continuous Integration; run tests constantly rather than waiting for users to manually initiate them.

### [docker](https://www.docker.com/)

Container/image infrastructure to let you specify sets of things to install in virtualized machines and run, in theory improving reliability with which you can run code on arbitrary machines.

### FTT

Full-Time Teammate. Internal abbreviation used to refer to full-time staff here.

### [GitHub](https://github.com/)

Large public source repository. SourceGraph has a number of useful integrations with GitHub, also we use it for issue tracking and pull requests for many of our projects. You will need to be added to the `sourcegraph/ftts` team to get access to the internal repositories. Note that you can change notification settings, and probably should; the defaults are incredibly spammy. (Unwatch all repositories, watch issues only when pinged or participating.)

Also used to refer to the Sourcegraph integration for GitHub, available as browser extensions for Chrome and Firefox.

### [GraphQL](http://graphql.org/)

Query language for APIs designed to allow ongoing revisions of APIs without breaking existing code. Used internally as interaction between the Sourcegraph server and other components, such as our UI.

### [Gusto](https://app.gusto.com/login)

Gusto is the third-party service which handles things like paychecks and benefits for most employees.

### [Iteration](https://docs.google.com/document/d/1W7Stwye0zYX1jjMCCUdmjxDjv1OIBouBdcNyF3DAAWs/edit#)

Sourcegraph does development in roughly-four-week "iterations"; one week of planning, two weeks development, one week testing and bug fixes. (Link is to a Google Docs document describing the process. Ask for access to the doc if you don't have it already.)

### [kubernetes](https://kubernetes.io/)

Container infrastructure, used to let us build images to run that have predictable qualities and can be started and stopped on cloud infrastructure conveniently.

### [Lattice](https://sourcegraph.latticehq.com/)

Lattice is the service used to do review and goal tracking. These are done quarterly.

### [LSP](https://langserver.org/)

Language Server Protocol. A Microsoft spec intended to provide an easier interface for interactions between tools which need code intelligence (such as editors, IDEs, and code search tools) and tools which can provide it, such as language analyzers.

### [PostgreSQL](https://www.postgresql.org/)

Relational database (SQL-based). Used in Sourcegraph for persistent data that needs to survive new installations, etcetera.

### [redis](https://redis.io/)

In-memory data store/cache, with fancy data structures and queries. Used in Sourcegraph for transient data storage.

### [sourcegraph/infrastructure](https://github.com/sourcegraph/infrastructure)

The infrastructure repository is what holds configuration, setup, and tools used for [Sourcegraph Data Center](#sourcegraph-data-center), the public sourcegraph.com site, and other internal developer tools and resources. This repository does not have an issue tracker; file issues against the [main sourcegraph repository](#sourcegraph-sourcegraph).

### [sourcegraph/issues](https://github.com/sourcegraph/issues)

Public issue tracker. This is where customers and interested third parties would file issues.

### [sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph)

Our main repository, also our primary internal issue tracker.

### [VSCode](https://code.visualstudio.com/)

Microsoft's "Visual Studio Code", a fairly powerful and flexible programming editor, which has [LSP](#lsp) support.

### [Zoom](https://zoom.us/)

Video conferencing software. You will need/want to download and install their application to allow joining conferences. Supports computer audio, or you can dial in with a phone.
