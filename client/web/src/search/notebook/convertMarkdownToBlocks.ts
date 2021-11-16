import { markdownLexer } from '@sourcegraph/shared/src/util/markdown'

import { parseBrowserRepoURL } from '../../util/url'

import { deserializeBlockInput } from './serialize'

import { BlockInput } from '.'

function isSourcegraphFileBlobURL(url: string): boolean {
    return !!parseBrowserRepoURL(url).filePath
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
            blocks.push(deserializeBlockInput('file', token.text))
        } else {
            markdownRawTokens.push(token.raw)
        }
    }
    addMarkdownBlock()

    return blocks
}
