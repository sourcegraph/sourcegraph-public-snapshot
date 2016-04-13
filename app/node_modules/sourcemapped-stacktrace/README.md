This is a simple module for applying source maps to JS stack traces in the browser. 

## The problem this solves

You have Error.stack() in JS (maybe you're logging a trace, or you're looking at
traces in Jasmine or Mocha), and you need to apply a sourcemap so you can
understand whats happening because you're using some fancy compiled-to-js thing
like coffeescript or traceur. Unfortunately, the browser only applies sourcemaps when the
trace is viewed in its console, not to the underlying stack object, so you're
out of luck.

## Demo

http://novocaine.github.io/sourcemapped-stacktrace-demo/public_html/smst.html

## Install from npm

```
npm install sourcemapped-stacktrace
```

## Setup

Include sourcemapped-stacktrace.js into your page using either an AMD module
loader or a plain old script include. As an AMD module it exposes the method
'mapStackTrace'. If an AMD loader is not found this function will be set on
window.sourceMappedStackTrace.mapStackTrace.

## API 

### mapStackTrace(stack, done)

Re-map entries in a stacktrace using sourcemaps if available.

**Arguments:**

*stack*: Array of strings from the browser's stack representation. Currently only Chrome 
and Firefox format is supported.

*done*: Callback invoked with the transformed stacktrace (an Array of Strings) passed as the first argument

## Example

```javascript
try {
  // break something
  bork();
} catch (e) {
  // pass e.stack to window.mapStackTrace
  window.mapStackTrace(e.stack, function(mappedStack) {
    // do what you want with mappedStack here
    console.log(mappedStack.join("\n"));
  });
}
```

## Longer Explanation

Several modern browsers support sourcemaps when viewing stack traces from errors in their native console, but as of the time of writing there is no support for applying a sourcemap to the (highly non-standardised) [Error.prototype.stack](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Error/Stack). Error.prototype.stack can be used for logging errors and for displaying errors in test frameworks, and it is not very convenient to have unmapped traces in either of those use cases.

This module fetches all the scripts referenced by the stack trace, determines
whether they have an applicable sourcemap, fetches the sourcemap from the
server, then uses the [Mozilla source-map library](https://github.com/mozilla/source-map/) to do the mapping. Browsers that support sourcemaps don't offer a standardised sourcemap API, so we have to do all that work ourselves.

The nice part about doing it ourselves is that the library could be extended to
work in browsers that don't support sourcemaps, which could be good for
logging and debugging problems. Currently, only Chrome and Firefox are supported, but it
would be easy to support those formats by ripping off [stacktrace.js](https://github.com/stacktracejs/stacktrace.js/).

## Known issues

* Doesn't support exception formats of any browser other than Chrome and
  Firefox
* Only supports JS containing //# sourceMappingURL= declarations (i.e. no
  support for the SourceMap: HTTP header (yet)
* Some prominent sourcemap generators (including CoffeeScript, Traceur, Babel)
  don't emit a list of 'names' in the source-map, which means that frames from transpiled code will have (unknown) instead of the original function name. Those generators should support this feature better.
