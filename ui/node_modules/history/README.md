# history [![Travis][build-badge]][build] [![npm package][npm-badge]][npm]

[build-badge]: https://img.shields.io/travis/ReactTraining/history/master.svg?style=flat-square
[build]: https://travis-ci.org/ReactTraining/history

[npm-badge]: https://img.shields.io/npm/v/history.svg?style=flat-square
[npm]: https://www.npmjs.org/package/history

[`history`](https://www.npmjs.com/package/history) is a JavaScript library that lets you easily manage session history anywhere JavaScript runs. `history` abstracts away the differences in various environments and provides a minimal API that lets you manage the history stack, navigate, confirm navigation, and persist state between sessions.

## Docs & Help

- [Guides and API Docs](/docs#readme)
- [Changelog](/CHANGES.md)

## Installation

Using [npm](https://www.npmjs.com/):

    $ npm install --save history

Then with a module bundler like [webpack](https://webpack.github.io/), use as you would anything else:

```js
// using an ES6 transpiler, like babel
import { createHistory } from 'history'

// not using an ES6 transpiler
var createHistory = require('history').createHistory
```

The UMD build is also available on [unpkg](https://unpkg.com):

```html
<script src="https://unpkg.com/history/umd/history.min.js"></script>
```

You can find the library on `window.History`.

## Basic Usage

A "history" encapsulates navigation between different screens in your app, and notifies listeners when the current screen changes.

```js
import { createHistory } from 'history'

const history = createHistory()

// Get the current location
const location = history.getCurrentLocation()

// Listen for changes to the current location
const unlisten = history.listen(location => {
  console.log(location.pathname)
})

// Push a new entry onto the history stack
history.push({
  pathname: '/the/path',
  search: '?a=query',

  // Extra location-specific state may be kept in session
  // storage instead of in the URL query string!
  state: { the: 'state' }
})

// When you're finished, stop the listener
unlisten()
```

You can find many more examples [in the documentation](https://github.com/ReactTraining/history/tree/master/docs)!

## Thanks

A big thank-you to [Dan Shaw](https://www.npmjs.com/~dshaw) for letting us use the `history` npm package name! Thanks Dan!

Also, thanks to [BrowserStack](https://www.browserstack.com/) for providing the infrastructure that allows us to run our build in real browsers.
