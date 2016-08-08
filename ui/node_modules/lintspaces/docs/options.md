## Options

### newline at end of file option

Tests for newlines at the end of all files. Default value is `false`.

```javascript
	newline: true
```

* returns code ```NEWLINE```, when a missing a newline at the end of the file.
* returns code ```NEWLINE_AMOUNT```, when found unexpected additional newlines
at the end of a file.
* returns type ```warning```

### maximum newlines option

Test for the maximum amount of newlines between code blocks. Default value is
`false`. To enable this validation a number larger than `0` is expected.

```javascript
	newlineMaximum: 2
```

* returns code ```NEWLINE_MAXIMUM```, when maximum amount of newlines exceeded
between code blocks.
* returns type ```warning```

### trailingspaces option

Tests for useless whitespaces (trailing whitespaces) at each lineending of all
files. Default value is `false`.

```javascript
	trailingspaces: true
```

* returns code ```TRAILINGSPACES```, when unexpected trailing spaces were found.
* returns type ```warning```

**Note:** If you like to to skip empty lines from reporting (for whatever
reason), use the option ```trailingspacesSkipBlanks``` and set them to ```true```.

### indentation options

Tests for correct indentation using tabs or spaces. Default value is `false`.
To enable indentation check use the value `'tabs'` or `'spaces'`.

```javascript
	indentation: 'tabs'
```

* returns code ```INDENTATION_TABS```, when spaces are used instead of tabs.
* returns type ```warning```

If the indentation option is set to `'spaces'`, there is also the possibility
to set the amount of spaces per indentation using the `spaces` option. Default
value is `4`.

```javascript
	indentation: 'spaces',
	spaces: 2
```

* returns code ```INDENTATION_SPACES```, when tabs are used instead of spaces.
* returns code ```INDENTATION_SPACES_AMOUNT```, when spaces are used but the
amound is not as expected.
* returns type ```warning```

### guess indentation option

This ```indentationGuess``` option _tries to guess_ the indention of a line
depending on previous lines. The report of this option can be incorrect,
because the _correct_ indentation depends on the actual programming language
and styleguide of the certain file. The default value is `false` - disabled.

This feature follows the following rules: _The indentation of the current
line is correct when:_

* the amount of indentations is equal to the previous or
* the amount of indentations is less than the previous line or
* the amount of indentations is one more than the previous line
* the amount of indentations is zero and the lines length is also zero which
is an empty line without trailing whitespaces

```javascript
	indentationGuess: true
```

* returns code ```NEWLINE_GUESS```
* returns type ```hint```

### allowsBOM option

Lintspaces fails with incorrect indentation errors when files contain Byte Order
Marks (BOM). If you don't want to give false positives for inconsistent tabs or
spaces, set the ```allowsBOM``` option to ```true```.  The default value is
`false` - disabled.

```javascript
	allowsBOM: true
```

### ignores option

Use the `ignores` option when special lines such as comments should be ignored.
Provide an array of regular expressions to the `ignores` property.

```javascript
	ignores: [
		/\/\*[\s\S]*?\*\//g,
		/foo bar/g
	]
```

There are some _**build in**_ ignores for comments which you can apply by using
these strings:

* 'js-comments'
* 'c-comments'
* 'java-comments'
* 'as-comments'
* 'xml-comments'
* 'html-comments'
* 'python-comments'
* 'ruby-comments'
* 'applescript-comments'

_(build in strings and userdefined regular expressions are mixable in the
`ignores` array)_

```javascript
	ignores: [
		'js-comments',
		/foo bar/g
	]
```

_Feel free to contribute some new regular expressions as build in!_

**Note:** Trailing spaces are not ignored by default, because they are always
evil!! If you still want to ignore them use the ```trailingspacesToIgnores```
option and set them to ```true```.

### .editorconfig option

It's possible to overwrite the default and given options by setting up a path
to an external editorconfig file by using the `editorconfig` option. For a basic
configuration of a _.editorconfig_ file check out the
[EditorConfig Documentation](http://editorconfig.org/).

```javascript
	editorconfig: '.editorconfig'
```

The following .editorconfig values are supported:

* `insert_final_newline` will check if a newline is set
* `indent_style` will check the indentation
* `indent_size` will check the amount of spaces
* `trim_trailing_whitespace` will check for useless whitespaces
