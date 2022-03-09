import { Observable, forkJoin, of } from 'rxjs'
import { startWith, catchError, mapTo, map } from 'rxjs/operators'
import * as uuid from 'uuid'

import { renderMarkdown, asError } from '@sourcegraph/common'
import { transformSearchQuery } from '@sourcegraph/shared/src/api/client/search'
import { aggregateStreamingSearch, emptyAggregateResults } from '@sourcegraph/shared/src/search/stream'

import { Block, BlockInit, BlockDependencies, BlockInput, BlockDirection } from '..'
import { NotebookFields, SearchPatternType } from '../../graphql-operations'
import { LATEST_VERSION } from '../../search/results/StreamingSearchResults'
import { createNotebook } from '../backend'
import { blockToGQLInput, serializeBlockToMarkdown } from '../serialize'

const DONE = 'DONE' as const

export interface CopyNotebookProps {
    title: string
    blocks: BlockInit[]
    namespace: string
}

export function copyNotebook({ title, blocks, namespace }: CopyNotebookProps): Observable<NotebookFields> {
    return createNotebook({
        notebook: {
            title,
            blocks: blocks.flatMap(block => (block.type === 'compute' ? [] : [blockToGQLInput(block)])),
            namespace,
            public: false,
        },
    })
}

export class Notebook {
    private blocks: Map<string, Block>
    private blockOrder: string[]

    constructor(initializerBlocks: BlockInit[], private dependencies: BlockDependencies) {
        const blocks = initializerBlocks.map(block => ({ ...block, output: null }))

        this.blocks = new Map(blocks.map(block => [block.id, block]))
        this.blockOrder = blocks.map(block => block.id)

        // Pre-run the markdown and file blocks, for a better user experience.
        for (const block of blocks) {
            if (block.type === 'md' || block.type === 'file') {
                this.runBlockById(block.id)
            }
        }
    }

    public getBlocks(): Block[] {
        return this.blockOrder.map(blockId => {
            const block = this.blocks.get(blockId)
            if (!block) {
                throw new Error(`Block with id:${blockId} does not exist.`)
            }
            return block
        })
    }

    public setBlockInputById(id: string, { type, input }: BlockInput): void {
        const block = this.blocks.get(id)
        if (!block) {
            return
        }
        if (block.type !== type) {
            throw new Error(`Input block type ${type} does not match existing block type ${block.type}.`)
        }
        // We checked that the existing block and the input block have the same type, so the cast below is safe.
        this.blocks.set(block.id, { ...block, input } as Block)
    }

    public runBlockById(id: string): void {
        const block = this.blocks.get(id)
        if (!block) {
            return
        }
        switch (block.type) {
            case 'md':
                this.blocks.set(block.id, { ...block, output: renderMarkdown(block.input) })
                break
            case 'query':
                this.blocks.set(block.id, {
                    ...block,
                    output: aggregateStreamingSearch(
                        transformSearchQuery({
                            // Removes comments
                            query: block.input.replace(/\/\/.*/g, ''),
                            extensionHostAPIPromise: this.dependencies.extensionHostAPI,
                        }),
                        {
                            version: LATEST_VERSION,
                            patternType: SearchPatternType.literal,
                            caseSensitive: false,
                            trace: undefined,
                        }
                    ).pipe(startWith(emptyAggregateResults)),
                })
                break
            case 'file':
                this.blocks.set(block.id, {
                    ...block,
                    output: this.dependencies
                        .fetchHighlightedFileLineRanges({
                            repoName: block.input.repositoryName,
                            commitID: block.input.revision || 'HEAD',
                            filePath: block.input.filePath,
                            ranges: block.input.lineRange
                                ? [block.input.lineRange]
                                : [{ startLine: 0, endLine: 2147483647 }], // entire file,
                            disableTimeout: false,
                        })
                        .pipe(
                            map(ranges => ranges[0]),
                            catchError(() => [asError('File not found')])
                        ),
                })
                break
        }
    }

    public runAllBlocks(): Observable<typeof DONE[]> {
        const observables: Observable<typeof DONE>[] = []
        // Iterate over block ids and run each block. We do not iterate over values
        // because `runBlockById` method assigns a new value for the id so we have
        // to fetch the block value separately.
        for (const blockId of this.blocks.keys()) {
            this.runBlockById(blockId)

            const block = this.getBlockById(blockId)
            if (!block?.output) {
                continue
            }
            // Identical if/else if branches to make the TS compiler happy
            if (block.type === 'query') {
                observables.push(block.output.pipe(mapTo(DONE)))
            } else if (block.type === 'file') {
                observables.push(block.output.pipe(mapTo(DONE)))
            }
        }
        // We store output observables and join them into a single observable,
        // to let the caller know when all async outputs have finished.
        return observables.length > 0 ? forkJoin(observables) : of([DONE])
    }

    public deleteBlockById(id: string): void {
        const index = this.blockOrder.indexOf(id)
        if (index === -1) {
            return
        }
        this.blocks.delete(id)
        this.blockOrder.splice(index, 1)
    }

    public insertBlockAtIndex(index: number, blockToInsert: BlockInput): Block {
        const id = uuid.v4()
        const block = { ...blockToInsert, id, output: null }
        // Insert block at the provided index
        this.blockOrder.splice(index, 0, id)
        this.blocks.set(id, block)
        return block
    }

    public getBlockById(id: string): Block | null {
        return this.blocks.get(id) ?? null
    }

    public moveBlockById(id: string, direction: BlockDirection): void {
        const index = this.blockOrder.indexOf(id)
        if ((direction === 'up' && index < 1) || (direction === 'down' && index === this.blockOrder.length - 1)) {
            return
        }
        const swapIndex = direction === 'up' ? index - 1 : index + 1
        const temporaryId = this.blockOrder[swapIndex]
        this.blockOrder[swapIndex] = id
        this.blockOrder[index] = temporaryId
    }

    public duplicateBlockById(id: string): Block | null {
        const block = this.blocks.get(id)
        if (!block) {
            return null
        }
        const index = this.blockOrder.indexOf(id)
        return this.insertBlockAtIndex(index + 1, block)
    }

    public getFirstBlockId(): string | null {
        return this.blockOrder.length > 0 ? this.blockOrder[0] : null
    }

    public getLastBlockId(): string | null {
        return this.blockOrder.length > 0 ? this.blockOrder[this.blockOrder.length - 1] : null
    }

    public getPreviousBlockId(id: string): string | null {
        const index = this.blockOrder.indexOf(id)
        return index >= 1 ? this.blockOrder[index - 1] : null
    }

    public getNextBlockId(id: string): string | null {
        const index = this.blockOrder.indexOf(id)
        return index >= 0 && index < this.blockOrder.length - 1 ? this.blockOrder[index + 1] : null
    }

    public exportToMarkdown(sourcegraphURL: string): string {
        return (
            this.getBlocks()
                .map(block => serializeBlockToMarkdown(block, sourcegraphURL))
                .join('\n\n') + '\n'
        )
    }
}
