import type { Observable } from 'rxjs'

import type { HighlightRange } from '@sourcegraph/branded'
import type { FetchFileParameters } from '@sourcegraph/shared/src/backend/file'
import type { AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'
import type { UIRangeSpec } from '@sourcegraph/shared/src/util/url'

import type { HighlightLineRange, SymbolKind } from '../graphql-operations'

// When adding a new block type, make sure to track its usage in internal/usagestats/notebooks.go.
export type BlockType = 'md' | 'query' | 'file' | 'compute' | 'symbol'

interface BaseBlock<I, O> {
    id: string
    type: BlockType
    input: I
    output: O | null
}

export interface QueryBlockInput {
    query: string
    initialFocusInput?: boolean
}

export interface QueryBlock extends BaseBlock<QueryBlockInput, Observable<AggregateStreamingSearchResults>> {
    type: 'query'
}

export interface MarkdownBlockInput {
    text: string
    initialFocusInput?: boolean
}

export interface MarkdownBlock extends BaseBlock<MarkdownBlockInput, string> {
    type: 'md'
}

export interface FileBlockInput {
    repositoryName: string
    revision: string
    filePath: string
    lineRange: HighlightLineRange | null
    initialQueryInput?: string
}

export interface FileBlock extends BaseBlock<FileBlockInput, Observable<string[] | Error>> {
    type: 'file'
}

export interface SymbolBlockInput {
    repositoryName: string
    revision: string
    filePath: string
    symbolName: string
    symbolKind: SymbolKind
    symbolContainerName: string
    lineContext: number
    initialQueryInput?: string
}

export interface SymbolBlockOutput {
    symbolFoundAtLatestRevision: boolean
    effectiveRevision: string
    symbolRange: UIRangeSpec['range']
    highlightLineRange: HighlightLineRange
    highlightedLines: string[]
    highlightSymbolRange: HighlightRange
}

export interface SymbolBlock extends BaseBlock<SymbolBlockInput, Observable<SymbolBlockOutput | Error>> {
    type: 'symbol'
}

export type Block = QueryBlock | MarkdownBlock | FileBlock | SymbolBlock

export type BlockInput =
    | Pick<FileBlock, 'type' | 'input'>
    | Pick<MarkdownBlock, 'type' | 'input'>
    | Pick<QueryBlock, 'type' | 'input'>
    | Pick<SymbolBlock, 'type' | 'input'>

export type BlockInit =
    | Omit<FileBlock, 'output'>
    | Omit<MarkdownBlock, 'output'>
    | Omit<QueryBlock, 'output'>
    | Omit<SymbolBlock, 'output'>

export type SerializableBlock =
    | Pick<FileBlock, 'type' | 'input'>
    | Pick<MarkdownBlock, 'type' | 'input'>
    | Pick<QueryBlock, 'type' | 'input'>
    | Pick<SymbolBlock, 'type' | 'input' | 'output'>

export type BlockDirection = 'up' | 'down'

export interface BlockProps<T extends Block = Block> {
    isReadOnly: boolean
    isSelected: boolean
    showMenu: boolean
    id: T['id']
    input: T['input']
    output: T['output']
    onRunBlock(id: string): void
    onDeleteBlock(id: string): void
    onBlockInputChange(id: string, blockInput: BlockInput): void
    onMoveBlock(id: string, direction: BlockDirection): void
    onDuplicateBlock(id: string): void
    onNewBlock(id: string): void
}

export interface BlockDependencies {
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
}

export interface NotebookProps {
    notebooksEnabled: boolean
}
