# Fixed-sticky [![Build Status](https://travis-ci.org/filamentgroup/fixed-sticky.svg?branch=master)](https://travis-ci.org/filamentgroup/fixed-sticky)

A CSS `position:sticky` polyfill.

[![Filament Group](http://filamentgroup.com/images/fg-logo-positive-sm-crop.png) ](http://www.filamentgroup.com/)

- ©2013 [@zachleat](https://github.com/zachleat), Filament Group
- MIT license

## Browser Support

CSS position:sticky is really in its infancy in terms of browser support. In stock browsers, it is currently only available in iOS 6.

~~In Chrome you can enable it by navigating to `chrome://flags` and enabling experimental “WebKit features” or “Web Platform features” (Canary).~~ Chrome temporarily removed their native `position: sticky` implementation.

In Firefox you you can go to `about:config` and set `layout.css.sticky.enabled` to "true".

## Important

The most overlooked thing about `position: sticky` is that `sticky` elements are constrained to the dimensions of their parent elements. This means if a `sticky` element is inside of a parent container that is the same dimensions as itself, the element will not stick.

Here’s an example of what a `sticky` element with CSS `top: 20px` behaves like:

![](demos/gifs/sticky-top-off.gif)

*Scrolling down.* The blue border represents the dimensions of the parent container element. If the element’s top is greater than `20px` to the top of the viewport, the element is not sticky.

![](demos/gifs/sticky-top-on.gif)

*Scrolling down.* When the element’s top is less than `20px` to the top of the viewport, the element is sticky.

Here’s an example of what a `sticky` element with CSS `bottom: 20px` behaves like:

![](demos/gifs/sticky-bottom-off.gif)

*Scrolling up.* Not sticky.

![](demos/gifs/sticky-bottom-on.gif)

*Scrolling up.* Sticky.

## Plugin Usage

Just qualify element you’d like to be `position:sticky` with a `fixedsticky` class.

    <div id="my-element" class="fixedsticky">

Add your own CSS to position the element. Supports any value for `top` or `bottom`.

    .fixedsticky { top: 0; }

Next, add the events and initialize your sticky nodes:

    $( '#my-element' ).fixedsticky();

*Note: if you’re going to use non-zero values for `top` or `bottom`, fixed-sticky is victim to a cross-browser incompatibility with jQuery’s `css` method (namely, IE8- doesn’t normalize non-pixel values to pixels). Use pixels (or `0`) for best cross-browser compatibility.*

Optionally, you may also destroy the component:

    $( '#my-element' ).fixedsticky( 'destroy' );

## Demos
* For a fixed-sticky demo, open [`demo.html`](http://filamentgroup.github.com/fixed-sticky/demos/demo.html).
* For a pure native position: sticky test, open [`demo-control.html`](http://filamentgroup.github.com/fixed-sticky/demos/demo-control.html).

## Native `position: sticky` Caveats

* Any non-default value (not `visible`) for `overflow`, `overflow-x`, or `overflow-y` on the parent element will disable `position: sticky` (via [@davatron5000](https://twitter.com/davatron5000/status/434357818498351104)).
* iOS ~~(and Chrome)~~ do not support `position: sticky;` with `display: inline-block;`.
* This plugin ~~(and Chrome’s implementation)~~ does not (yet) support use with `thead` and `tfoot`.
* Native `sticky` anchors to parent elements using their own overflow. This means scrolling the element fixes the sticky element to the parent dimensions. This plugin does not support overflow on parent elements.

### Using the polyfill instead of native

If you’re having weird issues with native `position: sticky`, you can tell fixed-sticky to use the polyfill instead of native. Just override the sticky feature test to always return false. Make sure you do this before any calls to `$( '#my-element' ).fixedsticky();`.

    // After fixed-sticky.js
    FixedSticky.tests.sticky = false;

* [`demo-opt-out-native.html`](http://filamentgroup.github.com/fixed-sticky/demos/demo-opt-out-native.html) shows this behavior.

## Installation

Use the provided `fixedsticky.js` and `fixedsticky.css` files.

### Also available in [Bower](http://bower.io/)

    bower install filament-sticky

## Browser Support

These tests were performed using fixed-sticky with fixed-fixed. It’s safest to use them together (`position:fixed` is a minefield on older devices), but they can be used independently.

### Native Sticky

* iOS 6.1+

### Polyfilled

* Internet Explorer 7, 8, 9, 10
* Firefox 24, Firefox 17 ESR
* Chrome 29
* Safari 6.0.5
* Opera 12.16
* Android 4.X

### Fallback (static positioning)

* Android 2.X
* Opera Mini
* Blackberry OS 5, 6, 7
* Windows Phone 7.5

## TODO

* Add support for table headers.
* Vanilla JS version.
* Make sticky smoother on transition between sticky/static for container based

## [Tests](http://filamentgroup.github.io/fixed-sticky/test/fixed-sticky.html)

## Release History

* `v0.1.0`: Initial release.
* `v0.1.3`: Bug fixes, rudimentary tests, destroy method.
