import { renderMarkdown } from './markdown'

describe('renderMarkdown', () => {
    test('plainText option', () => expect(renderMarkdown('A **b**', { plainText: true })).toBe('A b\n'))

    describe('allowDataUriLinksAndDownloads option', () => {
        const MARKDOWN_WITH_DOWNLOAD = '<a href="data:text/plain,foobar" download>D</a>\n[D2](data:text/plain,foobar)'
        test('default disabled', () =>
            expect(renderMarkdown(MARKDOWN_WITH_DOWNLOAD)).toBe('<p><a>D</a><br /><a>D2</a></p>\n'))
        test('enabled', () =>
            expect(renderMarkdown(MARKDOWN_WITH_DOWNLOAD, { allowDataUriLinksAndDownloads: true })).toBe(
                '<p><a href="data:text/plain,foobar" download>D</a><br /><a href="data:text/plain,foobar">D2</a></p>\n'
            ))
    })
})
