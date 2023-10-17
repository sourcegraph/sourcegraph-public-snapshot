import { markdownLexer } from '@sourcegraph/common'

import type { BlockInput } from '..'
import { parseBrowserRepoURL } from '../../util/url'

import { deserializeBlockInput } from '.'

function isSourcegraphFileBlobURL(url: string): boolean {
    return !!parseBrowserRepoURL(url).filePath
}

function isSymbolBlockURL(url: string): boolean {
    const parsedURL = new URL(url)
    const symbolParameters = new URLSearchParams(parsedURL.hash.slice(1))
    const symbolName = symbolParameters.get('symbolName')
    return symbolName !== null && symbolName.length > 0
}

export function convertMarkdownToBlocks(markdown: string): BlockInput[] {
    const blocks: BlockInput[] = []

    let markdownRawTokens: string[] = []
    const addMarkdownBlock = (): void => {
        if (markdownRawTokens.length === 0) {
            return
        }
        const markdownInput = markdownRawTokens.join('').trimStart()
        if (markdownInput.length > 0) {
            blocks.push(deserializeBlockInput('md', markdownInput))
        }
        markdownRawTokens = []
    }

    for (const token of markdownLexer(markdown)) {
        if (token.type === 'code' && token.lang === 'sourcegraph') {
            addMarkdownBlock()
            blocks.push(deserializeBlockInput('query', token.text))
        } else if (
            token.type === 'paragraph' &&
            token.tokens.length === 1 &&
            token.tokens[0].type === 'link' &&
            isSourcegraphFileBlobURL(token.tokens[0].href)
        ) {
            addMarkdownBlock()
            const blockType = isSymbolBlockURL(token.text) ? 'symbol' : 'file'
            blocks.push(deserializeBlockInput(blockType, token.text))
        } else {
            markdownRawTokens.push(token.raw)
        }
    }
    addMarkdownBlock()

    return blocks
}
