+++
title = "Getting started with Sourcegraph for your team"
navtitle = "Getting started"
+++

Setting up Sourcegraph for your team usually takes less than 7
minutes. Let's get started.

# 1. Install Sourcegraph on a server

Choose the **Server installation** link on the left for your preferred
platform. Follow those instructions, and come back here when you're
done.

For simplicity, the rest of this document assumes your new Sourcegraph
server is at `https://src.example.com`. Replace that with the actual
scheme (HTTP/HTTPS) and hostname of your own server.

# 2. Register your Sourcegraph account and server

If you didn't already do this in step 1, visit your new Sourcegraph
server and follow the prompts to register an account and your new
server with Sourcegraph.com.

After you're finished, you'll be back at the homepage of your
Sourcegraph server, and your username will be shown in the top right.

If there are any more installation steps necessary, the homepage will
tell you what to do.

# 3. Try it out on public code

Your Sourcegraph server can transparently load public code via
Sourcegraph.com. This means your team gets the power of Sourcegraph on
all code, including open-source dependencies---not just your team's
code (which stays private on your server).

Just put the name of any open-source project in the URL path after
your Sourcegraph server's hostname. For example, if your Sourcegraph
server is at `https://src.example.com`, visit these projects:

* <code>https://<i></i>src.example.com/<strong>github.com/gorilla/mux</strong></code>
* <code>https://<i></i>src.example.com/<strong>github.com/JodaOrg/joda-time</strong></code>

# 4. Add a repository

Now, let's add a repository to your Sourcegraph server. We'll assume
you already have a repository that you'd like to use on Sourcegraph,
but these steps also work for brand new repositories. (Note: Currently,
Sourcegraph only supports Go and Java.)

1. [Install the **`src`** CLI tool](https://sourcegraph.com/b86d5501a450ca38be78b112d88cb46d9bf27583/try-it)
   on your own machine. (Future steps assume it is in your `$PATH`,
   but that is not necessary.)
1. Log into your Sourcegraph server from the CLI: `src --endpoint https://src.example.com login`
1. Create a new empty repository: `src repo create myfirstrepo`
1. From your local clone of the repository, run:
  1. `git remote add sourcegraph https://src.example.com/myfirstrepo`
     * When you're ready to switch to Sourcegraph, you can rename this remote to `origin`.
  1. `git push sourcegraph master`
1. Visit <code>https://<i></i>src.example.com/myfirstrepo</code> in your browser.

On the repository page, you can see the progress of Sourcegraph's code
analysis. If there is an issue, it will be red. Need to
[troubleshoot a failed or incomplete build?]({{< relref
"troubleshooting/builds.md" >}})

Now, you can start using some of the killer features of Sourcegraph:

* **Search** your code by typing in anything in the top search bar.
* **Browse** your code and **find usages** of any function, type,
  etc., by opening any code file and clicking on any token.

# 5. Get the rest of your team on Sourcegraph

As the server admin, you need to [grant access to your teammates]({{<
relref "config/access-control.md" >}}) so they can access your
Sourcegraph server.
