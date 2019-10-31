# Documentation style guide

> NOTE: Adapted from [GitLab documentation guidelines](https://gitlab.com/gitlab-org/gitlab-ee/raw/master/doc/development/documentation/index.md).

The documentation style guide defines the markup structure used in Sourcegraph documentation. Check the [documentation guidelines](index.md) for general development instructions.

See the [Sourcegraph style guide](../style_guide.md) for general information.

For help adhering to the guidelines, see [Linting](index.md#linting).

## Files

- [Directory structure](index.md#location-and-naming-documents): place the docs in the correct location.
- [Documentation files](index.md#documentation-files): name the files accordingly.
<!-- - [Markdown](../../user/markdown.md): TODO(ryan): Fix in 3.3 once it has content. -->

- Don't use capital letters, spaces, or special characters in file names, branch names, directory names, headings, or in anything that generates a path or URL fragment.
- Don't create `README.md` files in docs. Name them `index.md` instead.

## Text

- Split up long lines (wrap text), this makes it much easier to review and edit. Only double line
breaks are shown as a full line Sourcegraph markdown 80-100. <!-- TODO(ryan): Link to ../../user/markdown.md once it has content. Fix in 3.3 -->
  characters is a good line length.
- Make sure that the documentation is added in the correct
  [directory](index.md#documentation-directory-structure) and that there's a link to it somewhere
  useful.
- Do not duplicate information.
- Be brief and clear.
- Unless there's a logical reason not to, add documents in alphabetical order.
- Write in US English.
- Use single spaces instead of double spaces.
- Jump a line between different markups (e.g., after every paragraph, header, list, etc)
- Capitalize the "S" in Sourcegraph (but not the "G").

## Formatting

- Use double asterisks (`**`) to mark a word or text in bold (`**bold**`).
- Use underscore (`_`) for text in italics (`_italic_`).
- Put an empty line between different markups. For example:

  ```md
  ## Header

  Paragraph.

  - List item
  - List item
  ```

### Ordered and unordered lists

- Use dashes (`-`) for unordered lists instead of asterisks (`*`).
- Use the number one (`1`) for ordered lists.
- Separate list items from explanatory text with a colon (`:`). For example:

  ```md
  The list is as follows:

  - First item: This explains the first item.
  - Second item: This explains the second item.
  ```

## Headings

- Add **only one H1** in each document, by adding `#` at the beginning of
  it (when using markdown). The `h1` will be the document `<title>`.
- Start with an h2 (`##`), and respect the order h2 > h3 > h4 > h5 > h6.
  Never skip the hierarchy level, such as h2 > h4
- Avoid putting numbers in headings. Numbers shift, hence documentation anchor
  links shift too, which eventually leads to dead links. If you think it is
  compelling to add numbers in headings, make sure to at least discuss it with
  someone in the pull request.
- Avoid using symbols and special chars in headers because they can't be reproduced in the URL fragment. Whenever possible, they should be plain and short text.
- Avoid adding things that show ephemeral statuses. For example, if a feature is
  considered beta or experimental, put this info in a note, not in the heading. This ensures the heading's URL fragment is stable.
- When introducing a new document, be careful for the filename and headings to be
  grammatically and syntactically correct. Mention one or all
  of the following Sourcegraph members for a review: `@ryan-blunden` or `@sqs`.
  This is to ensure that no document with wrong heading is going
  live without an audit, thus preventing dead links and redirection issues when
  corrected.
- Leave exactly one new line after a heading.

## Links

- Use the regular inline link markdown markup `[Text](https://example.com)`.
  It's easier to read, review, and maintain.
- If there's a link that repeats several times through the same document,
  you can use `[Text][identifier]` and at the bottom of the section or the
  document add: `[identifier]: https://example.com`, in which case, we do
  encourage you to also add an alternative text: `[identifier]: https://example.com "Alternative text"` that appears when hovering your mouse on a link.
- To link to internal documentation within the same repository, use relative links, not full URLs. Use `../` to navigate to higher-level directories, and always add the file name `file.md` at the
  end of the link with the `.md` extension, not `.html`.
  Example: instead of `[text](../../merge_requests/)`, use
  `[text](../../merge_requests/index.md)` or, `[text](../../ci/README.md)`, or,
  for anchor links, `[text](../../ci/README.md#examples)`.
  Using the markdown extension is necessary for the `/help` area in Sourcegraph.

## Navigation

To indicate the steps of navigation through the UI:

- Use the exact word as shown in the UI, including any capital letters as-is.
- Use bold text for navigation items and the char `>` as separator (e.g., `Navigate to the repository **Settings > Indexing**`).

## Images

- Place images in a separate directory named `img/` in the same directory where
  the `.md` document that you're working on is located. Always prepend their
  names with the name of the document that they will be included in. For
  example, if there is a document called `twitter.md`, then a valid image name
  could be `twitter_login_screen.png`.
- Images should have a specific, non-generic name that will differentiate them.
- Keep all file names in lower case.
- Consider using PNG images instead of JPEG.
- Compress all images with <https://tinypng.com/> or similar tool.
- Compress gifs with <https://ezgif.com/optimize> or similar tool.
- Images should be used (only when necessary) to _illustrate_ the description of a process, not to _replace_ it.
- Max image size: 100KB (GIFs included).
  - For larger assets, upload them to the `sourcegraph-assets` Google Cloud Storage bucket instead with `gsutil cp -a public-read local/path/to/myasset.png gs://sourcegraph-assets/` (and refer to it as `https://storage.googleapis.com/sourcegraph-assets/myasset.png`).

Inside the document:

- The Markdown way of using an image inside a document is:
  `![Proper description what the image is about](img/document_image_title.png)`
- Always use a proper description for what the image is about. That way, when a
  browser fails to show the image, this text will be used as an alternative
  description.
- If there are consecutive images with little text between them, always add
  three dashes (`---`) between the image and the text to create a horizontal
  line for better clarity.
- If a heading is placed right after an image, always add three dashes (`---`)
  between the image and the heading.

## Alert boxes

Whenever you want to call the attention to a particular sentence,
use the following markup for highlighting.

_Note that the alert boxes only work for one paragraph only. Multiple paragraphs,
lists, headers, etc will not render correctly._

### Note

```md
> NOTE: This is something to note.
```

How it renders on docs.sourcegraph.com:

> NOTE: This is something to note.

### Warning

```md
> WARNING: This is something to warning.
```

How it renders on docs.sourcegraph.com:

> WARNING: This is something to warning.

## Blockquotes

For highlighting a text within a blue blockquote, use this format:

```md
> This is a blockquote.
```

which renders in docs.sourcegraph.com to:

> This is a blockquote.

If the text spans across multiple lines it's OK to split the line.

## Specific sections and terms

To mention and/or reference specific terms in Sourcegraph, please follow the styles
below.

### Sourcegraph versions and tiers

> NOTE: Every feature should link to the blog post, issue, or pull request (in that order) that introduced it.

New features should be documented to include the Sourcegraph version in which they were introduced.

```md
> Introduced in Sourcegraph 8.3.
```

If the feature is only available in Sourcegraph Enterprise, mention that:

```md
> Introduced in Sourcegraph Enterprise 2.10.
```

<!-- TODO(sqs): Consider adding product tier badges. -->

<!-- TODO(sqs): Consider discussing how to describe things that differ on single-node vs. cluster. -->
