# scroll-behavior [![Travis][build-badge]][build] [![npm][npm-badge]][npm]

Scroll management for [`history`](https://github.com/ReactTraining/history).

**If you are using [React Router](https://github.com/reactjs/react-router), check out [react-router-scroll](https://github.com/taion/react-router-scroll), which wraps up the scroll management logic here into a router middleware.**

[![Codecov][codecov-badge]][codecov]
[![Discord][discord-badge]][discord]

## Usage

```js
import createHistory from 'history/lib/createBrowserHistory';
import withScroll from 'scroll-behavior';

const history = withScroll(createHistory());
```

## Guide

### Installation

```
$ npm i -S history
$ npm i -S scroll-behavior
```

### Scroll behaviors

### Basic usage

Extend your history object using `withScroll`. The extended history object will manage the scroll position for transitions.

### Custom scroll behavior

You can customize the scroll behavior by providing a `shouldUpdateScroll` callback when extending the history object. This callback is called with both the previous location and the current location.

You can return:

- a falsy value to suppress the scroll update
- a position array such as `[0, 100]` to scroll to that position
- a truthy value to get normal scroll behavior

```js
const history = withScroll(createHistory(), (prevLocation, location) => (
  // Don't scroll if the pathname is the same.
  !prevLocation || location.pathname !== prevLocation.pathname
));
```

```js
const history = withScroll(createHistory(), (prevLocation, location) => (
  // Scroll to top when attempting to vist the current path.
  prevLocation && location.pathname === prevLocation.pathname ? [0, 0] : true
));
```

### Scrolling elements other than `window`

The `withScroll`-extended history object has a `registerScrollElement` method. This method registers an element other than `window` to have managed scroll behavior on transitions. Each of these elements needs to be given a unique key at registration time, and can be given an optional `shouldUpdateScroll` callback that behaves as above.

```js
const history = withScroll(createHistory(), () => false);
history.listen(listener);

history.registerScrollElement(
  key, element, shouldUpdateScroll
);
```

The `registerScrollElement` method returns an `unregister` function that you can use to explicitly unregister the scroll behavior on the element, if necessary. In general, you will not need to do this, as `withScroll` will perform all necessary cleanup on removal of the last history listener.

[build-badge]: https://img.shields.io/travis/taion/scroll-behavior/master.svg
[build]: https://travis-ci.org/taion/scroll-behavior

[npm-badge]: https://img.shields.io/npm/v/scroll-behavior.svg
[npm]: https://www.npmjs.org/package/scroll-behavior

[codecov-badge]: https://img.shields.io/codecov/c/github/taion/scroll-behavior/master.svg
[codecov]: https://codecov.io/gh/taion/scroll-behavior

[discord-badge]: https://img.shields.io/badge/Discord-join%20chat%20%E2%86%92-738bd7.svg
[discord]: https://discord.gg/0ZcbPKXt5bYaNQ46
