import * as Monaco from 'monaco-editor'
import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql/schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { SearchStreamingProps } from '..'
import { fetchSuggestions } from '../backend'
import { StreamingSearchResultsListProps } from '../results/StreamingSearchResultsList'
import { useQueryIntelligence } from '../useQueryIntelligence'

import styles from './SearchNotebook.module.scss'
import { SearchNotebookAddBlockButtons } from './SearchNotebookAddBlockButtons'
import { SearchNotebookMarkdownBlock } from './SearchNotebookMarkdownBlock'
import { SearchNotebookQueryBlock } from './SearchNotebookQueryBlock'
import { isMonacoEditorDescendant } from './useBlockSelection'

import { Block, BlockDirection, BlockInitializer, BlockType, Notebook } from '.'

interface SearchNotebookProps
    extends SearchStreamingProps,
        ThemeProps,
        TelemetryProps,
        Omit<StreamingSearchResultsListProps, 'location' | 'allExpanded'> {
    globbing: boolean
    isMacPlatform: boolean

    onSerializeBlocks: (blocks: Block[]) => void
    blocks: BlockInitializer[]
}

export const SearchNotebook: React.FunctionComponent<SearchNotebookProps> = ({ onSerializeBlocks, ...props }) => {
    const notebook = useMemo(() => new Notebook(props.blocks), [props.blocks])

    const [selectedBlockId, setSelectedBlockId] = useState<string | null>(null)
    const [blocks, setBlocks] = useState<Block[]>(notebook.getBlocks())

    const updateBlocks = useCallback(
        (serialize = true) => {
            const blocks = notebook.getBlocks()
            setBlocks(blocks)
            if (serialize) {
                onSerializeBlocks(blocks)
            }
        },
        [notebook, setBlocks, onSerializeBlocks]
    )

    const onRunBlock = useCallback(
        (id: string) => {
            notebook.runBlockById(id)
            updateBlocks()

            props.telemetryService.log('SearchNotebookRunBlock', { type: notebook.getBlockById(id)?.type })
        },
        [notebook, props.telemetryService, updateBlocks]
    )

    const onBlockInputChange = useCallback(
        (id: string, value: string) => {
            notebook.setBlockInputById(id, value)
            updateBlocks(false)
        },
        [notebook, updateBlocks]
    )

    const onAddBlock = useCallback(
        (index: number, type: BlockType, input: string) => {
            const addedBlock = notebook.insertBlockAtIndex(index, type, input)
            if (addedBlock.type === 'md') {
                notebook.runBlockById(addedBlock.id)
            }
            setSelectedBlockId(addedBlock.id)
            updateBlocks()

            props.telemetryService.log('SearchNotebookAddBlock', { type: addedBlock.type })
        },
        [notebook, props.telemetryService, updateBlocks, setSelectedBlockId]
    )

    const onDeleteBlock = useCallback(
        (id: string) => {
            const block = notebook.getBlockById(id)
            const blockToFocusAfterDelete = notebook.getNextBlockId(id) ?? notebook.getPreviousBlockId(id)
            notebook.deleteBlockById(id)
            setSelectedBlockId(blockToFocusAfterDelete)
            updateBlocks()

            props.telemetryService.log('SearchNotebookDeleteBlock', { type: block?.type })
        },
        [notebook, props.telemetryService, setSelectedBlockId, updateBlocks]
    )

    const onMoveBlock = useCallback(
        (id: string, direction: BlockDirection) => {
            notebook.moveBlockById(id, direction)
            updateBlocks()

            props.telemetryService.log('SearchNotebookMoveBlock', { type: notebook.getBlockById(id)?.type, direction })
        },
        [notebook, props.telemetryService, updateBlocks]
    )

    const onDuplicateBlock = useCallback(
        (id: string) => {
            const duplicateBlock = notebook.duplicateBlockById(id)
            if (duplicateBlock) {
                setSelectedBlockId(duplicateBlock.id)
            }
            if (duplicateBlock?.type === 'md') {
                notebook.runBlockById(duplicateBlock.id)
            }
            updateBlocks()

            props.telemetryService.log('SearchNotebookDuplicateBlock', { type: duplicateBlock?.type })
        },
        [notebook, props.telemetryService, setSelectedBlockId, updateBlocks]
    )

    const onSelectBlock = useCallback(
        (id: string) => {
            setSelectedBlockId(id)
        },
        [setSelectedBlockId]
    )

    const onMoveBlockSelection = useCallback(
        (id: string, direction: BlockDirection) => {
            const blockId = direction === 'up' ? notebook.getPreviousBlockId(id) : notebook.getNextBlockId(id)
            if (blockId) {
                setSelectedBlockId(blockId)
            }
        },
        [notebook, setSelectedBlockId]
    )

    useEffect(() => {
        const handleEventOutsideBlockWrapper = (event: MouseEvent | FocusEvent): void => {
            const target = event.target as HTMLElement | null
            if (target && !target.closest('.block-wrapper')) {
                setSelectedBlockId(null)
            }
        }
        const handleKeyDown = (event: KeyboardEvent): void => {
            if (!selectedBlockId && event.key === 'ArrowDown') {
                setSelectedBlockId(notebook.getFirstBlockId())
            } else if (event.key === 'Escape' && !isMonacoEditorDescendant(event.target as HTMLElement)) {
                setSelectedBlockId(null)
            }
        }

        document.addEventListener('keydown', handleKeyDown)
        // Check all clicks on the document and deselect the currently selected block if it was triggered outside of a block.
        document.addEventListener('mousedown', handleEventOutsideBlockWrapper)
        // We're using the `focusin` event instead of the `focus` event, since the latter does not bubble up.
        document.addEventListener('focusin', handleEventOutsideBlockWrapper)
        return () => {
            document.removeEventListener('keydown', handleKeyDown)
            document.removeEventListener('mousedown', handleEventOutsideBlockWrapper)
            document.removeEventListener('focusin', handleEventOutsideBlockWrapper)
        }
    }, [notebook, selectedBlockId, onMoveBlockSelection, setSelectedBlockId])

    useQueryIntelligence(fetchSuggestions, {
        patternType: SearchPatternType.literal,
        globbing: props.globbing,
        interpretComments: true,
    })

    // Register dummy onCompletionSelected handler to prevent console errors
    useEffect(() => {
        const disposable = Monaco.editor.registerCommand('completionItemSelected', () => {})
        return () => disposable.dispose()
    }, [])

    const blockCallbackProps = {
        onSelectBlock,
        onRunBlock,
        onBlockInputChange,
        onMoveBlockSelection,
        onDeleteBlock,
        onMoveBlock,
        onDuplicateBlock,
    }

    return (
        <div className={styles.searchNotebook}>
            {blocks.map((block, blockIndex) => (
                <div key={block.id}>
                    <SearchNotebookAddBlockButtons onAddBlock={onAddBlock} index={blockIndex} />
                    <>
                        {block.type === 'md' && (
                            <SearchNotebookMarkdownBlock
                                {...props}
                                {...block}
                                {...blockCallbackProps}
                                isSelected={selectedBlockId === block.id}
                                isOtherBlockSelected={selectedBlockId !== null && selectedBlockId !== block.id}
                            />
                        )}
                        {block.type === 'query' && (
                            <SearchNotebookQueryBlock
                                {...props}
                                {...block}
                                {...blockCallbackProps}
                                isSelected={selectedBlockId === block.id}
                                isOtherBlockSelected={selectedBlockId !== null && selectedBlockId !== block.id}
                            />
                        )}
                    </>
                </div>
            ))}
            <SearchNotebookAddBlockButtons
                onAddBlock={onAddBlock}
                index={blocks.length}
                className="mt-2"
                alwaysVisible={true}
            />
        </div>
    )
}
