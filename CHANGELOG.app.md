<!--
###################################### READ ME ###########################################
### This changelog should always be read on `main` branch. Its contents on version     ###
### branches do not necessarily reflect the changes that have gone into that branch.   ###
### To update the changelog add your changes to the appropriate section under the      ###
### "Unreleased" heading.                                                              ###
##########################################################################################
-->

# Cody App Changelog

All notable changes to the Cody app are documented in this file.

<!-- START CHANGELOG -->

## Unreleased

- Users upgrading from the old Sourcegraph app version no longer need to delete directories / manually migrate.
- Fixed issues around Postgres, improved error messages, fixed issues with Postgres on startup, etc.
- The Cody chat window now shows a helpful error message if you aren’t signed in yet. https://github.com/sourcegraph/sourcegraph/pull/52344
- Settings menu items now go to the new settings pages.
- Improved setup experience with a new ‘You’re all set’ page at the end of the flow.
- An ‘Oops, something went wrong’ page is now displayed if the app has issues https://github.com/sourcegraph/sourcegraph/pull/52453
- Improved messaging if you haven’t verified your email on Sourcegraph.com yet
- Fixed various bugs in setup wizard / VS Code connection process.
- If Docker / `src` are not installed, and executors cannot run, the app will now start with precise code intel / batch changes disabled.

## v2023.5.23

- This is a new (super experimental, early release) version of the Souregraph app which lets you use Cody on your local Git repositories!
- Adding your Git repositories is now much easier
- We made the app a much more native experience (by adopting [Tauri](https://tauri.app)), and you can now access Cody through the system tray.
- You can now ask Cody questions side-by-side with your editor open
- If you use VS Code, the Cody extension now offers a `Connect Sourcegraph App` option (you may need to sign out using the settings gear to see it). Cody in VS Code will then talk to your Sourcegraph app to answer questions. In the future, this will enable much more powerful answers including context about all of your local code - not just what's open in your editor.
