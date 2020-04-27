import { renderMarkdown } from './markdown'

describe('renderMarkdown', () => {
    it('renders to plain text with plainText: true', () => {
        expect(renderMarkdown('A **b**', { plainText: true })).toBe('A b\n')
    })
    it('sanitizes script tags', () => {
        expect(renderMarkdown('<script>evil();</script>')).toBe('')
    })
    it('sanitizes event handlers', () => {
        expect(renderMarkdown('<svg><rect onclick="evil()"></rect></svg>')).toBe('<p><svg><rect></rect></svg></p>\n')
    })
    it('sanitizes non-SVG <object> tags', () => {
        expect(renderMarkdown('<object data="something"></object>')).toBe('<p></p>\n')
    })
    it('allows SVG <object> tags', () => {
        expect(renderMarkdown('<object data="something" type="image/svg+xml"></object>')).toBe(
            '<p><object data="something" type="image/svg+xml"></object></p>\n'
        )
    })
    it('allows <svg> tags', () => {
        const input =
            '<svg viewbox="10 10 10 10" width="100"><rect x="37.5" y="7.5" width="675.0" height="16.875" fill="#e05d44" stroke="white" stroke-width="1"><title>/</title></rect></svg>'
        expect(renderMarkdown(input)).toBe(`<p>${input}</p>\n`)
    })

    describe('allowDataUriLinksAndDownloads option', () => {
        const MARKDOWN_WITH_DOWNLOAD = '<a href="data:text/plain,foobar" download>D</a>\n[D2](data:text/plain,foobar)'
        test('default disabled', () => {
            expect(renderMarkdown(MARKDOWN_WITH_DOWNLOAD)).toBe('<p><a>D</a>\n<a>D2</a></p>\n')
        })
        test('enabled', () => {
            expect(renderMarkdown(MARKDOWN_WITH_DOWNLOAD, { allowDataUriLinksAndDownloads: true })).toBe(
                '<p><a href="data:text/plain,foobar" download>D</a>\n<a href="data:text/plain,foobar">D2</a></p>\n'
            )
        })
    })
})
