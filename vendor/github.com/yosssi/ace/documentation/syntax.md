# Syntax

## Indent

A unit of an indent must be 2 spaces.

```ace
html
  body
    div
    p
```

becomes

```html
<html>
  <body>
    <div></div>
    <p></p>
  </body>
</html>
```

## HTML Tags

A head word of a line is interpreted as an HTML tag. The rest words of the same line are interpreted as attributes or a text. An attribute value which contains spaces must be surrounded by double quotes. An attribute without value (like "check" and "required") can be defined by specifying no value and ending with an equal (=).

```ace
div id=container style="font-size: 12px; color: blue;"
  p class=row This is interpreted as a text.
  a href=https://github.com/ Go to GitHub
  input type=checkbox checked=
```

becomes

```html
<div id="container" style="font-size: 12px; color: blue;">
  <p class="row">This is interpreted as a text.</p>
  <a href="https://github.com/">Go to GitHub</a>
  <input type="checkbox" checked>
</div>
```

ID and classes can be defined with a head word of a line.

```ace
p#foo.bar
#container
.wrapper
```

becomes

```html
<p id="foo" class="bar"></p>
<div id="container"></div>
<div class="wrapper"></div>
```

Block texts can be defined as a child element of an HTML tag by appending a dot (.) or double dot (..) to the head word of a line. BR tags are inserted to each line except for the last line by appending a double dot (..) to the head word of a line.

```ace
script.
  var msg = 'Hello Ace';
  alert(msg);
p..
  This is a block text.
  BR tags are inserted
  automatically.
```

becomes

```html
<script>
  var msg = 'Hello Ace';
  alert(msg);
</script>
<p>
  This is a block text.<br>
  BR tags are inserted<br>
  automatically.
</p>
```

## Plain Texts

A line which starts with a pipe (|) or double pipe (||) is interpreted as a block of plain texts. BR tags are inserted to each line except for the last line by having a line start with a double pipe (||).

```ace
div
  | This is a single line.
div
  |
    This is a
    block line.
div
  ||
    This is a
    block line
    with BR tags.
```

becomes

```html
<div>
  This is a single line.
</div>
<div>
  This is a
  block line.
</div>
<div>
  This is a<br>
  block line<br>
  with BR tags.
</div>
```

## Helper Methods

A line which starts withs an equal (=) is interpreted as a helper method.

```ace
= helperMethodName
```

The following helper methods are available.

### Conditional Comment Helper Method

A conditional comment helper method generates a [conditional comment](http://en.wikipedia.org/wiki/Conditional_comment).

```ace
= conditionalComment commentType condition
```

The following comment types are acceptable:

| Comment Type | Generated HTML                         |
| ------------ |----------------------------------------|
| hidden       | <!--[if expression]> HTML <![endif]--> |
| revealed     | <![if expression]> HTML <![endif]>     |

```ace
= conditionalComment hidden IE 6
  <p>You are using Internet Explorer 6.</p>
= conditionalComment revealed !IE
  <link href="non-ie.css" rel="stylesheet">
```

becomes

```html
<!--[if IE 6]>
  <p>You are using Internet Explorer 6.</p>
<![endif]-->
<![if !IE]>
  <link href="non-ie.css" rel="stylesheet">
<![endif]>
```

### Content Helper Method

A content helper method defines a block content which is embedded in the base template. This helper method must be used only in the inner template.

```
= content main
  h2 Inner Template - Main : {{.Msg}}

= content sub
  h3 Inner Template - Sub : {{.Msg}}
```

### CSS Helper Method

A css helper method generates a style tag which has "text/css" type.

```ace
= css
  body {
    margin: 0;
  }
  h1 {
    font-size: 200%;
    color: blue;
  }
```

becomes

```html
<style type="text/css">
  body {
    margin: 0;
  }
  h1 {
    font-size: 200%;
    color: blue;
  }
</style>
```

### Doctype Helper Method

A doctype helper method generates a doctype tag.

```ace
= doctype doctypeName
```

The following doctype names are acceptable:

| Doctype Name | Generated HTML |
| ------------ |--------------------------------------|
| html       | <!DOCTYPE html> |
| xml     | <?xml version="1.0" encoding="utf-8" ?> |
| transitional     | <!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd"> |
| strict     | <!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd"> |
| frameset     | <!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Frameset//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-frameset.dtd"> |
| 1.1     | <!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd"> |
| basic     | <!DOCTYPE html PUBLIC "-//W3C//DTD XHTML Basic 1.1//EN" "http://www.w3.org/TR/xhtml-basic/xhtml-basic11.dtd"> |
| mobile     | <!DOCTYPE html PUBLIC "-//WAPFORUM//DTD XHTML Mobile 1.2//EN" "http://www.openmobilealliance.org/tech/DTD/xhtml-mobile12.dtd"> |

```ace
= doctype html
```

becomes

```html
<!DOCTYPE html>
```

### Include Helper Method

An include helper method includes another template. You can pass a pipeline (parameter) from the including template to the included template.

```ace
= include templatePathWithoutExtension pipeline
```

### Javascript Helper Method

A javascript helper method generates a script tag which has "text/javascript" type.

```ace
= javascript
  var msg = 'Hello Ace';
  alert(msg);
```

becomes

```html
<script type="text/javascript">
  var msg = 'Hello Ace';
  alert(msg);
</script>
```

### Yield Helper Method

A yield helper method generates the HTML tags which are defined in the inner template. This helper method must be used only in the base template.

```
= yield main
  | This message is rendered if the "main" content is not defined in the inner template.

= yield sub
  | This message is rendered if the "sub" content is not defined in the inner template.
```

## Comments

A line which starts with a slash (/) or double slash (//) is interpreted as a comment. A line which starts with a slash (/) is not rendered. A line which starts with a double slash (//) is renderd as an HTML comment.

```ace
/ This is a single line comment which is not rendered.
/
  This is a multiple lines comment
  which is not rendered.
// This is a single line comment which is rendered as an HTML comment.
//
  This is a multiple lines comment
  which is rendered as an HTML comment.
```

becomes

```html
<!-- This is a single line comment which is rendered as an HTML comment. -->
<!--
  This is a multiple lines comment
  which is rendered as an HTML comment.
-->
```

## Actions

[Actions](http://golang.org/pkg/text/template/#hdr-Actions) of the template package can be embedded in Ace templates.

```ace
body
  h1 Base Template : {{.Msg}}
  {{if true}}
    p Conditional block
  {{end}}
```

The following functions are predefined.

### HTML function

HTML function returns a non-escaped stirng.

```ace
{{"<div>"}}
{{HTML "<div>"}}
```

becomes

```html
&lt;br&gt;
<br>
```
