import { toAbsoluteBlobURL } from '@sourcegraph/shared/src/util/url'

import { parseBrowserRepoURL } from '../../util/url'

import { Block, BlockInput } from '.'

export function serializeBlockInput(block: BlockInput, sourcegraphURL: string): string {
    switch (block.type) {
        case 'md':
        case 'query':
            return block.input
        case 'file':
            return toAbsoluteBlobURL(sourcegraphURL, {
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
    }
}

export function deserializeBlockInput(type: Block['type'], input: string): BlockInput {
    switch (type) {
        case 'md':
        case 'query':
            return { type, input }
        case 'file': {
            const { repoName, rawRevision, filePath, range } = parseBrowserRepoURL(input)
            return {
                type,
                input: {
                    repositoryName: repoName,
                    revision: rawRevision ?? '',
                    filePath: filePath ?? '',
                    lineRange: range ? { startLine: range.start.line - 1, endLine: range.end.line } : undefined,
                },
            }
        }
    }
}
