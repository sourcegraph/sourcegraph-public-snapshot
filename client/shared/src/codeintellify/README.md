# CodeIntellify

![build](https://github.com/sourcegraph/codeintellify/actions/workflows/build.yml/badge.svg)
[![codecov](https://codecov.io/gh/sourcegraph/codeintellify/branch/master/graph/badge.svg?token=1Xk7sdvG0y)](https://codecov.io/gh/sourcegraph/codeintellify)
[![code style: prettier](https://img.shields.io/badge/code_style-prettier-ff69b4.svg)](https://github.com/prettier/prettier)
[![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://github.com/semantic-release/semantic-release)
[![sourcegraph: search](https://img.shields.io/badge/sourcegraph-search-brightgreen.svg)](https://sourcegraph.com/github.com/sourcegraph/codeintellify)

This library manages all of the inputs (mouse/keyboard events, location changes, hover information, and hover actions) necessary to display hover tooltips on with a code view. All together, this makes it easier to add code intelligence to code views on the web. Used in [Sourcegraph](https://sourcegraph.com).

## What it does

- Listens to hover and click events on the code view
- On mouse hovers, determines the line+column position, performs a hover request, and renders it in a nice tooltip overlay at the token
- Shows actions in the hover
- When clicking a token, pins the tooltip to that token
- Highlights the hovered token

You need to provide your own UI component (referred to as the HoverOverlay) that actually displays this information and exposes these actions to the user.

## Usage

- Call `createHoverifier()` to create a `Hoverifier` object (there should only be one on the page, to have only one HoverOverlay shown).
- The Hoverifier exposes an Observable `hoverStateUpdates` that a consumer can subscribe to, which emits all data needed to render the HoverOverlay
- For each code view on the page, call `hoverifier.hoverify()`, passing the position events coming from `findPositionsFromEvents()`.
- `hoverify()` returns a `Subscription` that will "unhoverify" the code view again if unsubscribed from

## Development

```sh
yarn
yarn test

# Helpful options:
yarn test -- --single-run      # Don't rerun on changes
yarn test -- --browsers Chrome # Only run in Chrome
```

Development is done by running tests. [Karma](https://github.com/karma-runner/karma) is used to run
[Mocha](https://github.com/mochajs/mocha) tests in the browser. You can debug by opening http://localhost:9876/debug.html in
a browser while the test running is active. The tests will rerun automatically when files are changed.

You can run specific tests by [adding `.only` to `describe` or `it` calls](https://mochajs.org/#exclusive-tests).

## Releases

Releases are done automatically in CI when commits are merged into master by analyzing [Conventional Commit Messages](https://conventionalcommits.org/).
After running `yarn`, commit messages will be linted automatically when committing.
You may have to rebase a branch before merging to ensure it has a proper commit history.

## Glossary

| Term                | Definition                                                                                                                                                                                                    |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Code view           | The DOM element that contains all the line elements                                                                                                                                                           |
| Line number element | The DOM element that contains the line number label for that line                                                                                                                                             |
| Code element        | The DOM element that contains the code for one line                                                                                                                                                           |
| Diff part           | The part of the diff, either base, head or both (if the line didn't change). Each line belongs to one diff part, and therefor to a different commit ID and potentially different file path.                   |
| Hover overlay       | Also called tooltip                                                                                                                                                                                           |
| hoverify            | To attach all the listeners needed to a code view so that it will display overlay on hovers and clicks.                                                                                                       |
| unhoverify          | To unsubscribe from the Subscription returned by `hoverifier.hoverify()`. Removes all event listeners from the code view again and hides the hover overlay if it was triggered by the unhoverified code view. |
