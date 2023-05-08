import { parseMarkdown } from './markdown'

describe('parseMarkdown', () => {
    it('parses paragraphs', () => {
        const markdown = 'Paragraph 1\n\nParagraph 2'
        const result = parseMarkdown(markdown)
        expect(result).toMatchInlineSnapshot(`
            Array [
              Object {
                "isCodeBlock": false,
                "line": "Paragraph 1",
              },
              Object {
                "isCodeBlock": false,
                "line": "",
              },
              Object {
                "isCodeBlock": false,
                "line": "Paragraph 2",
              },
            ]
        `)
    })

    it('parses code blocks', () => {
        const markdown = '```js\ncode block\n```'
        const result = parseMarkdown(markdown)
        expect(result).toMatchInlineSnapshot(`
            Array [
              Object {
                "isCodeBlock": true,
                "line": "\`\`\`js",
              },
              Object {
                "isCodeBlock": true,
                "line": "code block",
              },
              Object {
                "isCodeBlock": true,
                "line": "\`\`\`",
              },
            ]
        `)
    })

    it('parses code blocks with languages', () => {
        const markdown = '```python\ncode block\n```'
        const result = parseMarkdown(markdown)
        expect(result).toMatchInlineSnapshot(`
            Array [
              Object {
                "isCodeBlock": true,
                "line": "\`\`\`python",
              },
              Object {
                "isCodeBlock": true,
                "line": "code block",
              },
              Object {
                "isCodeBlock": true,
                "line": "\`\`\`",
              },
            ]
        `)
    })

    it('parses a mix of elements', () => {
        const markdown = `
Paragraph

\`\`\`js
code block
\`\`\`

# Header

Paragraph
    `
        const result = parseMarkdown(markdown)
        expect(result).toMatchInlineSnapshot(`
            Array [
              Object {
                "isCodeBlock": false,
                "line": "",
              },
              Object {
                "isCodeBlock": false,
                "line": "Paragraph",
              },
              Object {
                "isCodeBlock": false,
                "line": "",
              },
              Object {
                "isCodeBlock": true,
                "line": "\`\`\`js",
              },
              Object {
                "isCodeBlock": true,
                "line": "code block",
              },
              Object {
                "isCodeBlock": true,
                "line": "\`\`\`",
              },
              Object {
                "isCodeBlock": false,
                "line": "",
              },
              Object {
                "isCodeBlock": false,
                "line": "# Header",
              },
              Object {
                "isCodeBlock": false,
                "line": "",
              },
              Object {
                "isCodeBlock": false,
                "line": "Paragraph",
              },
              Object {
                "isCodeBlock": false,
                "line": "    ",
              },
            ]
        `)
    })

    it('parses tiled code blocks', () => {
        const markdown = `
      ~~~js
      code block
      ~~~

      Paragraph

      ~~~python
      code block
      ~~~
      `
        const result = parseMarkdown(markdown)
        expect(result).toMatchInlineSnapshot(`
            Array [
              Object {
                "isCodeBlock": false,
                "line": "",
              },
              Object {
                "isCodeBlock": true,
                "line": "      ~~~js",
              },
              Object {
                "isCodeBlock": true,
                "line": "      code block",
              },
              Object {
                "isCodeBlock": true,
                "line": "      ~~~",
              },
              Object {
                "isCodeBlock": false,
                "line": "",
              },
              Object {
                "isCodeBlock": false,
                "line": "      Paragraph",
              },
              Object {
                "isCodeBlock": false,
                "line": "",
              },
              Object {
                "isCodeBlock": true,
                "line": "      ~~~python",
              },
              Object {
                "isCodeBlock": true,
                "line": "      code block",
              },
              Object {
                "isCodeBlock": true,
                "line": "      ~~~",
              },
              Object {
                "isCodeBlock": false,
                "line": "      ",
              },
            ]
        `)
    })
})
