import { renderMarkdown, registerHighlightContributions } from '.'

registerHighlightContributions()

describe('renderMarkdown', () => {
    it('renders code blocks, with syntax highlighting', () => {
        const markdown = [
            '# This is a heading',
            '',
            '## This is a subheading',
            '',
            'Some text',
            'in the same paragraph',
            'with a [link](./destination).',
            '',
            '```ts',
            'const someTypeScriptCode = funcCall()',
            '```',
            '',
            '- bullet list item 1',
            '- bullet list item 2',
            '',
            '1. item 1',
            '  ```ts',
            '  const codeInsideTheBulletPoint = "string"',
            '  ```',
            '1. item 2',
            '',
            '> quoted',
            '> text',
            '',
            '| col 1 | col 2 |',
            '|-------|-------|',
            '| A     | B     |',
            '',
            '![image alt text](./src.jpg)',
        ].join('\n')
        expect(renderMarkdown(markdown)).toMatchInlineSnapshot(`
            "<h1 id=\\"this-is-a-heading\\">This is a heading</h1>
            <h2 id=\\"this-is-a-subheading\\">This is a subheading</h2>
            <p>Some text
            in the same paragraph
            with a <a href=\\"./destination\\">link</a>.</p>
            <pre><code class=\\"language-ts\\"><span class=\\"hljs-keyword\\">const</span> someTypeScriptCode = funcCall()
            </code></pre>
            <ul>
            <li>bullet list item 1</li>
            <li>bullet list item 2</li>
            </ul>
            <ol>
            <li>item 1<pre><code class=\\"language-ts\\"><span class=\\"hljs-keyword\\">const</span> codeInsideTheBulletPoint = <span class=\\"hljs-string\\">\\"string\\"</span>
            </code></pre>
            </li>
            <li>item 2</li>
            </ol>
            <blockquote>
            <p>quoted
            text</p>
            </blockquote>
            <table>
            <thead>
            <tr>
            <th>col 1</th>
            <th>col 2</th>
            </tr>
            </thead>
            <tbody><tr>
            <td>A</td>
            <td>B</td>
            </tr>
            </tbody></table>
            <p><img alt=\\"image alt text\\" src=\\"./src.jpg\\"></p>"
        `)
    })
    it('renders to plain text with plainText: true', () => {
        expect(renderMarkdown('A **b**', { plainText: true })).toBe('A b')
    })
    it('sanitizes script tags', () => {
        expect(renderMarkdown('<script>evil();</script>')).toBe('')
    })
    it('sanitizes event handlers', () => {
        expect(renderMarkdown('<svg><rect onclick="evil()"></rect></svg>')).toBe('<p></p>')
    })
    it('does not allow arbitrary <object> tags', () => {
        expect(renderMarkdown('<object data="something"></object>')).toBe('<p></p>')
    })
    it('drops SVG <object> tags', () => {
        expect(renderMarkdown('<object data="something" type="image/svg+xml"></object>')).toBe('<p></p>')
    })
    it('forbids <svg> tags', () => {
        const input =
            '<svg viewbox="10 10 10 10" width="100"><rect x="37.5" y="7.5" width="675.0" height="16.875" fill="#e05d44" stroke="white" stroke-width="1"><title>/</title></rect></svg>'
        expect(renderMarkdown(input)).toBe('<p></p>')
    })
    it('forbids rel and style attributes', () => {
        const input = '<a href="/" rel="evil" style="foo:bar">Link</a><script>alert("x")</script>'
        expect(renderMarkdown(input)).toBe('<p><a href="/">Link</a></p>')
    })
    test('forbids data URI links', () => {
        const input = '<a href="data:text/plain,foobar" download>D</a>\n[D2](data:text/plain,foobar)'
        expect(renderMarkdown(input)).toBe('<p><a download="">D</a>\n<a>D2</a></p>')
    })
})
