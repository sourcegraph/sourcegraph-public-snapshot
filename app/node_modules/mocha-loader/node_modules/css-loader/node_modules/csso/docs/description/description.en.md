CSSO (CSS Optimizer) is a CSS minimizer unlike others. In addition to usual minification techniques it can perform structural optimization of CSS files, resulting in smaller file size compared to other minifiers.

## Minification

Minification is a process of transforming a CSS document into a smaller document without losses. The typical strategies of achieving this are:

* basic transformations, such as removal of unnecessary elements (e.g. trailing semicolons) or transforming the values into more compact representations (e.g. `0px` to `0`);
* structural optimizations, such as removal of overridden properties or merging of blocks.

### Basic transformations

#### Removal of whitespace

In some cases, whitespace characters (` `, `\n`, `\r`, `\t`, `\f`) are unnecessary and do not affect rendering.

* Before:
```css
        .test
        {
            margin-top: 1em;

            margin-left  : 2em;
        }
```

* After:
```css
        .test{margin-top:1em;margin-left:2em}
```

The following examples are provided with whitespace left intact for better readability.

#### Removal of trailing ';'

The last semicolon in a block is not required and does not affect rendering.

* Before:
```css
        .test {
            margin-top: 1em;;
        }
```

* After:
```css
        .test {
            margin-top: 1em
        }
```

#### Removal of comments

Comments do not affect rendering: \[[CSS 2.1 / 4.1.9 Comments](http://www.w3.org/TR/CSS21/syndata.html#comments)\].

* Before:
```css
        /* comment */

        .test /* comment */ {
            /* comment */ margin-top: /* comment */ 1em;
        }
```

* After:
```css
        .test {
            margin-top: 1em
        }
```

If you want to save the comment, CSSO can do it with only one first comment in case it starts with `!`.

* Before:
```css
        /*! MIT license */
        /*! will be removed */

        .test {
            color: red
        }
```

* After:
```css
        /*! MIT license */
        .test {
            color: red
        }
```

#### Removal of invalid @charset and @import declarations

According to the specification, `@charset` must be placed at the very beginning of the stylesheet: \[[CSS 2.1 / 4.4 CSS style sheet representation](http://www.w3.org/TR/CSS21/syndata.html#charset)\].

CSSO handles this rule in a slightly relaxed manner - we keep the `@charset` rule which immediately follows whitespace and comments in the beginning of the stylesheet.

Incorrectly placed `@import` rules are deleted according to \[[CSS 2.1 / 6.3 The @import rule](http://www.w3.org/TR/CSS21/cascade.html#at-import)\].

* Before:
```css
        /* comment */
        @charset 'UTF-8';
        @import "test0.css";
        @import "test1.css";
        @charset 'wrong';

        h1 {
            color: red
        }

        @import "wrong";
```

* After:
```css
        @charset 'UTF-8';
        @import "test0.css";
        @import "test1.css";
        h1 {
            color: red
        }
```

#### Minification of color properties

Some color values are minimized according to \[[CSS 2.1 / 4.3.6 Colors](http://www.w3.org/TR/CSS21/syndata.html#color-units)\].

* Before:
```css
        .test {
            color: yellow;
            border-color: #c0c0c0;
            background: #ffffff;
            border-top-color: #f00;
            outline-color: rgb(0, 0, 0);
        }
```

* After:
```css
        .test {
            color: #ff0;
            border-color: silver;
            background: #fff;
            border-top-color: red;
            outline-color: #000
        }
```

#### Minification of 0

In some cases, the numeric values can be compacted to `0` or even dropped.

The `0%` value is not being compacted to avoid the following situation: `rgb(100%, 100%, 0)`.

* Before:
```css
        .test {
            fakeprop: .0 0. 0.0 000 00.00 0px 0.1 0.1em 0.000em 00% 00.00% 010.00
        }
```

* After:
```css
        .test {
            fakeprop: 0 0 0 0 0 0 .1 .1em 0 0% 0% 10
        }
```

#### Minification of multi-line strings

Multi-line strings are minified according to \[[CSS 2.1 / 4.3.7 Strings](http://www.w3.org/TR/CSS21/syndata.html#strings)\].

* Before:
```css
        .test[title="abc\
        def"] {
            background: url("foo/\
        bar")
        }
```

* After:
```css
        .test[title="abcdef"] {
            background: url("foo/bar")
        }
```

#### Minification of the font-weight property

The `bold` and `normal` values of the `font-weight` property are minimized according to \[[CSS 2.1 / 15.6 Font boldness: the 'font-weight' property](http://www.w3.org/TR/CSS21/fonts.html#font-boldness)\].

* Before:
```css
        .test0 {
            font-weight: bold
        }

        .test1 {
            font-weight: normal
        }
```

* After:
```css
        .test0 {
            font-weight: 700
        }

        .test1 {
            font-weight: 400
        }
```

### Structural optimizations

#### Merging blocks with identical selectors

Adjacent blocks with identical selectors are merged.

* Before:
```css
        .test0 {
            margin: 0
        }

        .test1 {
            border: none
        }

        .test1 {
            background-color: green
        }

        .test0 {
            padding: 0
        }
```

* After:
```css
        .test0 {
            margin: 0
        }

        .test1 {
            border: none;
            background-color: green
        }

        .test0 {
            padding: 0
        }
```

#### Merging blocks with identical properties

Adjacent blocks with identical properties are merged.

* Before:
```css
        .test0 {
            margin: 0
        }

        .test1 {
            border: none
        }

        .test2 {
            border: none
        }

        .test0 {
            padding: 0
        }
```

* After:
```css
        .test0 {
            margin: 0
        }

        .test1, .test2 {
            border: none
        }

        .test0 {
            padding: 0
        }
```

#### Removal of overridden properties

Properties ignored by the browser can be removed using the following rules:

* the last property in a CSS rule is applied, if none of the properties have an `!important` declaration;
* among `!important` properties, the last one is applied.

* Before:
```css
        .test {
            color: red;
            margin: 0;
            line-height: 3cm;
            color: green;
        }
```

* After:
```css
        .test {
            margin: 0;
            line-height: 3cm;
            color: green
        }
```

##### Removal of overridden shorthand properties

In case of `border`, `margin`, `padding`, `font` and `list-style` properties, the following removal rule will be applied: if the last property is a 'general' one (for example, `border`), then all preceding overridden properties will be removed (for example, `border-top-width` or `border-style`).

* Before:
```css
        .test {
            border-top-color: red;
            border-color: green
        }
```

* After:
```css
        .test {
            border-color:green
        }
```

#### Removal of repeating selectors

Repeating selectors can be removed.

* Before:
```css
        .test, .test {
            color: red
        }
```

* After:
```css
        .test {
            color: red
        }
```

#### Partial merging of blocks

Given two adjacent blocks where one of the blocks is a subset of the other one, the following optimization is possible:

* overlapping properties are removed from the source block;
* the remaining properties of the source block are copied into a receiving block.

Minification will take place if character count of the properties to be copied is smaller than character count of the overlapping properties.

* Before:
```css
        .test0 {
            color: red
        }

        .test1 {
            color: red;
            border: none
        }

        .test2 {
            border: none
        }
```

* After:
```css
        .test0, .test1 {
            color: red
        }

        .test1, .test2 {
            border: none
        }
```

Minification won't take place if character count of the properties to be copied is larger than character count of the overlapping properties.

* Before:
```css
        .test0 {
            color: red
        }

        .longlonglong {
            color: red;
            border: none
        }

        .test1 {
            border: none
        }
```

* After:
```css
        .test0 {
            color: red
        }

        .longlonglong {
            color: red;
            border: none
        }

        .test1 {
            border: none
        }
```

#### Partial splitting of blocks

If two adjacent blocks contain intersecting properties the following minification is possible:

* property intersection is determined;
* a new block containing the intersection is created in between the two blocks.

Minification will take place if there's a gain in character count.

* Before:
```css
        .test0 {
            color: red;
            border: none;
            margin: 0
        }

        .test1 {
            color: green;
            border: none;
            margin: 0
        }
```

* After:
```css
        .test0 {
            color: red
        }

        .test0, .test1 {
            border: none;
            margin: 0
        }

        .test1 {
            color: green
        }
```

Minification won't take place if there's no gain in character count.

* Before:
```css
        .test0 {
            color: red;
            border: none;
            margin: 0
        }

        .longlonglong {
            color: green;
            border: none;
            margin: 0
        }
```

* After:
```css
        .test0 {
            color: red;
            border: none;
            margin: 0
        }

        .longlonglong {
            color: green;
            border: none;
            margin: 0
        }
```

#### Removal of empty ruleset and at-rule

Empty ruleset and at-rule will be removed.

* Before:
```css
        .test {
            color: red
        }

        .empty {}

        @font-face {}

        @media print {
            .empty {}
        }

        .test {
            border: none
        }
```

* After:
```css
        .test{color:red;border:none}
```

#### Minification of margin and padding properties

The `margin` and `padding` properties are minimized according to \[[CSS 2.1 / 8.3 Margin properties](http://www.w3.org/TR/CSS21/box.html#margin-properties)\] and \[[CSS 2.1 / 8.4 Padding properties](http://www.w3.org/TR/CSS21/box.html#padding-properties)\].

* Before:
```css
        .test0 {
            margin-top: 1em;
            margin-right: 2em;
            margin-bottom: 3em;
            margin-left: 4em;
        }

        .test1 {
            margin: 1 2 3 2
        }

        .test2 {
            margin: 1 2 1 2
        }

        .test3 {
            margin: 1 1 1 1
        }

        .test4 {
            margin: 1 1 1
        }

        .test5 {
            margin: 1 1
        }
```

* After:
```css
        .test0 {
            margin: 1em 2em 3em 4em
        }

        .test1 {
            margin: 1 2 3
        }

        .test2 {
            margin: 1 2
        }

        .test3, .test4, .test5 {
            margin: 1
        }
```

## Recommendations

Some stylesheets compress better than the others. Sometimes, one character difference can turn a well-compressible stylesheet to a very inconvenient one.

You can help the minimizer by following these recommendations.

### Length of selectors

Shorter selectors are easier to re-group.

### Order of properties

Stick to the same order of properties throughout the stylesheet - it will allow you to not use the guards. The less manual intervention there is, the easier it is for the minimizer to work optimally.

### Positioning of similar blocks

Keep blocks with similar sets of properties close to each other.

Bad:

* Before:
```css
        .test0 {
            color: red
        }

        .test1 {
            color: green
        }

        .test2 {
            color: red
        }
```

* After (53 characters):
```css
        .test0{color:red}.test1{color:green}.test2{color:red}
```

Good:

* Before:
```css
        .test1 {
            color: green
        }

        .test0 {
            color: red
        }

        .test2 {
            color: red
        }
```

* After (43 characters):
```css
        .test1{color:green}.test0,.test2{color:red}
```

### Using !important

It should go without saying that using the `!important` declaration harms minification performance.

Bad:

* Before:
```css
        .test {
            margin-left: 2px !important;
            margin: 1px;
        }
```

* After (43 characters):
```css
        .test{margin-left:2px!important;margin:1px}
```

Good:

* Before:
```css
        .test {
            margin-left: 2px;
            margin: 1px;
        }
```

* After (17 characters):
```css
        .test{margin:1px}
```
