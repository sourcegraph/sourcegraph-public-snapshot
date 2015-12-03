+++
title = "Jenkins"
+++

You may configure Jenkins to poll Sourcegraph for code changes to trigger your builds.

# Configure Jenkins

Make sure you have installed the Git Plugin(s) and any other necessary tools (e.g. Maven) to
run your build configuration.

# Configure your project

First, create a Jenkins project for your builds. Then set the following configuration:

- **Source Code Management:** Git
	- Enter your repository URL (e.g. http://src.mycompany.com/repo).
	- Enter credentials; you may use your Sourcegraph username/password or [SSH keypair]({{< relref "config/ssh.md" >}}).
	- Select which branches to build.
- **Build Triggers:** Poll SCM
	- Choose a polling interval suitable for your project.
