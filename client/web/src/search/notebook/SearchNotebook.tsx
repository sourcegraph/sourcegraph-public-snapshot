import * as Monaco from 'monaco-editor'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Observable, Subscription, Unsubscribable } from 'rxjs'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql/schema'
import { getProvidersNoCache } from '@sourcegraph/shared/src/search/query/providers'
import { SearchSuggestion } from '@sourcegraph/shared/src/search/suggestions'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { SearchStreamingProps } from '..'
import { fetchSuggestions } from '../backend'
import { SOURCEGRAPH_SEARCH } from '../input/MonacoQueryInput'
import { StreamingSearchResultsListProps } from '../results/StreamingSearchResultsList'

import { SearchNotebookAddBlockButtons } from './SearchNotebookAddBlockButtons'
import { SearchNotebookMarkdownBlock } from './SearchNotebookMarkdownBlock'
import { SearchNotebookQueryBlock } from './SearchNotebookQueryBlock'

import { Block, BlockInitializer, BlockType, Notebook } from '.'

interface SearchNotebookProps
    extends SearchStreamingProps,
        ThemeProps,
        Omit<StreamingSearchResultsListProps, 'allExpanded'> {
    globbing: boolean
    isMacPlatform: boolean

    onBlocksChange: (blocks: Block[]) => void
    blocks: BlockInitializer[]
}

const toUnsubscribable = (disposable: Monaco.IDisposable): Unsubscribable => ({
    unsubscribe: () => disposable.dispose(),
})

// TODO: Consolidate with MonacoQueryInput#addSourcegraphSearchCodeIntelligence
function addSourcegraphSearchCodeIntelligence(
    fetchSuggestions: (query: string) => Observable<SearchSuggestion[]>,
    options: {
        patternType: SearchPatternType
        globbing: boolean
        interpretComments?: boolean
        isSourcegraphDotCom?: boolean
    }
): Subscription {
    const subscriptions = new Subscription()

    // Register language ID
    Monaco.languages.register({ id: SOURCEGRAPH_SEARCH })

    // Register providers
    const providers = getProvidersNoCache(fetchSuggestions, options)
    subscriptions.add(toUnsubscribable(Monaco.languages.setTokensProvider(SOURCEGRAPH_SEARCH, providers.tokens)))
    subscriptions.add(toUnsubscribable(Monaco.languages.registerHoverProvider(SOURCEGRAPH_SEARCH, providers.hover)))
    subscriptions.add(
        toUnsubscribable(Monaco.languages.registerCompletionItemProvider(SOURCEGRAPH_SEARCH, providers.completion))
    )

    // TODO
    // subscriptions.add(
    //     providers.diagnostics.subscribe(markers => {
    //         monaco.editor.setModelMarkers(monaco.editor.getModels()[0], 'diagnostics', markers)
    //     })
    // )

    return subscriptions
}

export const SearchNotebook: React.FunctionComponent<SearchNotebookProps> = ({ onBlocksChange, ...props }) => {
    const notebook = useMemo(() => new Notebook(props.blocks), [props.blocks])

    const [selectedBlockId, setSelectedBlockId] = useState<string | null>(null)
    const [blocks, setBlocks] = useState<Block[]>(notebook.getBlocks())

    const onRunBlock = useCallback(
        (id: string) => {
            notebook.runBlockById(id)
            const blocks = notebook.getBlocks()
            setBlocks(blocks)
            onBlocksChange(blocks)
        },
        [notebook, onBlocksChange]
    )

    const onBlockInputChange = useCallback(
        (id: string, value: string) => {
            notebook.setBlockInputById(id, value)
            setBlocks(notebook.getBlocks())
        },
        [notebook]
    )

    const onAddBlock = useCallback(
        (index: number, type: BlockType, input: string) => {
            const addedBlock = notebook.insertBlockAtIndex(index, type, input)
            if (addedBlock.type === 'md') {
                notebook.runBlockById(addedBlock.id)
            }
            setSelectedBlockId(addedBlock.id)
            setBlocks(notebook.getBlocks())
        },
        [notebook, setBlocks]
    )

    const onDeleteBlock = useCallback(
        (id: string) => {
            const blockToFocusAfterDelete = notebook.getNextBlockId(id) ?? notebook.getPreviousBlockId(id)
            notebook.deleteBlockById(id)
            setSelectedBlockId(blockToFocusAfterDelete)
            setBlocks(notebook.getBlocks())
        },
        [notebook]
    )

    const onSelectBlock = useCallback(
        (id: string) => {
            setSelectedBlockId(id)
        },
        [setSelectedBlockId]
    )

    const onMoveBlockSelection = useCallback(
        (id: string, direction: 'up' | 'down') => {
            const blockId = direction === 'up' ? notebook.getPreviousBlockId(id) : notebook.getNextBlockId(id)
            if (blockId) {
                setSelectedBlockId(blockId)
            }
        },
        [notebook, setSelectedBlockId]
    )

    // Check all clicks on the document and deselect the currently selected block
    // if it was triggered outside of a block.
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent): void => {
            if (!event.target) {
                return
            }
            const target = event.target as HTMLElement
            // Check if the event target has a block ancestor
            const closestTargetCell = target.closest('[data-block-id]')
            if (!closestTargetCell) {
                setSelectedBlockId(null)
            }
        }
        const handleKeyDown = (event: KeyboardEvent): void => {
            // TODO: Scroll selected block into view (without being annoying)
            if (!selectedBlockId && event.key === 'ArrowDown') {
                setSelectedBlockId(notebook.getFirstBlockId())
            } else if (event.key === 'Escape') {
                setSelectedBlockId(null)
            }
        }

        document.addEventListener('mousedown', handleClickOutside)
        document.addEventListener('keydown', handleKeyDown)
        return () => {
            document.removeEventListener('mousedown', handleClickOutside)
            document.removeEventListener('keydown', handleKeyDown)
        }
    }, [notebook, selectedBlockId, onMoveBlockSelection])

    useEffect(() => {
        // Initialize Sourcegraph Monaco code intelligence (hovers, completions)
        const subscription = addSourcegraphSearchCodeIntelligence(fetchSuggestions, {
            // TODO?
            patternType: SearchPatternType.literal,
            globbing: props.globbing,
            interpretComments: true,
        })
        return () => subscription.unsubscribe()
        // Only initialize on mount
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    return (
        <div className="w-100">
            {blocks.map((block, blockIndex) => (
                <div key={block.id}>
                    <SearchNotebookAddBlockButtons onAddBlock={onAddBlock} index={blockIndex} />
                    <>
                        {block.type === 'md' && (
                            <SearchNotebookMarkdownBlock
                                {...props}
                                {...block}
                                isSelected={selectedBlockId === block.id}
                                onSelectBlock={onSelectBlock}
                                onRunBlock={onRunBlock}
                                onBlockInputChange={onBlockInputChange}
                                onMoveBlockSelection={onMoveBlockSelection}
                                onDeleteBlock={onDeleteBlock}
                            />
                        )}
                        {block.type === 'query' && (
                            <SearchNotebookQueryBlock
                                {...props}
                                {...block}
                                isSelected={selectedBlockId === block.id}
                                onSelectBlock={onSelectBlock}
                                onRunBlock={onRunBlock}
                                onBlockInputChange={onBlockInputChange}
                                onMoveBlockSelection={onMoveBlockSelection}
                                onDeleteBlock={onDeleteBlock}
                            />
                        )}
                    </>
                </div>
            ))}
            <SearchNotebookAddBlockButtons
                onAddBlock={onAddBlock}
                index={blocks.length}
                className="mt-4"
                alwaysVisible={true}
            />
        </div>
    )
}
