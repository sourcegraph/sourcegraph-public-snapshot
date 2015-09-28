# [postcss][postcss]-font-family [![Build Status](https://travis-ci.org/ben-eb/postcss-font-family.svg?branch=master)][ci] [![NPM version](https://badge.fury.io/js/postcss-font-family.svg)][npm] [![Dependency Status](https://gemnasium.com/ben-eb/postcss-font-family.svg)][deps]

> Normalize font family declarations with PostCSS.

## Install

With [npm](https://npmjs.org/package/postcss-font-family) do:

```
npm install postcss-font-family --save
```

## Example

This module will try to minimise the `font-family` and `font` shorthand
properties; it can unquote font families where necessary, detect & remove
duplicates, and cut short a declaration after it finds a keyword. For more
examples, see the [tests](test.js).

### Input

```css
h1 {
    font-family: "Helvetica Neue", Arial, sans-serif, Helvetica;
}
```

### Output

```css
h1 {
    font-family: Helvetica Neue,Arial,sans-serif;
}
```

## Usage

See the [PostCSS documentation](https://github.com/postcss/postcss#usage) for
examples for your environment.

## Contributing

Pull requests are welcome. If you add functionality, then please add unit tests
to cover it.

## License

MIT © [Ben Briggs](http://beneb.info)

[ci]:      https://travis-ci.org/ben-eb/postcss-font-family
[deps]:    https://gemnasium.com/ben-eb/postcss-font-family
[npm]:     http://badge.fury.io/js/postcss-font-family
[postcss]: https://github.com/postcss/postcss
