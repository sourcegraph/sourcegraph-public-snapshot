import { markdownLexer } from '@sourcegraph/shared/src/util/markdown'

import { deserializeBlockInput } from './serialize'

import { BlockInput } from '.'

export function convertMarkdownToBlocks(markdown: string): BlockInput[] {
    const blocks: BlockInput[] = []

    let markdownRawTokens: string[] = []
    const addMarkdownBlock = (): void => {
        if (markdownRawTokens.length === 0) {
            return
        }
        blocks.push(deserializeBlockInput('md', markdownRawTokens.join('').trimStart()))
        markdownRawTokens = []
    }

    for (const token of markdownLexer(markdown)) {
        if (token.type === 'code' && token.lang === 'sourcegraph') {
            addMarkdownBlock()
            blocks.push(deserializeBlockInput('query', token.text))
        } else if (token.type === 'code' && token.lang === 'sourcegraph:file') {
            addMarkdownBlock()
            blocks.push(deserializeBlockInput('file', token.text))
        } else {
            markdownRawTokens.push(token.raw)
        }
    }
    addMarkdownBlock()

    return blocks
}
