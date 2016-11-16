import { insertRule } from './index.js';

/*! normalize.css v3.0.2 | MIT License | git.io/normalize */

/**
 * 1. Set default font family to sans-serif.
 * 2. Prevent iOS text size adjust after orientation change, without disabling
 *    user zoom.
 */

insertRule('html {\n  font-family: sans-serif; /* 1 */\n  -ms-text-size-adjust: 100%; /* 2 */\n  -webkit-text-size-adjust: 100%; /* 2 */\n}');

/**
 * Remove default margin.
 */

insertRule('body {\n  margin: 0;\n}');

/* HTML5 display definitions
   ========================================================================== */

/**
 * Correct `block` display not defined for any HTML5 element in IE 8/9.
 * Correct `block` display not defined for `details` or `summary` in IE 10/11
 * and Firefox.
 * Correct `block` display not defined for `main` in IE 11.
 */

insertRule('article,\naside,\ndetails,\nfigcaption,\nfigure,\nfooter,\nheader,\nhgroup,\nmain,\nmenu,\nnav,\nsection,\nsummary {\n  display: block;\n}');

/**
 * 1. Correct `inline-block` display not defined in IE 8/9.
 * 2. Normalize vertical alignment of `progress` in Chrome, Firefox, and Opera.
 */

insertRule('audio,\ncanvas,\nprogress,\nvideo {\n  display: inline-block; /* 1 */\n  vertical-align: baseline; /* 2 */\n}');

/**
 * Prevent modern browsers from displaying `audio` without controls.
 * Remove excess height in iOS 5 devices.
 */

insertRule('audio:not([controls]) {\n  display: none;\n  height: 0;\n}');

/**
 * Address `[hidden]` styling not present in IE 8/9/10.
 * Hide the `template` element in IE 8/9/11, Safari, and Firefox < 22.
 */

insertRule('[hidden], template {\n  display: none;\n}');

/* Links
   ========================================================================== */

/**
 * Remove the gray background color from active links in IE 10.
 */

insertRule('a {\n  background-color: transparent;\n}');

/**
 * Improve readability when focused and also mouse hovered in all browsers.
 */

insertRule('a:active,\na:hover {\n  outline: 0;\n}');

/* Text-level semantics
   ========================================================================== */

/**
 * Address styling not present in IE 8/9/10/11, Safari, and Chrome.
 */

insertRule('abbr[title] {\n  border-bottom: 1px dotted;\n}');

/**
 * Address style set to `bolder` in Firefox 4+, Safari, and Chrome.
 */

insertRule('b,\nstrong {\n  font-weight: bold;\n}');

/**
 * Address styling not present in Safari and Chrome.
 */

insertRule('dfn {\n  font-style: italic;\n}');

/**
 * Address variable `h1` font-size and margin within `section` and `article`
 * contexts in Firefox 4+, Safari, and Chrome.
 */

insertRule('h1 {\n  font-size: 2em;\n  margin: 0.67em 0;\n}');

/**
 * Address styling not present in IE 8/9.
 */

insertRule('mark {\n  background: #ff0;\n  color: #000;\n}');

/**
 * Address inconsistent and variable font size in all browsers.
 */

insertRule('small {\n  font-size: 80%;\n}');

/**
 * Prevent `sub` and `sup` affecting `line-height` in all browsers.
 */

insertRule('sub, sup {\n  font-size: 75%;\n  line-height: 0;\n  position: relative;\n  vertical-align: baseline;\n}');

insertRule('sup {\n  top: -0.5em;\n}');

insertRule('sub {\n  bottom: -0.25em;\n}');

/* Embedded content
   ========================================================================== */

/**
 * Remove border when inside `a` element in IE 8/9/10.
 */

insertRule('img {\n  border: 0;\n}');

/**
 * Correct overflow not hidden in IE 9/10/11.
 */

insertRule('svg:not(:root) {\n  overflow: hidden;\n}');

/* Grouping content
   ========================================================================== */

/**
 * Address margin not present in IE 8/9 and Safari.
 */

insertRule('figure {\n  margin: 1em 40px;\n}');

/**
 * Address differences between Firefox and other browsers.
 */

insertRule('hr {\n  -moz-box-sizing: content-box;\n  box-sizing: content-box;\n  height: 0;\n}');

/**
 * Contain overflow in all browsers.
 */

insertRule('pre {\n  overflow: auto;\n}');

/**
 * Address odd `em`-unit font size rendering in all browsers.
 */

insertRule('code, kbd, pre, samp {\n  font-family: monospace, monospace;\n  font-size: 1em;\n}');

/* Forms
   ========================================================================== */

/**
 * Known limitation: by default, Chrome and Safari on OS X allow very limited
 * styling of `select`, unless a `border` property is set.
 */

/**
 * 1. Correct color not being inherited.
 *    Known issue: affects color of disabled elements.
 * 2. Correct font properties not being inherited.
 * 3. Address margins set differently in Firefox 4+, Safari, and Chrome.
 */

insertRule('button,\ninput,\noptgroup,\nselect,\ntextarea {\n  color: inherit; /* 1 */\n  font: inherit; /* 2 */\n  margin: 0; /* 3 */\n}');

/**
 * Address `overflow` set to `hidden` in IE 8/9/10/11.
 */

insertRule('button {\n  overflow: visible;\n}');

/**
 * Address inconsistent `text-transform` inheritance for `button` and `select`.
 * All other form control elements do not inherit `text-transform` values.
 * Correct `button` style inheritance in Firefox, IE 8/9/10/11, and Opera.
 * Correct `select` style inheritance in Firefox.
 */

insertRule('button, select {\n  text-transform: none;\n}');

/**
 * 1. Avoid the WebKit bug in Android 4.0.* where (2) destroys native `audio`
 *    and `video` controls.
 * 2. Correct inability to style clickable `input` types in iOS.
 * 3. Improve usability and consistency of cursor style between image-type
 *    `input` and others.
 */

insertRule('button,\nhtml input[type="button"], /* 1 */\ninput[type="reset"],\ninput[type="submit"] {\n  -webkit-appearance: button; /* 2 */\n  cursor: pointer; /* 3 */\n}');

/**
 * Re-set default cursor for disabled elements.
 */

insertRule('button[disabled],\nhtml input[disabled] {\n  cursor: default;\n}');

/**
 * Remove inner padding and border in Firefox 4+.
 */

insertRule('button::-moz-focus-inner,\ninput::-moz-focus-inner {\n  border: 0;\n  padding: 0;\n}');

/**
 * Address Firefox 4+ setting `line-height` on `input` using `!important` in
 * the UA stylesheet.
 */

insertRule('input {\n  line-height: normal;\n}');

/**
 * It's recommended that you don't attempt to style these elements.
 * Firefox's implementation doesn't respect box-sizing, padding, or width.
 *
 * 1. Address box sizing set to `content-box` in IE 8/9/10.
 * 2. Remove excess padding in IE 8/9/10.
 */

insertRule('input[type="checkbox"], input[type="radio"] {\n  box-sizing: border-box; /* 1 */\n  padding: 0; /* 2 */\n}');

/**
 * Fix the cursor style for Chrome's increment/decrement buttons. For certain
 * `font-size` values of the `input`, it causes the cursor style of the
 * decrement button to change from `default` to `text`.
 */

insertRule('input[type="number"]::-webkit-inner-spin-button,\ninput[type="number"]::-webkit-outer-spin-button {\n  height: auto;\n}');

/**
 * 1. Address `appearance` set to `searchfield` in Safari and Chrome.
 * 2. Address `box-sizing` set to `border-box` in Safari and Chrome
 *    (include `-moz` to future-proof).
 */

insertRule('input[type="search"] {\n  -webkit-appearance: textfield; /* 1 */\n  -moz-box-sizing: content-box;\n  -webkit-box-sizing: content-box; /* 2 */\n  box-sizing: content-box;\n}');

/**
 * Remove inner padding and search cancel button in Safari and Chrome on OS X.
 * Safari (but not Chrome) clips the cancel button when the search input has
 * padding (and `textfield` appearance).
 */

insertRule('input[type="search"]::-webkit-search-cancel-button,\ninput[type="search"]::-webkit-search-decoration {\n  -webkit-appearance: none;\n}');

/**
 * Define consistent border, margin, and padding.
 */

insertRule('fieldset {\n  border: 1px solid #c0c0c0;\n  margin: 0 2px;\n  padding: 0.35em 0.625em 0.75em;\n}');

/**
 * 1. Correct `color` not being inherited in IE 8/9/10/11.
 * 2. Remove padding so people aren't caught out if they zero out fieldsets.
 */

insertRule('legend {\n  border: 0; /* 1 */\n  padding: 0; /* 2 */\n}');

/**
 * Remove default vertical scrollbar in IE 8/9/10/11.
 */

insertRule('textarea {\n  overflow: auto;\n}');

/**
 * Don't inherit the `font-weight` (applied by a rule above).
 * NOTE: the default cannot safely be changed in Chrome and Safari on OS X.
 */

insertRule('optgroup {\n  font-weight: bold;\n}');

/* Tables
   ========================================================================== */

/**
 * Remove most spacing between table cells.
 */

insertRule('table {\n  border-collapse: collapse;\n  border-spacing: 0;\n}');

insertRule('td, th {\n  padding: 0;\n}');