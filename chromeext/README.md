# Sourcegraph chrome extension

Enhance code browsing on GitHub: code search, instant documentaton tooltips, and jump-to-definition links for code on GitHub.

This extension enhances file pages on GitHub by adding instant documentation and type tooltips to code on GitHub, and making every identifer a jump-to-definition link. It also adds a search button and keyboard shortcut (shift-T) that allow you to search for functions, types, and other code definitions within your repository. The search function also provides complete text results. You'll wonder how you browsed code on GitHub without it.

Here's what people are saying about Sourcegraph:

- "it's already my favorite online doc tool. give it a try" - @edapx
- "Sourcegraph has been quickly moving in priority on my Chrome bookmarks. You need to check out this tool." - @goinggodotnet
- "Loving Sourcegraph.com to find documentation and real code examples" - @IndianGuru
- "Being able to search actual open source code is amazing. Very very fast as well!" and Sourcegraph is blowing my mind right now." - @joshtaylor
- "Sourcegraph is amazing" - Jakub W.
- "Sourcegraph is an essential tool for every Pythonist, Gopher, and Nodester." - @nimolix
- "Impressive #golang code navigation by Sourcegraph!" - @francesc

You can also search and browse code on Sourcegraph itself at https://sourcegraph.com.

Currently, Sourcegraph supports Go repositories. This extension will work on all public Go code. To use code search on private Go repositories, sign up for Sourcegraph at https://sourcegraph.com and link your GitHub account. Support for more languages will be rolled out soon. Stay tuned!   

Prerequisite: Using the extension requires the repository to be built on Sourcegraph. All popular Go repositories are already built. To trigger a build of a repository, visit https://sourcegraph.com/github.com/USER/REPO.    

## Development

## Prerequisites

## Building

Go to `chrome://extensions`in Chrome, check `Developer mode` and use `Load Unpacked Extension` to load the
`chromeext` directory.

To reload the Chrome extension when you change files, install
[Extensions Reloader](https://chrome.google.com/webstore/detail/fimgfedafeadlieiabdeeaodndnlbhid), 
or simply `Update extensions` and `Reload` the extension at chrome://extensions.

## Publishing

To publish a new version:

1. Visit the [Chrome Web Store](https://chrome.google.com/webstore).
1. Click the Settings gear in top right corner -> `Developer Dashboard`.
1. Choose the right extension with the `Status` of `Published` and click `Edit`.
1. Bump the version in `manifest.json` to be greater than the current one.
1. Test that it works by loading it into Chrome via `Load Unpacked Extension`.
1. `cd chromeext` and `zip -r chromeext.zip *` to create a zip of the `chromeext` directory.
1. Click `Upload Updated Package` on the developer dashboard to upload the new zip file.
1. Scroll to the bottom of the page and hit `Publish Changes`.



