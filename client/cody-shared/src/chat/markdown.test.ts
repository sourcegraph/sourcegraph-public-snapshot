import { escapeCodyMarkdown } from './markdown'

describe('escapeCodyMarkdown', () => {
    it('escapes paragraphs', () => {
        const markdown = 'Paragraph &1\n\nParagraph &2'
        const result = escapeCodyMarkdown(markdown, false)
        expect(result).toMatchInlineSnapshot(`
            "Paragraph &amp;1

            Paragraph &amp;2"
        `)
    })

    it('does not escape inside code blocks', () => {
        const markdown = '&outside\n```js\ncode &block\n```'
        const result = escapeCodyMarkdown(markdown, false)
        expect(result).toMatchInlineSnapshot(`
            "&amp;outside
            \`\`\`js
            code &block
            \`\`\`"
        `)
    })

    it('does not escape inside code blocks with languages', () => {
        const markdown = '&outside\n```python\ncode &block\n```'
        const result = escapeCodyMarkdown(markdown, false)
        expect(result).toMatchInlineSnapshot(`
            "&amp;outside
            \`\`\`python
            code &block
            \`\`\`"
        `)
    })

    it('properly escapes with a mix of elements', () => {
        const markdown = `
&Paragraph

\`\`\`js
code &block
\`\`\`

# &Header

&Paragraph
    `
        const result = escapeCodyMarkdown(markdown, false)
        expect(result).toMatchInlineSnapshot(`
            "
            &amp;Paragraph

            \`\`\`js
            code &block
            \`\`\`

            # &amp;Header

            &amp;Paragraph
                "
        `)
    })

    it('does not escape inside tiled code blocks', () => {
        const markdown = `
      ~~~js
      code &block
      ~~~

      Paragraph

      ~~~python
      code &block
      ~~~
      `
        const result = escapeCodyMarkdown(markdown, false)
        expect(result).toMatchInlineSnapshot(`
            "
                  ~~~js
                  code &block
                  ~~~

                  Paragraph

                  ~~~python
                  code &block
                  ~~~
                  "
        `)
    })

    it('does not escape inside inline code blocks', () => {
        const markdown = '&outside `code &block`'
        const result = escapeCodyMarkdown(markdown, false)
        expect(result).toMatchInlineSnapshot('"&amp;outside `code &block`"')
    })
})
