import { renderCodyMarkdown } from './markdown'

describe('URL handling', () => {
    it('allows vscode links', () => {
        expect(renderCodyMarkdown('[A vscode link](vscode://foo)')).toBe(
            // TODO: do we actually want VS code links to open this way?
            '<p><a href="vscode://foo" target="_blank" rel="noopener">A vscode link</a></p>'
        )
    })

    it('allows https links', () => {
        expect(renderCodyMarkdown('[An eraser link](https://app.eraser.io/foo)')).toBe(
            '<p><a href="https://app.eraser.io/foo" target="_blank" rel="noopener">An eraser link</a></p>'
        )
    })

    it('allows google storage images', () => {
        const imageUrl =
            'https://storage.googleapis.com/foo.appspot.com/elements/elements%3A8676d010c2a70b75eecf9f2406fbb06e466673345ad4a144d09bc380b49b848a.png'

        expect(renderCodyMarkdown(`![A google storage image](${imageUrl})`)).toBe(
            `<p><img alt="A google storage image" src="${imageUrl}"></p>`
        )
    })
})
