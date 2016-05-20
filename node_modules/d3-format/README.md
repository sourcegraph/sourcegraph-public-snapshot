# d3-format

Ever noticed how sometimes JavaScript doesn’t display numbers the way you expect? Like, you tried to print tenths with a simple loop:

```js
for (var i = 0; i < 10; i++) {
  console.log(0.1 * i);
}
```

And you got this:

```js
0
0.1
0.2
0.30000000000000004
0.4
0.5
0.6000000000000001
0.7000000000000001
0.8
0.9
```

Welcome to [binary floating point](https://en.wikipedia.org/wiki/Double-precision_floating-point_format)! ಠ_ಠ

Yet rounding error is not the only reason to customize number formatting. A table of numbers should be formatted consistently for comparison; above, 0.0 would be better than 0. Large numbers should have grouped digits (e.g., 42,000) or be in scientific or metric notation (4.2e+4, 42k). Currencies should have fixed precision ($3.50). Reported numerical results should be rounded to significant digits (4021 becomes 4000). Number formats should appropriate to the reader’s locale (42.000,00 or 42,000.00). The list goes on.

Formatting numbers for human consumption is the purpose of d3-format, which is modeled after Python 3’s [format specification mini-language](https://docs.python.org/3/library/string.html#format-specification-mini-language) ([PEP 3101](https://www.python.org/dev/peps/pep-3101/)). Revisiting the example above:

```js
var f = d3.format(".1f");
for (var i = 0; i < 10; i++) {
  console.log(f(0.1 * i));
}
```

Now you get this:

```js
0.0
0.1
0.2
0.3
0.4
0.5
0.6
0.7
0.8
0.9
```

But d3-format is much more than an alias for [number.toFixed](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/toFixed)! A few more examples:

```js
d3.format(".0%")(0.123);  // rounded percentage, "12%"
d3.format("($.2f")(-3.5); // localized fixed-point currency, "(£3.50)"
d3.format("+20")(42);     // space-filled and signed, "                 +42"
d3.format(".^20")(42);    // dot-filled and centered, ".........42........."
d3.format(".2s")(42e6);   // SI-prefix with two significant digits, "42M"
d3.format("#x")(48879);   // prefixed lowercase hexadecimal, "0xbeef"
d3.format(",.2r")(4223);  // grouped thousands with two significant digits, "4,200"
```

See [*locale*.format](#locale_format) for a detailed specification, and try running [formatSpecifier](#formatSpecifier) on the above formats to decode their meaning.

## Installing

If you use NPM, `npm install d3-format`. Otherwise, download the [latest release](https://github.com/d3/d3-format/releases/latest). The released bundle supports AMD, CommonJS, and vanilla environments. Create a custom build using [Rollup](https://github.com/rollup/rollup) or your preferred bundler. You can also load directly from [d3js.org](https://d3js.org):

```html
<script src="https://d3js.org/d3-format.v0.5.min.js"></script>
```

In a vanilla environment, a `d3_format` global is exported. [Try d3-format in your browser.](https://tonicdev.com/npm/d3-format)

## API Reference

<a name="format" href="#format">#</a> d3.<b>format</b>(<i>specifier</i>)

An alias for [*locale*.format](#locale_format) on the [U.S. English locale](#formatEnUs). See the other [locales](#locales), or use [formatLocale](#formatLocale) to define a new locale.

<a name="formatPrefix" href="#formatPrefix">#</a> d3.<b>formatPrefix</b>(<i>specifier</i>, <i>value</i>)

An alias for [*locale*.formatPrefix](#locale_formatPrefix) on the [U.S. English locale](#formatEnUs). See the other [locales](#locales), or use [formatLocale](#formatLocale) to define a new locale.

<a name="locale_format" href="#locale_format">#</a> <i>locale</i>.<b>format</b>(<i>specifier</i>)

Returns a new format function for the given string *specifier*. The returned function takes a number as the only argument, and returns a string representing the formatted number. The general form of a specifier is:

```
[​[fill]align][sign][symbol][0][width][,][.precision][type]
```

The *fill* can be any character. The presence of a fill character is signaled by the *align* character following it, which must be one of the following:

* `>` - Forces the field to be right-aligned within the available space. (Default behavior).
* `<` - Forces the field to be left-aligned within the available space.
* `^` - Forces the field to be centered within the available space.
* `=` - like `>`, but with any sign and symbol to the left of any padding.

The *sign* can be:

* `-` - nothing for positive and a minus sign for negative. (Default behavior.)
* `+` - a plus sign for positive and a minus sign for negative.
* `(` - nothing for positive and parentheses for negative.
* ` ` (space) - a space for positive and a minus sign for negative.

The *symbol* can be:

* `$` - apply currency symbols per the locale definition.
* `#` - for binary, octal, or hexadecimal notation, prefix by `0b`, `0o`, or `0x`, respectively.

The *zero* (`0`) option enables zero-padding; this implicitly sets *fill* to `0` and *align* to `=`. The *width* defines the minimum field width; if not specified, then the width will be determined by the content. The *comma* (`,`) option enables the use of a group separator, such as a comma for thousands.

Depending on the *type*, the *precision* either indicates the number of digits that follow the decimal point (types `f` and `%`), or the number of significant digits (types `​`, `e`, `g`, `r`, `s` and `p`). If the precision is not specified, it defaults to 6 for all types except `​` (none), which defaults to 12. Precision is ignored for integer formats (types `b`, `o`, `d`, `x`, `X` and `c`). See [precisionFixed](#precisionFixed) and [precisionRound](#precisionRound) for help picking an appropriate precision.

The available *type* values are:

* `e` - exponent notation.
* `f` - fixed point notation.
* `g` - either decimal or exponent notation, rounded to significant digits.
* `r` - decimal notation, rounded to significant digits.
* `s` - decimal notation with an [SI prefix](#locale_formatPrefix), rounded to significant digits.
* `%` - multiply by 100, and then decimal notation with a percent sign.
* `p` - multiply by 100, round to significant digits, and then decimal notation with a percent sign.
* `b` - binary notation, rounded to integer.
* `o` - octal notation, rounded to integer.
* `d` - decimal notation, rounded to integer.
* `x` - hexadecimal notation, using lower-case letters, rounded to integer.
* `X` - hexadecimal notation, using upper-case letters, rounded to integer.
* `c` - converts the integer to the corresponding unicode character before printing.
* `​` (none) - like `g`, but trim insignificant trailing zeros.

The type `n` is also supported as shorthand for `,g`. For the `g`, `n` and `​` (none) types, decimal notation is used if the resulting string would have *precision* or fewer digits; otherwise, exponent notation is used. For example:

```js
d3.format(".2")(42);  // "42"
d3.format(".2")(4.2); // "4.2"
d3.format(".1")(42);  // "4e+1"
d3.format(".1")(4.2); // "4"
```

<a name="locale_formatPrefix" href="#locale_formatPrefix">#</a> <i>locale</i>.<b>formatPrefix</b>(<i>specifier</i>, <i>value</i>)

Equivalent to [*locale*.format](#locale_format), except the returned function will convert values to the units of the appropriate [SI prefix](https://en.wikipedia.org/wiki/Metric_prefix#List_of_SI_prefixes) for the specified numeric reference *value* before formatting in fixed point notation. The following prefixes are supported:

* `y` - yocto, 10⁻²⁴
* `z` - zepto, 10⁻²¹
* `a` - atto, 10⁻¹⁸
* `f` - femto, 10⁻¹⁵
* `p` - pico, 10⁻¹²
* `n` - nano, 10⁻⁹
* `µ` - micro, 10⁻⁶
* `m` - milli, 10⁻³
* `​` (none) - 10⁰
* `k` - kilo, 10³
* `M` - mega, 10⁶
* `G` - giga, 10⁹
* `T` - tera, 10¹²
* `P` - peta, 10¹⁵
* `E` - exa, 10¹⁸
* `Z` - zetta, 10²¹
* `Y` - yotta, 10²⁴

Unlike [*locale*.format](#locale_format) with the `s` format type, this method returns a formatter with a consistent SI prefix, rather than computing the prefix dynamically for each number. In addition, the *precision* for the given *specifier* represents the number of digits past the decimal point (as with `f` fixed point notation), not the number of significant digits. For example:

```js
var f = d3.formatPrefix(",.0", 1e-6);
f(0.00042); // "420µ"
f(0.0042); // "4,200µ"
```

This method is useful when formatting multiple numbers in the same units for easy comparison. See [precisionPrefix](#precisionPrefix) for help picking an appropriate precision, and [bl.ocks.org/9764126](http://bl.ocks.org/mbostock/9764126) for an example.

<a name="formatSpecifier" href="#formatSpecifier">#</a> d3.<b>formatSpecifier</b>(<i>specifier</i>)

Parses the specified *specifier*, returning an object with exposed fields that correspond to the [format specification mini-language](#locale_format) and a toString method that reconstructs the specifier. For example, `formatSpecifier("s")` returns:

```js
{
  "fill": " ",
  "align": ">",
  "sign": "-",
  "symbol": "",
  "zero": false,
  "width": undefined,
  "comma": false,
  "precision": 6,
  "type": "s"
}
```

This method is useful for understanding how format specifiers are parsed and for deriving new specifiers. For example, you might compute an appropriate precision based on the numbers you want to format using [precisionFixed](#precisionFixed) and then create a new format:

```js
var s = d3.formatSpecifier("f");
s.precision = precisionFixed(0.01);
var f = d3.format(s);
f(42); // "42.00";
```

<a name="precisionFixed" href="#precisionFixed">#</a> d3.<b>precisionFixed</b>(<i>step</i>)

Returns a suggested decimal precision for fixed point notation given the specified numeric *step* value. The *step* represents the minimum absolute difference between values that will be formatted. (This assumes that the values to be formatted are also multiples of *step*.) For example, given the numbers 1, 1.5, and 2, the *step* should be 0.5 and the suggested precision is 1:

```js
var p = d3.precisionFixed(0.5),
    f = d3.format("." + p + "f");
f(1);   // "1.0"
f(1.5); // "1.5"
f(2);   // "2.0"
```

Whereas for the numbers 1, 2 and 3, the *step* should be 1 and the suggested precision is 0:

```js
var p = d3.precisionFixed(1),
    f = d3.format("." + p + "f");
f(1); // "1"
f(2); // "2"
f(3); // "3"
```

Note: for the `%` format type, subtract two:

```js
var p = Math.max(0, d3.precisionFixed(0.05) - 2),
    f = d3.format("." + p + "%");
f(0.45); // "45%"
f(0.50); // "50%"
f(0.55); // "55%"
```

<a name="precisionPrefix" href="#precisionPrefix">#</a> d3.<b>precisionPrefix</b>(<i>step</i>, <i>value</i>)

Returns a suggested decimal precision for use with [*locale*.formatPrefix](#locale_formatPrefix) given the specified numeric *step* and reference *value*. The *step* represents the minimum absolute difference between values that will be formatted, and *value* determines which SI prefix will be used. (This assumes that the values to be formatted are also multiples of *step*.) For example, given the numbers 1.1e6, 1.2e6, and 1.3e6, the *step* should be 1e5, the *value* could be 1.3e6, and the suggested precision is 1:

```js
var p = d3.precisionPrefix(1e5, 1.3e6),
    f = d3.formatPrefix("." + p, 1.3e6);
f(1.1e6); // "1.1M"
f(1.2e6); // "1.2M"
f(1.3e6); // "1.3M"
```

<a name="precisionRound" href="#precisionRound">#</a> d3.<b>precisionRound</b>(<i>step</i>, <i>max</i>)

Returns a suggested decimal precision for format types that round to significant digits given the specified numeric *step* and *max* values. The *step* represents the minimum absolute difference between values that will be formatted, and the *max* represents the largest absolute value that will be formatted. (This assumes that the values to be formatted are also multiples of *step*.) For example, given the numbers 0.99, 1.0, and 1.01, the *step* should be 0.01, the *max* should be 1.01, and the suggested precision is 3:

```js
var p = d3.precisionRound(0.01, 1.01),
    f = d3.format("." + p + "r");
f(0.99); // "0.990"
f(1.0);  // "1.00"
f(1.01); // "1.01"
```

Whereas for the numbers 0.9, 1.0, and 1.1, the *step* should be 0.1, the *max* should be 1.1, and the suggested precision is 2:

```js
var p = d3.precisionRound(0.1, 1.1),
    f = d3.format("." + p + "r");
f(0.9); // "0.90"
f(1.0); // "1.0"
f(1.1); // "1.1"
```

Note: for the `e` format type, subtract one:

```js
var p = Math.max(0, d3.precisionRound(0.01, 1.01) - 1),
    f = d3.format("." + p + "e");
f(0.01); // "1.00e-2"
f(1.01); // "1.01e+0"
```

### Locales

<a name="formatLocale" href="#formatLocale">#</a> d3.<b>formatLocale</b>(<i>definition</i>)

Returns a *locale* object for the specified *definition* with [*locale*.format](#locale_format) and [*locale*.formatPrefix](#locale_formatPrefix) methods. The *definition* must include the following properties:

* `decimal` - the decimal point (e.g., `"."`).
* `thousands` - the group separator (e.g., `","`).
* `grouping` - the array of group sizes (e.g., `[3]`), cycled as needed.
* `currency` - the currency prefix and suffix (e.g., `["$", ""]`).

Note that the *thousands* property is a misnomer, as the grouping definition allows groups other than thousands.

<a name="formatCaEs" href="#formatCaEs">#</a> d3.<b>formatCaEs</b>

[Catalan (Spain)](https://github.com/d3/d3-format/tree/master/src/locale/ca-ES.js)

<a name="formatCsCz" href="#formatCsCz">#</a> d3.<b>formatCsCz</b>

[Czech (Czech Republic)](https://github.com/d3/d3-format/tree/master/src/locale/cs-CZ.js)

<a name="formatDeCh" href="#formatDeCh">#</a> d3.<b>formatDeCh</b>

[German (Switzerland)](https://github.com/d3/d3-format/tree/master/src/locale/de-CH.js)

<a name="formatDeDe" href="#formatDeDe">#</a> d3.<b>formatDeDe</b>

[German (Germany)](https://github.com/d3/d3-format/tree/master/src/locale/de-DE.js)

<a name="formatEnCa" href="#formatEnCa">#</a> d3.<b>formatEnCa</b>

[English (Canada)](https://github.com/d3/d3-format/tree/master/src/locale/en-CA.js)

<a name="formatEnGb" href="#formatEnGb">#</a> d3.<b>formatEnGb</b>

[English (United Kingdom)](https://github.com/d3/d3-format/tree/master/src/locale/en-GB.js)

<a name="formatEnUs" href="#formatEnUs">#</a> d3.<b>formatEnUs</b>

[English (United States)](https://github.com/d3/d3-format/tree/master/src/locale/en-US.js)

<a name="formatEsEs" href="#formatEsEs">#</a> d3.<b>formatEsEs</b>

[Spanish (Spain)](https://github.com/d3/d3-format/tree/master/src/locale/es-ES.js)

<a name="formatFiFi" href="#formatFiFi">#</a> d3.<b>formatFiFi</b>

[Finnish (Finland)](https://github.com/d3/d3-format/tree/master/src/locale/fi-FI.js)

<a name="formatFrCa" href="#formatFrCa">#</a> d3.<b>formatFrCa</b>

[French (Canada)](https://github.com/d3/d3-format/tree/master/src/locale/fr-CA.js)

<a name="formatFrFr" href="#formatFrFr">#</a> d3.<b>formatFrFr</b>

[French (France)](https://github.com/d3/d3-format/tree/master/src/locale/fr-FR.js)

<a name="formatHeIl" href="#formatHeIl">#</a> d3.<b>formatHeIl</b>

[Hebrew (Israel)](https://github.com/d3/d3-format/tree/master/src/locale/he-IL.js)

<a name="formatHuHu" href="#formatHuHu">#</a> d3.<b>formatHuHu</b>

[Hungarian (Hungary)](https://github.com/d3/d3-format/tree/master/src/locale/hu-HU.js)

<a name="formatItIt" href="#formatItIt">#</a> d3.<b>formatItIt</b>

[Italian (Italy)](https://github.com/d3/d3-format/tree/master/src/locale/it-IT.js)

<a name="formatJaJp" href="#formatJaJp">#</a> d3.<b>formatJaJp</b>

[Japanese (Japan)](https://github.com/d3/d3-format/tree/master/src/locale/ja-JP.js)

<a name="formatKoKr" href="#formatKoKr">#</a> d3.<b>formatKoKr</b>

[Korean (South Korea)](https://github.com/d3/d3-format/tree/master/src/locale/ko-KR.js)

<a name="formatMkMk" href="#formatMkMk">#</a> d3.<b>formatMkMk</b>

[Macedonian (Macedonia)](https://github.com/d3/d3-format/tree/master/src/locale/mk-MK.js)

<a name="formatNlNl" href="#formatNlNl">#</a> d3.<b>formatNlNl</b>

[Dutch (Netherlands)](https://github.com/d3/d3-format/tree/master/src/locale/nl-NL.js)

<a name="formatPlPl" href="#formatPlPl">#</a> d3.<b>formatPlPl</b>

[Polish (Poland)](https://github.com/d3/d3-format/tree/master/src/locale/pl-PL.js)

<a name="formatPtBr" href="#formatPtBr">#</a> d3.<b>formatPtBr</b>

[Portuguese (Brazil)](https://github.com/d3/d3-format/tree/master/src/locale/pt-BR.js)

<a name="formatRuRu" href="#formatRuRu">#</a> d3.<b>formatRuRu</b>

[Russian (Russia)](https://github.com/d3/d3-format/tree/master/src/locale/ru-RU.js)

<a name="formatSvSe" href="#formatSvSe">#</a> d3.<b>formatSvSe</b>

[Swedish (Sweden)](https://github.com/d3/d3-format/tree/master/src/locale/sv-SE.js)

<a name="formatZhCn" href="#formatZhCn">#</a> d3.<b>formatZhCn</b>

[Chinese (China)](https://github.com/d3/d3-format/tree/master/src/locale/zh-CN.js)
