import { type Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { isErrorLike } from '@sourcegraph/common'
import { toAbsoluteBlobURL } from '@sourcegraph/shared/src/util/url'

import type { Block, BlockInit, BlockInput, FileBlockInput, SerializableBlock, SymbolBlockInput } from '..'
import {
    type CreateNotebookBlockInput,
    NotebookBlockType,
    SymbolKind,
    type HighlightLineRange,
    type NotebookFields,
} from '../../graphql-operations'
import { parseBrowserRepoURL } from '../../util/url'

export function serializeBlockToMarkdown(block: SerializableBlock, sourcegraphURL: string): Observable<string> {
    const serializedInput = serializeBlockInput(block, sourcegraphURL)
    switch (block.type) {
        case 'md': {
            return serializedInput.pipe(map(input => input.trimEnd()))
        }
        case 'query': {
            return serializedInput.pipe(map(input => `\`\`\`sourcegraph\n${input}\n\`\`\``))
        }
        case 'file':
        case 'symbol': {
            return serializedInput
        }
    }
}

export function serializeBlockInput(block: SerializableBlock, sourcegraphURL: string): Observable<string> {
    switch (block.type) {
        case 'md': {
            return of(block.input.text)
        }
        case 'query': {
            return of(block.input.query)
        }
        case 'file': {
            return of(
                toAbsoluteBlobURL(sourcegraphURL, {
                    repoName: block.input.repositoryName,
                    revision: block.input.revision,
                    filePath: block.input.filePath,
                    range: block.input.lineRange
                        ? {
                              start: { line: block.input.lineRange.startLine + 1, character: 0 },
                              end: { line: block.input.lineRange.endLine, character: 0 },
                          }
                        : undefined,
                })
            )
        }
        case 'symbol': {
            if (!block.output) {
                return of('')
            }
            return block.output.pipe(
                map(output => {
                    if (isErrorLike(output)) {
                        return ''
                    }
                    const blobURL = toAbsoluteBlobURL(sourcegraphURL, {
                        repoName: block.input.repositoryName,
                        revision: output.effectiveRevision,
                        filePath: block.input.filePath,
                        range: output.symbolRange,
                    })
                    const symbolParameters = new URLSearchParams([
                        ['symbolName', block.input.symbolName],
                        ['symbolContainerName', block.input.symbolContainerName],
                        ['symbolKind', block.input.symbolKind.toString()],
                        ['lineContext', block.input.lineContext.toString()],
                    ])
                    return blobURL + '#' + symbolParameters.toString()
                })
            )
        }
    }
}

export function parseFileBlockInput(input: string): FileBlockInput {
    try {
        const { repoName, rawRevision, filePath, position, range } = parseBrowserRepoURL(input)

        const lineRange = range
            ? { startLine: range.start.line - 1, endLine: range.end.line }
            : position
            ? { startLine: position.line - 1, endLine: position.line }
            : null

        return { repositoryName: repoName, revision: rawRevision ?? '', filePath: filePath ?? '', lineRange }
    } catch {
        return { repositoryName: '', revision: '', filePath: '', lineRange: null }
    }
}

function parseSymbolBlockInput(input: string): SymbolBlockInput {
    const defaultLineContext = 3
    try {
        const { repoName, rawRevision, filePath } = parseBrowserRepoURL(input)
        const url = new URL(input)
        const symbolParameters = new URLSearchParams(url.hash.slice(1))
        const lineContextValue = symbolParameters.get('lineContext')
        const lineContext = lineContextValue ? parseInt(lineContextValue, 10) : defaultLineContext
        return {
            repositoryName: repoName,
            revision: rawRevision ?? '',
            filePath: filePath ?? '',
            symbolName: symbolParameters.get('symbolName') ?? '',
            symbolContainerName: symbolParameters.get('symbolContainerName') ?? '',
            symbolKind: (symbolParameters.get('symbolKind') as SymbolKind) ?? SymbolKind.UNKNOWN,
            lineContext: !isNaN(lineContext) ? lineContext : defaultLineContext,
        }
    } catch {
        return {
            repositoryName: '',
            revision: '',
            filePath: '',
            symbolName: '',
            symbolContainerName: '',
            symbolKind: SymbolKind.UNKNOWN,
            lineContext: defaultLineContext,
        }
    }
}

export function deserializeBlockInput(type: Block['type'], input: string): BlockInput {
    switch (type) {
        case 'md': {
            return { type, input: { text: input } }
        }
        case 'query': {
            return { type, input: { query: input } }
        }
        case 'file': {
            return { type, input: parseFileBlockInput(input) }
        }
        case 'symbol': {
            return { type, input: parseSymbolBlockInput(input) }
        }
    }
}

export function isSingleLineRange(lineRange: HighlightLineRange | null): boolean {
    return lineRange ? lineRange.startLine + 1 === lineRange.endLine : false
}

export function serializeLineRange(lineRange: HighlightLineRange | null): string {
    if (!lineRange) {
        return ''
    }

    return isSingleLineRange(lineRange)
        ? `${lineRange.startLine + 1}`
        : `${lineRange.startLine + 1}-${lineRange.endLine}`
}

const LINE_RANGE_REGEX = /^(\d+)(-\d+)?$/

export function parseLineRange(value: string): HighlightLineRange | null {
    const matches = value.match(LINE_RANGE_REGEX)
    if (matches === null) {
        return null
    }
    const startLine = parseInt(matches[1], 10) - 1
    const endLine = matches[2] ? parseInt(matches[2].slice(1), 10) : startLine + 1
    return { startLine, endLine }
}

export function blockToGQLInput(block: BlockInit): CreateNotebookBlockInput {
    switch (block.type) {
        case 'md': {
            return { id: block.id, type: NotebookBlockType.MARKDOWN, markdownInput: block.input.text }
        }
        case 'query': {
            return { id: block.id, type: NotebookBlockType.QUERY, queryInput: block.input.query }
        }
        case 'file': {
            return { id: block.id, type: NotebookBlockType.FILE, fileInput: block.input }
        }
        case 'symbol': {
            return { id: block.id, type: NotebookBlockType.SYMBOL, symbolInput: block.input }
        }
    }
}

export function GQLBlockToGQLInput(block: NotebookFields['blocks'][number]): CreateNotebookBlockInput {
    switch (block.__typename) {
        case 'MarkdownBlock': {
            return { id: block.id, type: NotebookBlockType.MARKDOWN, markdownInput: block.markdownInput }
        }
        case 'QueryBlock': {
            return { id: block.id, type: NotebookBlockType.QUERY, queryInput: block.queryInput }
        }
        case 'FileBlock': {
            return {
                id: block.id,
                type: NotebookBlockType.FILE,
                fileInput: block.fileInput,
            }
        }
        case 'SymbolBlock': {
            return {
                id: block.id,
                type: NotebookBlockType.SYMBOL,
                symbolInput: block.symbolInput,
            }
        }
    }
}

export function convertNotebookTitleToFileName(title: string): string {
    return title.replaceAll(/[^\da-z]/gi, '_').replaceAll(/_+/g, '_')
}
