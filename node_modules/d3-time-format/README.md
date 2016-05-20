# d3-time-format

This module provides a JavaScript implementation of the venerable [strptime](http://pubs.opengroup.org/onlinepubs/009695399/functions/strptime.html) and [strftime](http://pubs.opengroup.org/onlinepubs/007908799/xsh/strftime.html) functions from the C standard library, and can be used to parse or format [dates](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date) in a variety of locale-specific representations. To format a date, create a [formatter](#locale_format) from a specifier (a string with the desired format *directives*, indicated by `%`); then pass a date to the formatter, which returns a string. For example, to convert the current date to a human-readable string:

```js
var formatTime = d3.timeFormat("%B %d, %Y");
formatTime(new Date); // "June 30, 2015"
```

Likewise, to convert a string back to a date, create a [parser](#locale_parse):

```js
var parseTime = d3.timeParse("%B %d, %Y");
parseTime("June 30, 2015"); // Tue Jun 30 2015 00:00:00 GMT-0700 (PDT)
```

You can implement more elaborate conditional time formats, too. For example, here’s a [multi-scale time format](http://bl.ocks.org/mbostock/4149176) using [time intervals](https://github.com/d3/d3-time):

```js
var formatMillisecond = d3.timeFormat(".%L"),
    formatSecond = d3.timeFormat(":%S"),
    formatMinute = d3.timeFormat("%I:%M"),
    formatHour = d3.timeFormat("%I %p"),
    formatDay = d3.timeFormat("%a %d"),
    formatWeek = d3.timeFormat("%b %d"),
    formatMonth = d3.timeFormat("%B"),
    formatYear = d3.timeFormat("%Y");

function multiFormat(date) {
  return (d3.timeSecond(date) < date ? formatMillisecond
      : d3.timeMinute(date) < date ? formatSecond
      : d3.timeHour(date) < date ? formatMinute
      : d3.timeDay(date) < date ? formatHour
      : d3.timeMonth(date) < date ? (d3.timeWeek(date) < date ? formatDay : formatWeek)
      : d3.timeYear(date) < date ? formatMonth
      : formatYear)(date);
}
```

This module is used by D3 [time scales](https://github.com/d3/d3-scale#time-scales) to generate human-readable ticks.

## Installing

If you use NPM, `npm install d3-time-format`. Otherwise, download the [latest release](https://github.com/d3/d3-time-format/releases/latest). You can also load directly from [d3js.org](https://d3js.org), either as a [standalone library](https://d3js.org/d3-time-format.v0.3.min.js) or as part of [D3 4.0 alpha](https://github.com/mbostock/d3/tree/4). AMD, CommonJS, and vanilla environments are supported. In vanilla, a `d3_time_format` global is exported:

```html
<script src="https://d3js.org/d3-time.v0.2.min.js"></script>
<script src="https://d3js.org/d3-time-format.v0.3.min.js"></script>
<script>

var format = d3_time_format.timeFormat("%x");

</script>
```

[Try d3-time-format in your browser.](https://tonicdev.com/npm/d3-time-format)

## API Reference

<a name="timeFormat" href="#timeFormat">#</a> d3.<b>timeFormat</b>(<i>specifier</i>)

An alias for [*locale*.format](#locale_format) on the [U.S. English locale](#timeFormatEnUs). See the other [locales](#locales).

<a name="timeParse" href="#timeParse">#</a> d3.<b>timeParse</b>(<i>specifier</i>)

An alias for [*locale*.parse](#locale_parse) on the [U.S. English locale](#timeFormatEnUs). See the other [locales](#locales).

<a name="utcFormat" href="#utcFormat">#</a> d3.<b>utcFormat</b>(<i>specifier</i>)

An alias for [*locale*.utcFormat](#locale_utcFormat) on the [U.S. English locale](#localeEnUs). See the other [locales](#locales).

<a name="utcParse" href="#utcParse">#</a> d3.<b>utcParse</b>(<i>specifier</i>)

An alias for [*locale*.utcParse](#locale_utcParse) on the [U.S. English locale](#localeEnUs). See the other [locales](#locales).

<a name="isoFormat" href="#isoFormat">#</a> d3.<b>isoFormat</b>

The full [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601) UTC time formatter. Where available, this method will use [Date.toISOString](https://developer.mozilla.org/en-US/docs/JavaScript/Reference/Global_Objects/Date/toISOString) to format.

<a name="isoParse" href="#isoParse">#</a> d3.<b>isoParse</b>

The full [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601) UTC time parser. Where available, this method will use the [Date constructor](https://developer.mozilla.org/en-US/docs/JavaScript/Reference/Global_Objects/Date) to parse strings. If you depend on strict validation of the input format according to ISO 8601, you should construct a [UTC parser function](#utcParse):

```js
var strictIsoParse = d3.utcParse("%Y-%m-%dT%H:%M:%S.%LZ");
```

<a name="locale_format" href="#locale_format">#</a> <i>locale</i>.<b>format</b>(<i>specifier</i>)

Returns a new formatter for the given string *specifier*. The specifier string may contain the following directives:

* `%a` - abbreviated weekday name.*
* `%A` - full weekday name.*
* `%b` - abbreviated month name.*
* `%B` - full month name.*
* `%c` - the locale’s date and time, such as `%a %b %e %H:%M:%S %Y`.*
* `%d` - zero-padded day of the month as a decimal number [01,31].
* `%e` - space-padded day of the month as a decimal number [ 1,31]; equivalent to `%_d`.
* `%H` - hour (24-hour clock) as a decimal number [00,23].
* `%I` - hour (12-hour clock) as a decimal number [01,12].
* `%j` - day of the year as a decimal number [001,366].
* `%m` - month as a decimal number [01,12].
* `%M` - minute as a decimal number [00,59].
* `%L` - milliseconds as a decimal number [000, 999].
* `%p` - either AM or PM.*
* `%S` - second as a decimal number [00,61].
* `%U` - Sunday-based week of the year as a decimal number [00,53].
* `%w` - Sunday-based weekday as a decimal number [0,6].
* `%W` - Monday-based week of the year as a decimal number [00,53].
* `%x` - the locale’s date, such as `%m/%d/%Y`.*
* `%X` - the locale’s time, such as `%H:%M:%S`.*
* `%y` - year without century as a decimal number [00,99].
* `%Y` - year with century as a decimal number.
* `%Z` - time zone offset, such as `-0700`, `-07:00`, `-07`, or `Z`.
* `%%` - a literal percent sign (`%`).

Directives marked with an asterisk (*) may be affected by the [locale definition](#localeFormat). For `%U`, all days in a new year preceding the first Sunday are considered to be in week 0. For `%W`, all days in a new year preceding the first Monday are considered to be in week 0. Week numbers are computed using [*interval*.count](https://github.com/d3/d3-time#interval_count).

The `%` sign indicating a directive may be immediately followed by a padding modifier:

* `0` - zero-padding
* `_` - space-padding
* `-` - disable padding

If no padding modifier is specified, the default is `0` for all directives except `%e`, which defaults to `_`. (In some implementations of strftime and strptime, a directive may include an optional field width or precision; this feature is not yet implemented.)

The returned function formats a specified *[date](https://developer.mozilla.org/en/JavaScript/Reference/Global_Objects/Date)*, returning the corresponding string.

```js
var formatMonth = d3.timeFormat("%B"),
    formatDay = d3.timeFormat("%A"),
    date = new Date(2014, 4, 1); // Thu May 01 2014 00:00:00 GMT-0700 (PDT)

formatMonth(date); // "May"
formatDay(date); // "Thursday"
```

<a name="locale_parse" href="#locale_parse">#</a> <i>locale</i>.<b>parse</b>(<i>specifier</i>)

Returns a new parser for the given string *specifier*. The specifier string may contain the same directives as [*locale*.format](#locale_format). The `%d` and `%e` directives are considered equivalent for parsing.

The returned function parses a specified *string*, returning the corresponding [date](https://developer.mozilla.org/en/JavaScript/Reference/Global_Objects/Date) or null if the string could not be parsed according to this format’s specifier. Parsing is strict: if the specified <i>string</i> does not exactly match the associated specifier, this method returns null. For example, if the associated specifier is `%Y-%m-%dT%H:%M:%SZ`, then the string `"2011-07-01T19:15:28Z"` will be parsed as expected, but `"2011-07-01T19:15:28"`, `"2011-07-01 19:15:28"` and `"2011-07-01"` will return null. (Note that the literal `Z` here is different from the time zone offset directive `%Z`.) If a more flexible parser is desired, try multiple formats sequentially until one returns non-null.

<a name="locale_utcFormat" href="#locale_utcFormat">#</a> <i>locale</i>.<b>utcFormat</b>(<i>specifier</i>)

Equivalent to [*locale*.format](#locale_format), except all directives are interpreted as [Coordinated Universal Time (UTC)](https://en.wikipedia.org/wiki/Coordinated_Universal_Time) rather than local time.

<a name="locale_utcParse" href="#locale_utcParse">#</a> <i>locale</i>.<b>utcParse</b>(<i>specifier</i>)

Equivalent to [*locale*.parse](#locale_parse), except all directives are interpreted as [Coordinated Universal Time (UTC)](https://en.wikipedia.org/wiki/Coordinated_Universal_Time) rather than local time.

### Locales

<a name="timeFormatLocale" href="#timeFormatLocale">#</a> d3.<b>timeFormatLocale</b>(<i>definition</i>)

Returns a *locale* object for the specified *definition* with [*locale*.format](#locale_format), [*locale*.parse](#locale_parse), [*locale*.utcFormat](#locale_utcFormat), [*locale*.utcParse](#locale_utcParse) methods. The *definition* must include the following properties:

* `dateTime` - the date and time (`%c`) format specifier (<i>e.g.</i>, `"%a %b %e %X %Y"`).
* `date` - the date (`%x`) format specifier (<i>e.g.</i>, `"%m/%d/%Y"`).
* `time` - the time (`%X`) format specifier (<i>e.g.</i>, `"%H:%M:%S"`).
* `periods` - the A.M. and P.M. equivalents (<i>e.g.</i>, `["AM", "PM"]`).
* `days` - the full names of the weekdays, starting with Sunday.
* `shortDays` - the abbreviated names of the weekdays, starting with Sunday.
* `months` - the full names of the months (starting with January).
* `shortMonths` - the abbreviated names of the months (starting with January).

<a name="timeFormatCaEs" href="#timeFormatCaEs">#</a> d3.<b>timeFormatCaEs</b>

[Catalan (Spain)](https://github.com/d3/d3-time-format/tree/master/src/locale/ca-ES.js)

<a name="timeFormatDeCh" href="#timeFormatDeCh">#</a> d3.<b>timeFormatDeCh</b>

[German (Switzerland)](https://github.com/d3/d3-time-format/tree/master/src/locale/de-CH.js)

<a name="timeFormatDeDe" href="#timeFormatDeDe">#</a> d3.<b>timeFormatDeDe</b>

[German (Germany)](https://github.com/d3/d3-time-format/tree/master/src/locale/de-DE.js)

<a name="timeFormatEnCa" href="#timeFormatEnCa">#</a> d3.<b>timeFormatEnCa</b>

[English (Canada)](https://github.com/d3/d3-time-format/tree/master/src/locale/en-CA.js)

<a name="timeFormatEnGb" href="#timeFormatEnGb">#</a> d3.<b>timeFormatEnGb</b>

[English (United Kingdom)](https://github.com/d3/d3-time-format/tree/master/src/locale/en-GB.js)

<a name="timeFormatEnUs" href="#timeFormatEnUs">#</a> d3.<b>timeFormatEnUs</b>

[English (United States)](https://github.com/d3/d3-time-format/tree/master/src/locale/en-US.js)

<a name="timeFormatEsEs" href="#timeFormatEsEs">#</a> d3.<b>timeFormatEsEs</b>

[Spanish (Spain)](https://github.com/d3/d3-time-format/tree/master/src/locale/es-ES.js)

<a name="timeFormatFiFi" href="#timeFormatFiFi">#</a> d3.<b>timeFormatFiFi</b>

[Finnish (Finland)](https://github.com/d3/d3-time-format/tree/master/src/locale/fi-FI.js)

<a name="timeFormatFrCa" href="#timeFormatFrCa">#</a> d3.<b>timeFormatFrCa</b>

[French (Canada)](https://github.com/d3/d3-time-format/tree/master/src/locale/fr-CA.js)

<a name="timeFormatFrFr" href="#timeFormatFrFr">#</a> d3.<b>timeFormatFrFr</b>

[French (France)](https://github.com/d3/d3-time-format/tree/master/src/locale/fr-FR.js)

<a name="timeFormatHeIl" href="#timeFormatHeIl">#</a> d3.<b>timeFormatHeIl</b>

[Hebrew (Israel)](https://github.com/d3/d3-time-format/tree/master/src/locale/he-IL.js)

<a name="timeFormatHuHu" href="#timeFormatHuHu">#</a> d3.<b>timeFormatHuHu</b>

[Hungarian (Hungary)](https://github.com/d3/d3-time-format/tree/master/src/locale/hu-HU.js)

<a name="timeFormatItIt" href="#timeFormatItIt">#</a> d3.<b>timeFormatItIt</b>

[Italian (Italy)](https://github.com/d3/d3-time-format/tree/master/src/locale/it-IT.js)

<a name="timeFormatJaJp" href="#timeFormatJaJp">#</a> d3.<b>timeFormatJaJp</b>

[Japanese (Japan)](https://github.com/d3/d3-time-format/tree/master/src/locale/ja-JP.js)

<a name="timeFormatKoKr" href="#timeFormatKoKr">#</a> d3.<b>timeFormatKoKr</b>

[Korean (South Korea)](https://github.com/d3/d3-time-format/tree/master/src/locale/ko-KR.js)

<a name="timeFormatMkMk" href="#timeFormatMkMk">#</a> d3.<b>timeFormatMkMk</b>

[Macedonian (Macedonia)](https://github.com/d3/d3-time-format/tree/master/src/locale/mk-MK.js)

<a name="timeFormatNlNl" href="#timeFormatNlNl">#</a> d3.<b>timeFormatNlNl</b>

[Dutch (Netherlands)](https://github.com/d3/d3-time-format/tree/master/src/locale/nl-NL.js)

<a name="timeFormatPlPl" href="#timeFormatPlPl">#</a> d3.<b>timeFormatPlPl</b>

[Polish (Poland)](https://github.com/d3/d3-time-format/tree/master/src/locale/pl-PL.js)

<a name="timeFormatPtBr" href="#timeFormatPtBr">#</a> d3.<b>timeFormatPtBr</b>

[Portuguese (Brazil)](https://github.com/d3/d3-time-format/tree/master/src/locale/pt-BR.js)

<a name="timeFormatRuRu" href="#timeFormatRuRu">#</a> d3.<b>timeFormatRuRu</b>

[Russian (Russia)](https://github.com/d3/d3-time-format/tree/master/src/locale/ru-RU.js)

<a name="timeFormatSvSe" href="#timeFormatSvSe">#</a> d3.<b>timeFormatSvSe</b>

[Swedish (Sweden)](https://github.com/d3/d3-time-format/tree/master/src/locale/sv-SE.js)

<a name="timeFormatZhCn" href="#timeFormatZhCn">#</a> d3.<b>timeFormatZhCn</b>

[Chinese (China)](https://github.com/d3/d3-time-format/tree/master/src/locale/zh-CN.js)
