# scroll-behavior [![Travis][build-badge]][build] [![npm][npm-badge]][npm]

Pluggable browser scroll management.

**If you use [React Router](https://github.com/reactjs/react-router), use [react-router-scroll](https://github.com/taion/react-router-scroll), which wraps up the scroll management logic here into a React Router middleware.**

[![Codecov][codecov-badge]][codecov]
[![Discord][discord-badge]][discord]

## Usage

```js
import ScrollBehavior from 'scroll-behavior';

/* ... */

const scrollBehavior = new ScrollBehavior({
  addTransitionHook,
  stateStorage,
  getCurrentLocation,
  /* shouldUpdateScroll, */
});

// After a transition:
scrollBehavior.updateScroll(/* prevContext, context */);
```

## Guide

### Installation

```
$ npm i -S scroll-behavior
```

### Basic usage

Create a `ScrollBehavior` object with the following arguments:
- `addTransitionHook`: this function should take a transition hook function and return an unregister function
  - The transition hook function should be called immediately before a transition updates the page
  - The unregister function should remove the transition hook when called
- `stateStorage`: this object should implement `read` and `save` methods
  - The `save` method should take a location object, a nullable element key, and a truthy value; it should save that value for the duration of the page session
  - The `read` method should take a location object and a nullable element key; it should return the value that `save` was called with for that location and element key, or a falsy value if no saved value is available
- `getCurrentLocation`: this function should return the current location object

This object will keep track of the scroll position. Call the `updateScroll` method on this object after transitions to emulate the default browser scroll behavior on page changes.

Call the `stop` method to tear down all listeners.

### Custom scroll behavior

You can customize the scroll behavior by providing a `shouldUpdateScroll` callback when constructing the `ScrollBehavior` object. When you call `updateScroll`, you can pass in up to two additional context arguments, which will get passed to this callback.

The callback can return:

- a falsy value to suppress updating the scroll position
- a position array of `x` and `y`, such as `[0, 100]`, to scroll to that position
- a truthy value to emulate the browser default scroll behavior

Assuming we call `updateScroll` with the previous and current location objects:

```js
const scrollBehavior = new ScrollBehavior({
  ...options,
  shouldUpdateScroll: (prevLocation, location) => (
    // Don't scroll if the pathname is the same.
    !prevLocation || location.pathname !== prevLocation.pathname
  ),
});
```

```js
const scrollBehavior = new ScrollBehavior({
  ...options,
  shouldUpdateScroll: (prevLocation, location) => (
    // Scroll to top when attempting to vist the current path.
    prevLocation && location.pathname === prevLocation.pathname ? [0, 0] : true
  ),
});
```

### Scrolling elements other than `window`

Call the `registerElement` method to register an element other than `window` to have managed scroll behavior. Each of these elements needs to be given a unique key at registration time, and can be given an optional `shouldUpdateScroll` callback that behaves as above. This method should also be called with the current context per `updateScroll` above, if applicable, to set up the element's initial scroll position.

```js
scrollBehavior.registerScrollElement(
  key, element, shouldUpdateScroll, context,
);
```

To unregister an element, call the `unregisterElement` method with the key used to register that element.

[build-badge]: https://img.shields.io/travis/taion/scroll-behavior/master.svg
[build]: https://travis-ci.org/taion/scroll-behavior

[npm-badge]: https://img.shields.io/npm/v/scroll-behavior.svg
[npm]: https://www.npmjs.org/package/scroll-behavior

[codecov-badge]: https://img.shields.io/codecov/c/github/taion/scroll-behavior/master.svg
[codecov]: https://codecov.io/gh/taion/scroll-behavior

[discord-badge]: https://img.shields.io/badge/Discord-join%20chat%20%E2%86%92-738bd7.svg
[discord]: https://discord.gg/0ZcbPKXt5bYaNQ46
