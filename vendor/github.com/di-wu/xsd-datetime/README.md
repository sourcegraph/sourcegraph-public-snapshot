# xsd:dateTime
[![](https://godoc.org/github.com/di-wu/xsd-datetime?status.svg)](http://godoc.org/github.com/di-wu/xsd-datetime)
## Description
**xsd:dateTime** describes instances identified by the combination of a date and a time. 
Its value space is described as a combination of date and time of day in *Chapter 5.4 of ISO 8601*. 

Its lexical space is the extended format:
```
[-]YYYY-MM-DDThh:mm:ss[.fffffffff][zzzzzz]
```

The time zone may be specified as `Z` (UTC) or `(+|-)hh:mm`.
Time zones that are not specified are considered undetermined.

The literal Z, which represents the time in UTC (Z represents Zulu time, which is equivalent to UTC). 
Specifying Z for the time zone is equivalent to specifying +00:00 or -00:00.

## Lexical Form
**YYYY**
A four-digit numeral that represents the year.
The value cannot begin with a plus (+) sign. \
**-**
Separators between parts of the date portion. \
**MM**
A two-digit numeral that represents the month. \
**DD**
A two-digit numeral that represents the day. \
**T**
A separator to indicate that the time of day follows. \
**hh**
A two-digit numeral that represents the hour. \
**mm**
A two-digit numeral that represents the minute. \
**ss**
A two-digit numeral that represents the whole seconds. \
**.fffffffff**
Optional. If present, a 1-to-9 digit numeral that represents the fractional seconds. \
**zzzzzz**
Optional. If present, represents the time zone. If a time zone is not specified the dateTime has no timezone.
However, an implicit time zone of UTC is used for comparison and arithmetic operations.

### Timezone Indicator
**hh**
A two-digit numeral (with leading zeros as required) that represents the hours. 
The value must be between -14 and +14, inclusive. \
**mm**
A two-digit numeral that represents the minutes. 
The value of the minutes property must be zero when the hours property is equal to 14.\
**+**
Indicates that the specified time instant is in a time zone that is ahead of the UTC time by hh hours and mm minutes. \
**-**
Indicates that the specified time instant is in a time zone that is behind UTC time by hh hours and mm minutes.

## Regular Expression
*(once whitespace is removed)*
```regexp
-?([1-9][0-9]{3,}|0[0-9]{3})
-(0[1-9]|1[0-2])
-(0[1-9]|[12][0-9]|3[01])
T(([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9](\.[0-9]+)?|(24:00:00(\.0+)?))
(Z|(\+|-)((0[0-9]|1[0-3]):[0-5][0-9]|14:00))?
```

## Restrictions
- Each part of the datetime value that is expressed as a numeric value is constrained to the maximum value within 
the interval that is determined by the next-higher part of the datetime value.
- There is no support for any calendar system other than Gregorian.
- There is no support for any localization such as different orders for date parts or named months.
- The basic format of *ISO 8601* calendar datetime, `CCYYMMDDThhmmss`, is not supported.
- The other forms of date-times available in *ISO 8601* are not supported. \
*(ordinal dates defined by the year, the number of the day in the year, dates identified by calendar week and day numbers)*

### Sources
[Relax NG](http://www.oreilly.com/catalog/relax/) (book), 
[W3](https://www.w3.org/TR/xmlschema11-2/#dateTime) and
[ISO 8601](https://www.iso.org/obp/ui#iso:std:iso:8601:-1:ed-1:v1:en)
