import { Remote } from 'comlink'
import { Observable } from 'rxjs'

import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { IHighlightLineRange } from '@sourcegraph/shared/src/schema'
import { AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'

export type BlockType = 'md' | 'query' | 'file' | 'compute'

interface BaseBlock<I, O> {
    id: string
    type: BlockType
    input: I
    output: O | null
}

export interface QueryBlock extends BaseBlock<string, Observable<AggregateStreamingSearchResults>> {
    type: 'query'
}

export interface MarkdownBlock extends BaseBlock<string, string> {
    type: 'md'
}

export interface FileBlockInput {
    repositoryName: string
    revision: string
    filePath: string
    lineRange: IHighlightLineRange | null
}

export interface FileBlock extends BaseBlock<FileBlockInput, Observable<string[] | Error>> {
    type: 'file'
}

export interface ComputeBlock extends BaseBlock<string, string> {
    type: 'compute'
}

export type Block = QueryBlock | MarkdownBlock | FileBlock | ComputeBlock

export type BlockInput =
    | Pick<FileBlock, 'type' | 'input'>
    | Pick<MarkdownBlock, 'type' | 'input'>
    | Pick<QueryBlock, 'type' | 'input'>
    | Pick<ComputeBlock, 'type' | 'input'>

export type BlockInit =
    | Omit<FileBlock, 'output'>
    | Omit<MarkdownBlock, 'output'>
    | Omit<QueryBlock, 'output'>
    | Omit<ComputeBlock, 'output'>

export type BlockDirection = 'up' | 'down'

export interface BlockProps {
    isReadOnly: boolean
    isSelected: boolean
    isOtherBlockSelected: boolean
    onRunBlock(id: string): void
    onDeleteBlock(id: string): void
    onBlockInputChange(id: string, blockInput: BlockInput): void
    onSelectBlock(id: string | null): void
    onMoveBlockSelection(id: string, direction: BlockDirection): void
    onMoveBlock(id: string, direction: BlockDirection): void
    onDuplicateBlock(id: string): void
}

export interface BlockDependencies {
    extensionHostAPI: Promise<Remote<FlatExtensionHostAPI>>
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
}
