Bootstrap Tokenfield
====================
[![NPM version][npm-badge]](http://badge.fury.io/js/bootstrap-tokenfield)
[![Build status][travis-badge]](https://travis-ci.org/sliptree/bootstrap-tokenfield)
[npm-badge]: https://badge.fury.io/js/bootstrap-tokenfield.png
[travis-badge]: https://travis-ci.org/sliptree/bootstrap-tokenfield.png?branch=master

A jQuery tagging / tokenizer input plugin for Twitter's Bootstrap, by the guys from [Sliptree](https://sliptree.com)

Check out the [demo and docs](http://sliptree.github.io/bootstrap-tokenfield/)

### Installation

Requirements: jQuery 1.9+, Bootstrap 3+ (only CSS)

1. Install via npm or bower (recommended) or manually download the package
2. Include `dist/bootstrap-tokenfield.js` or `dist/bootstrap-tokenfield.min.js` in your HTML
3. Include `dist/css/bootstrap-tokenfield.css` in your HTML

### Usage

```js	
$('input').tokenfield()
```

### Features

* Copy & paste tokens with Ctrl+C and Ctrl+V
* Keyboard navigation, delete tokens with keyboard (arrow keys, Shift + arrow keys)
* Select specific tokens with Ctrl + click and Shift + click
* Twitter Typeahead and jQuery UI Autocomplete support

### FAQ

#### How can I prevent duplicate tokens from being entered?

You can use the `tokenfield:createtoken` event for that. Check the `event.attrs` property for token value and label,
and the run your duplicate detection logic. If it's a duplicate token, simply do `event.preventDefault()`.

Here's a simple example that checks if token's value is equal to any of the existing tokens' values.

```js
$('#my-tokenfield').on('tokenfield:createtoken', function (event) {
	var existingTokens = $(this).tokenfield('getTokens');
	$.each(existingTokens, function(index, token) {
		if (token.value === event.attrs.value)
			event.preventDefault();
	});
});
```

#### And how about limiting tokens to my typeahead/autocomplete data?

Similarly, using `tokenfield:createtoken`, you can check to see if a token exists in your autocomplete/typeahead
data. This example checks if the given token already exists and stops its entry if it doesn't.

```js
$('#my-tokenfield').on('tokenfield:createtoken', function (event) {
	var available_tokens = bloodhound_tokens.index.datums
	var exists = true;
	$.each(available_tokens, function(index, token) {
		if (token.value === event.attrs.value)
			exists = false;
	});
	if(exists === true)
		event.preventDefault();
})
```



### Changelog

See [release notes](https://github.com/sliptree/bootstrap-tokenfield/releases)

Previous releases:

0.10.0

* Fixed: Entering a duplicate token does not submit the underlying form anymore
* Fixed: Selecting a duplicate token from autocomplete or typeahead suggestions no longer clears the input
* Improved: Trying to enter a duplicate tag now animates the existing tag for a little while
* Improved: Tokenfield input has now `autocomplete="off"` to prevent browser-specific autocomplete suggestions
* Changed: `triggerKeys` has been renamed to `delimiter` and accepts a single or an array of characters instead of character codes.

0.9.9-1

* Fixed: setTokens now respects `triggerKeys` option

0.9.8 

* New: `triggerKeys` option
* Fixed: Long placeholders are not being cut off anymore when initializing tokenfield with no tokens #37
* Fixed: createTokensOnBlur no more breaks token editing #35

0.9.7 Valuable

* Fixed: Twitter Typeahead valueKey support #18
* Fixed: Removing multiple tokens returned wrong data #30
* Fixed: If token is removed in beforeEdit event, no longer falls over #27, #28
* Fixed: Change event was triggered on initialization #22
* Fixed: When token is removed in tokenfield:preparetoken event, no longer tries to create a token
* Fixed: Pressing comma key was not handled reliably
* New: `prevetDuplicateToken` event
* Improved: Typeahead integration
* Improved: styling
* Minor tweaks, fixes, improvements 

0.9.5 Typeable

* New: Twitter Typeahead support
* New: When triggering 'change' event on original input, setTokens is now called. This allows you to update tokens externally.
* Fixed: Nnput labels did not work with tokenfield
* Fixed: Set correct input width on fixed-width inputs

0.9.2 Maintenance release

* Many small fixes and improvements

0.9.0 Bootstrappable

* New: Bootstrap 3 support
* New: Input group support
* New: Disable/enable tokenfield
* New: Tokenfield is now responsive
* Deprecated: Bootstrap 2 support

0.7.1 

* Fixed: pressing comma did not create a token in Firefox
* Fixed: tokenfield('getTokensList') returned array instead of string

0.7.0 Autocompleted

* New feature: jQuery UI Autocomplete support

0.6.7 Crossable

* Fixed: Firefox close icon was misplaced
* New: IE 8-10 support (both CSS and Javascript)

0.6.5 Shiftable

* New feature: select specific tokens with Ctrl/Shift + click
* New feature: select specific tokens with Shift + arrow keys
* Internal API improvements

0.6 Editable

* New feature: Edit existing tokens by double-clicking or pressing enter
* A lot of improvements and bugfixes

0.5 Initial release
