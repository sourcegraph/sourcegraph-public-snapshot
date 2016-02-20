% Title = "Using mmark to create I-Ds and RFCs"
% abbrev = "mmark2rfc"
% category = "info"
% docName = "draft-gieben-mmark2rfc-00"
% ipr= "trust200902"
% area = "Internet"
% workgroup = ""
% keyword = ["markdown", "xml", "mmark"]
%
% date = 2014-12-10T00:00:00Z
%
% [[author]]
% initials="R."
% surname="Gieben"
% fullname="R. (Miek) Gieben"
% #role="editor"
% organization = "Google"
%   [author.address]
%   email = "miek@google.com"
%   [author.address.postal]
%   street = "Buckingham Palace Road"

.# Abstract

This document describes an markdown variant called mmark [@?mmark] that can
be used to create RFC documents. The aim of mmark is to make writing document
as natural as possible, while providing a lot of power on how to structure and layout
the document.

The
[source of this document](https://raw.githubusercontent.com/miekg/mmark/master/mmark2rfc.md)
provides a good example.

{mainmatter}

# Introduction

Mmark [@mmark] is a markdown processor. It supports the markdown syntax
and has been extended with (syntax) features found in other markdown
implementations like [kramdown], [PHP markdown extra], [@pandoc],
[Scholarly markdown], [leanpub] and even [asciidoc]. This allows mmark to be used
to write larger, structured documents such as RFC and I-Ds or even books, while
not deviating too far from markdown.

Mmark is a fork of blackfriday [@blackfriday], is written in Golang and very fast.
Input to mmark must be UTF-8, the output is also UTF-8. Mmark converts tabs to 4 spaces.

The goals of mmark are:

{style="format (%I)"}
1. Self contained: a single file can be converted to XML2RFC v2 or (v3) or HTML5.
2. Make the markdown "source code" look as natural as possible.
3. Provide seemless upgrade path to XML2RFC v3.
4. Consistent interface, aim to minimize the number of weird corner cases you need
    to remember while typing.

Using Figure 1 from [@!RFC7328], mmark can be positioned as follows:

{#fig:mmark align=left callout="true"}
     +-------------------+   pandoc   +---------+
     | ALMOST PLAIN TEXT |   ------>  | DOCBOOK | <2>
     +-------------------+            +---------+
                   |      \                 |
     non-existent  |       \_________       | xsltproc
       faster way  |    <1> *mmark*  \      |
                   v                  v     v
           +------------+    xml2rfc  +---------+
           | PLAIN TEXT |  <--------  |   XML   | <3>
           +------------+             +---------+
Figure: Mmark <1> skips the conversion to DOCBOOK <1> and directly outputs XML2RFC XML <3> (or HTML5).

Note that [kramdown-2629](https://github.com/cabo/kramdown-rfc2629) fills the same niche as mmark.

# Terminology

The folloing terms are used in this document:

v2:
:   Refers to XML2RFC version 2 [@!RFC2926] output created by mmark.

v3:
:   Refers  to XML2RFC version 2 [@!I-D.hoffman-xml2rfc#15] output created by mmark.


# Mmark Syntax

In the following sections we go over some of the differences, and the extra syntax features of mmark.

Note that there are no wrong markdown documents, but once converted to XML may lead to
an invalid doc, case in point: having a table in a list and converting to v2.

# TOML header

Mmark uses TOML [@!toml] document header to specify the document's meta data. Each line of this
header must start with an `% `. The document header is also different in v3, for instance the
`docName` is not used anymore.

# Citations

A citation can be entered using the syntax from @pandoc: `[@reference]`,
such a reference is "informative" by default. Making a reference informative or normative
can be done with a `?` and `!` respectively: `[@!reference]` is a normative reference.

For RFC and I-Ds the references are generated automatically, meaning you don't need to include
an XML reference element in source of document.

For I-Ds you might need to include a draft version in the reference
`[@?I-D.blah#06]`, creates an informative reference to the seventh version of
draft-blah.

Once a citation has been defined the brackets can be omited, so once `[@pandoc]` is used, you
can just use `@pandoc`.

If the need arises (usually when citing a document that is not in the XML2RFC database)
an XML reference fragment should be included, note that this needs to happen
*before* the back matter is started, because that is the point when the references are outputted
(right now the implementation does not scan the entire file for citations, also see (#bugs)).

# Internal References

The cross reference syntax is `[](#id)`, which allows for an optional title between the brackets.
Usually this is left empty, for this use case mmark allows the shortcut form `(#id)` which omits
the brackets in its entirely.

The external reference syntax is `[](url)`.

# Document divisions

Using `{mainmatter}` on a line by itself starts the main matter (middle) of the document, `{backmatter}`
starts the appendix. There is also a `{frontmatter}` that starts the front matter (front) of the document,
but is normally not needed because the TOML header ([](#toml-header)) starts that by default.

# Abstract

An abstract is defined by using the special header syntax `.#`. The name of the section, when lowercased,
must be "abstract".
In the future mmark might also support Preface and Colophon (special) sections.

# Captions

Whenever an blockquote, fenced codeblock or image has caption text, the entire block is wrapped
in a `<figure>` and the caption text is put in a `<name>` tag for v3.

In mmark you can put a
caption under either a table, indented code block (even after a fenced code block) or even after a block quote.
Referencing these elements (and thus
creating an document `id` for them), is done with an IAL ([](#inline-attribute-lists)):

    {#identifier}
    Name    | Age
    --------|-----:
    Bob     | 27
    Alice   | 23

An empty line between the IAL and the table or indented code block is allowed.

## Tables

A table caption is signalled by using `Table: ` directly after the table.


## Figures

Any text directly after the code block/fenced code block starting with `Figure: ` is used as the caption.

## Quotes

After a quote (a paragraph prefixed with `> `) you can add a caption:

    Quote: Name -- URI for attribution

In v3 this is used in the block quote attributes, for v2 it is discarded. If you need
the string `Quote: ` after an quote, escape the colon: `Quote\: `.

# Tables

Tables can be created by drawing them in the input using a simple syntax:

```
Name    | Age
--------|-----:
Bob     | 27
Alice   | 23
```

Tables can also have a footer: use equal signs instead of dashes for the separator,
to start a table footer. If there are multiple footer lines, the first one is used as a
starting point for the table footer.

```
Name    | Age
--------|-----:
Bob     | 27
Alice   | 23
======= | ====
Charlie | 4
```

If a table is started with a *block table header*, which starts
with a pipe or plus sign and a minimum of three dashes,
it is a **Block Table**. A block table may include block level elements in each
(body) cell. If we want to start a new cell use the block table header
syntax. In the example below we include a list in one of the cells.

```
|-----------------+------------+-----------------|
| Default aligned |Left aligned| Center aligned  |
|-----------------|:-----------|:---------------:|
| Bob             |Second cell | Third cell      |
| Alice           |foo         | **strong**      |
| Charlie         |quux        | baz             |
|-----------------+------------+-----------------|
| Bob             | foot       | 1. Item2        |
| Alice           | quuz       | 2. Item2        |
|=================+============+=================|
| Footer row      | more footer| and more        |
|-----------------+------------+-----------------|
```

Note that the header and footer can't contain block level elements.
The table syntax used that one of
[Markdown Extra](https://michelf.ca/projects/php-markdown/extra/#table).

# Inline Attribute Lists

This borrows from [kramdown](http://kramdown.gettalong.org/syntax.html#block-ials), with
the difference that the colon is dropped and each IAL must be typeset *before* the block element
(see (#bugs)).
Added an anchor to blockquote can be done like so:

    {#quote:ref1}
    > A block quote

You can specify classes with `.class` (although these are not used when converting to XML2RFC), and
arbitrary key value pairs where each key/value becomes an attribute. Different elements in the IAL
must be seperated using spaces: `{#id lang=go}`.

For the following elements a IAL is processed:

* Table
* Code Block
* Fenced Code Block
* List (any type)
* Section Header
* Image
* Quote
* ...

For all other elements they are ignored, but not disgarded. This means they will be applied to the
next element that does use the IALs.

# Lists

## Ordered Lists

The are several ways to start an ordered lists. You can use numbers, roman numbers, letters and uppercase
letters. When using roman numbers and letter you **MUST** use two spaces after the dot or the brace (the
underscore signals a space here):

    a)__
    A)__

Note that mmark (just as @pandoc) pays attention to the starting number of a list (when using decimal numbers), thus
a list started with:

    4) Item4
    5) Item5

Will use for `4` as the starting number.

## Unordered Lists

Unordered lists can be started with `*`, `+` or `-` and follow the normal markdown syntax rules. <!-- * -->

## Definition Lists

Mmark supports the definition list syntax from [PHP Markdown
Extra](https://michelf.ca/projects/php-markdown/extra/#def-list), meaning there
can not be a empty line between the term and the definition. Note the multiple
terms and definition syntax is *not* supported.

## Example Lists

This is the example list syntax
[from pandoc](http://johnmacfarlane.net/pandoc/README.html#extension-example_lists). References
to example lists work as well. Note that an example list always needs to have an identifier,
`(@good)` works, `(@)` does not.

Example:

    @good)  This is a good example.

    As (@good) illustrates, ...

# Figures and Images

When an figure has a caption it will be wrapped in `<figure`> tags. A figure can
wrap source code (v3) or artwork (v2/v3).

An image is wrapped in a figure when the optional title syntax is used. But images
are only useful when outputting v3. For v2 the actual image can not be shown, see
(#images-in-v2) for this.

Multiple artworks/sources can be put in one figure. This done by prefixing the
section containing the figures with a figure quote: `F> `.

## Details

*   A Fenced Code Block will becomes a source code in v3 and artwork in v2.
    We can use the language to signal the type.

        ``` c
        printf("%s\n", "hello");
        ```

*   An Indented Code Block becomes artwork in v3 and artwork in v2. The only way
    to indicate the type is by using an IAL. So one has to use:

        {type="ascii-art"}
            +-----+
            | ART |
            +-----+

    v3 allows the usage of a `src` attribute to link to external files with images.
    We use the image syntax for that.

*   An image `![Alt text](/path/to/img.jpg "Optional title")`, will be converted
    to an artwork with a `src` attribute in v3. Again the type needs to be specified
    as an IAL.

    If the "Optional title" is specified the generated artwork will be wrapped in a
    figure with name set to "Optional title"

    Creating an artwork with an anchor and type will become:

        {#fig-id type="ascii-art"}
        ![](/path/to/art.txt "Optional title")

    For v2 this presents difficulties as there is no way to display any of this, see
    (#images-in-v2) for a treatment on how to deal with that.

*   To group artworks and code blocks into figures, we need an extra syntax element.
    [Scholarly markdown] has a neat syntax
    for this. It uses a special section syntax and all images in that section become
    subfigures of a larger figure. Disadvantage of this syntax is that it can not be
    used in lists. Hence we use a quote like solution, just like asides and notes,
    but for figures: we prefix the entire paragraph with `F>` .

    Basic usage:

        F>  {type="ascii-art"}
        F>      +-----+
        F>      | ART |
        F>      +-----+
        F>  Figure: This caption is ignored in v3, but used in v2.
        F>
        F>  ``` c
        F>  printf("%s\n", "hello");
        F>  ```
        F>
        Figure: Caption for both figures in v3 (in v2 this is ignored).

    In v2 this is not supported so the above will result in one figure. Yes one, because
    the fenced code block does not have a caption, so it will not be wrapped in a figure.

    To summerize in v2 the inner captions *are* used and the outer one is discarded, for v3 it
    is the other way around.

    The figure from above will be rendered as:

    F> {type="ascii-art"}
    F>      +-----+
    F>      | ART |
    F>      +-----+
    F>  Figure: This caption is ignored in v3, but used in v2.
    F>
    F>  ``` c
    F>  printf("%s\n", "hello");
    F>  ```
    F>
    Figure: Caption for both figures in v3 (in v2 it's ignored).


## Images in v2

Images (real images, not ascii-art) are non-existent in v2, but are allowed in v3. To allow
writers to use images *and* output v2 and v3 formats, the following hack is used in v2 output.
Any image will be converted to a figure with an title attribute set to the "Optional title".
And the url in the image will be type set as a link in the postamble.
So `![](/path/to/art.txt "Optional title")` will be converted to:

    <figure title="Optional title">
     <artwork>
     </artwork>
      <postamble>
       <eref target="/path/to/art.txt"/>
      </postamble>
    </figure>

If a image does not have a title, the `figure` is dropped and only the link remains. The default
is to center the entire element. Note that is you don't give the image an anchor, `xml2rfc` won't
typeset it with a `Figure X`, so for an optional "image" rendering, you should use the folowing:

    {#fig-id}
    ![](/path/to/art.txt "Optional title")

Which when rendered becomes:

{#fig-id}
![](/path/to/art.txt "Optional title")

Note that ideas to improve/change on this are welcome.

# Miscellaneous Features

## HTML Comment

If a HTML comment contains `--`, it will be rendered as a `cref` comment in the resulting
XML file. Typically `<!-- Miek Gieben -- you want to include the next paragraph? -->`.

## Including Files

Files can be included using `{{filename}}`, `filename` is relative to the current working
directory if it is not absolute.

## Including Code Fragments

This borrows from the Go present tool, which got its inspiration from the Sam editor. The syntax was gleaned from leanpub.
But the syntax presented here is more powerful than the one used by leanpub.
Use the
syntax: `<{{file}}[address]` to include a code snippet. The `address` identifier specifies
what lines of code are to be included in the fragment.

Any line in the program that ends with the four characters `OMIT`
is deleted from the source before inclusion, making it easy to write things like

    <{{test.go}}[/START OMIT/,/END OMIT/]

So you can include snippets like this:
~~~
tedious_code = boring_function()
// START OMIT
interesting_code = fascinating_function()
// END OMIT
~~~

To aid in including HTML or XML framents, where the `OMIT` key words is probably embedded in
comments, line the in in `OMIT -->` are excluded as well.
Note that the default is put out an artwork, but if the extension of the included file matches
a computer language, `<sourcecode>` will be emitted for v3.

Note that the attribute `prefix` (which you can specify with an IAL) can be used to prefix
all lines of the code to be included to prefixed with the value of the attribute, so

```
{prefix="C:"}
    <{{test.go}}[/START OMIT/,/END OMIT/]
```

Will prefix all lines of test.go with 'C:' when included.

# XML2RFC V3 features

The v3 syntax adds some new features and those can already be used in mmark (even for documents targeting
v2 -- but there they will be faked with the limited constructs of the v2 syntax).

## Asides

Any paragraph prefixed with `A> `. For v2 this becomes a indented paragraph.

## Notes

Any paragraph prefixed with `N> `. For v2 this becomes a indented paragraph.

## RFC 2119 Keywords

Any [@?RFC2119] keyword used with strong emphasis *and* in uppercase  will be typeset
within `bcp14` tags, that is `**MUST**` becomes `<bcp14>MUST</bcp14>`, but `**must**` will not.
For v2 they are stripped of the emphasis and outputted as-is.

## Super- and Subscripts

Use H~2~O and 2^10^ is 1024. In v2 these are outputted as-is.

# Converting from RFC 7328 Syntax

Converting from an RFC 7328 ([@!RFC7328]) document can be done using the quick
and dirty [Perl script](https://raw.githubusercontent.com/miekg/mmark/master/convert/parts.pl),
which uses pandoc to output markdown PHP extra and converts that into proper mmark:
(mmark is more like markdown PHP extra, than like pandoc).

    for i in middle.mkd back.mkd; do \
        pandoc --atx-headers -t markdown_phpextra < $i |\
        ./parts.pl
    done

Note this:

* Does not convert the abstract to a prefixed paragraph;
* Makes all RFC references normative;
* Handles all figure and table captions and adds references (if appropriate);
* Probably has other bugs, so a manual review should be in order.

There is also [titleblock.pl](https://raw.githubusercontent.com/miekg/mmark/master/convert/titleblock.pl)
which can be given an @RFC7328 `template.xml` file and will output a TOML titleblock, that can
be used as a starting point.

A> Yes, this uses pandoc and Perl.. why? Becasue if mmark could parse the file by itself, there wasn't much
A> of problem. Two things are holding this back: mmark cannot parse definition lists with empty spaces and
A> there isn't renderer that can output markdown syntax.

For now the mmark parser will not get any features that makes it backwards compatible with pandoc2rfc.

# Acknowledgements

<!-- reference we need to include -->

<reference anchor='mmark' target='http://github.com/miekg/mmark'>
    <front>
        <title>Mmark git repository</title>
        <author initials='R.' surname='Gieben' fullname='R. (Miek) Gieben'>
            <address>
                <email>miek@miek.nl</email>
            </address>
        </author>
        <date year='2014' month='December'/>
    </front>
</reference>

<reference anchor='blackfriday' target='http://github.com/russross/blackfriday'>
    <front>
        <title>Blackfriday git repository</title>
        <author initials='' surname='' fullname=''>
            <address>
                <email>miek@miek.nl</email>
            </address>
        </author>
        <date year='2011' month='November'/>
    </front>
</reference>

<reference anchor='toml' target='https://github.com/toml-lang/toml'>
    <front>
        <title>TOML git repository</title>
        <author initials='T.' surname='Preston-Werner' fullname='Tom Preston-Werner'>
            <address>
                <email></email>
                </address>
            </author>
        <date year='2013' month='March' />
    </front>
</reference>

<reference anchor='pandoc' target='http://johnmacfarlane.net/pandoc/'>
    <front>
        <title>Pandoc, a universal document converter</title>
        <author initials='J.' surname='MacFarlane' fullname='John MacFarlane'>
            <organization>University of California, Berkeley</organization>
            <address>
                <email>jgm@berkeley.edu</email>
                <uri>http://johnmacfarlane.net/</uri>
            </address>
        </author>
        <date year='2006' />
    </front>
</reference>

{backmatter}

# Tips and Tricks

How do I type set:

Multiple paragraphs in a list:
:   Indent the list with four spaces. Text indented with three spaces will
    be seen as a new paragraph that breaks the list.

# Bugs

*  Citations must be included in the text before the `{backmatter}` starts.
   otherwise they are not available in the appendix.
*  Inline Attribute Lists must be given *before* the block element.
*  Mmark cannot correctly parse @RFC728 markdown.
*  Multiple terms and definitions are not supported in definition lists.
*  Mmark uses two scans when converting a document and does not build an
   internal AST of the document, this means it can not adhere 100% to the
   [CommonMark] specification, however the CommonMark test suite is used when
   developing mmark. Currently mmark passes ~60% of the tests.

# Changes

## xx

* Abstract are designated using a special header `.# Abstract`
* Removed exercises and answers, this needs a better syntax.
* Add math `$$`.

[kramdown]: http://http://kramdown.gettalong.org/
[leanpub]: https://leanpub.com/help/manual
[asciidoc]: http://www.methods.co.nz/asciidoc/
[PHP markdown extra]: http://michelf.com/projects/php-markdown/extra/
[pandoc]: http://johnmacfarlane.net/pandoc/
[CommonMark]: http://commonmark.org/
[Scholarly markdown]: http://scholarlymarkdown.com/Scholarly-Markdown-Guide.html

