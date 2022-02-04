import { noop } from 'lodash'
import PlayCircleOutlineIcon from 'mdi-react/PlayCircleOutlineIcon'
import * as Monaco from 'monaco-editor'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Observable } from 'rxjs'
import { startWith, switchMap, tap } from 'rxjs/operators'

import { StreamingSearchResultsListProps } from '@sourcegraph/search-ui'
import { useQueryIntelligence } from '@sourcegraph/search/src/useQueryIntelligence'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, useEventObservable } from '@sourcegraph/wildcard'

import { SearchStreamingProps } from '..'
import { AuthenticatedUser } from '../../auth'

import { SearchNotebookFileBlock } from './fileBlock/SearchNotebookFileBlock'
import { FileBlockValidationFunctions } from './fileBlock/useFileBlockInputValidation'
import styles from './SearchNotebook.module.scss'
import { SearchNotebookAddBlockButtons } from './SearchNotebookAddBlockButtons'
import { SearchNotebookMarkdownBlock } from './SearchNotebookMarkdownBlock'
import { SearchNotebookQueryBlock } from './SearchNotebookQueryBlock'
import { isMonacoEditorDescendant } from './useBlockSelection'

import { Block, BlockDirection, BlockInit, BlockInput, BlockType, Notebook } from '.'

export interface SearchNotebookProps
    extends SearchStreamingProps,
        ThemeProps,
        TelemetryProps,
        Omit<StreamingSearchResultsListProps, 'location' | 'allExpanded'>,
        ExtensionsControllerProps<'extHostAPI'>,
        FileBlockValidationFunctions {
    globbing: boolean
    isMacPlatform: boolean
    isReadOnly?: boolean
    onSerializeBlocks: (blocks: Block[]) => void
    blocks: BlockInit[]
    authenticatedUser: AuthenticatedUser | null
}

const LOADING = 'LOADING' as const

type BlockCounts = { [blockType in BlockType]: number }

function countBlockTypes(blocks: Block[]): BlockCounts {
    return blocks.reduce((aggregate, block) => ({ ...aggregate, [block.type]: aggregate[block.type] + 1 }), {
        md: 0,
        file: 0,
        query: 0,
    })
}

export const SearchNotebook: React.FunctionComponent<SearchNotebookProps> = ({
    onSerializeBlocks,
    isReadOnly = false,
    extensionsController,
    ...props
}) => {
    const notebook = useMemo(
        () =>
            new Notebook(props.blocks, {
                extensionHostAPI: extensionsController.extHostAPI,
                fetchHighlightedFileLineRanges: props.fetchHighlightedFileLineRanges,
            }),
        [props.blocks, props.fetchHighlightedFileLineRanges, extensionsController.extHostAPI]
    )

    const [selectedBlockId, setSelectedBlockId] = useState<string | null>(null)
    const [blocks, setBlocks] = useState<Block[]>(notebook.getBlocks())

    const updateBlocks = useCallback(
        (serialize = true) => {
            const blocks = notebook.getBlocks()
            setBlocks(blocks)
            if (serialize) {
                onSerializeBlocks(blocks)
            }
            const blockCountsPerType = countBlockTypes(blocks)
            props.telemetryService.log(
                'SearchNotebookBlocksUpdated',
                { blocksCount: blocks.length, blockCountsPerType },
                { blocksCount: blocks.length, blockCountsPerType }
            )
        },
        [notebook, setBlocks, onSerializeBlocks, props.telemetryService]
    )

    // Update the blocks if the notebook instance changes (when new initializer blocks are provided)
    useEffect(() => setBlocks(notebook.getBlocks()), [notebook])

    const onRunBlock = useCallback(
        (id: string) => {
            notebook.runBlockById(id)
            updateBlocks()

            props.telemetryService.log(
                'SearchNotebookRunBlock',
                { type: notebook.getBlockById(id)?.type },
                { type: notebook.getBlockById(id)?.type }
            )
        },
        [notebook, props.telemetryService, updateBlocks]
    )

    const [runAllBlocks, runningAllBlocks] = useEventObservable(
        useCallback(
            (click: Observable<React.MouseEvent>) =>
                click.pipe(
                    switchMap(() => notebook.runAllBlocks().pipe(startWith(LOADING))),
                    tap(() => {
                        updateBlocks()
                        props.telemetryService.log('SearchNotebookRunAllBlocks')
                    })
                ),
            [notebook, props.telemetryService, updateBlocks]
        )
    )

    const onBlockInputChange = useCallback(
        (id: string, blockInput: BlockInput) => {
            notebook.setBlockInputById(id, blockInput)
            updateBlocks(false)
        },
        [notebook, updateBlocks]
    )

    const onAddBlock = useCallback(
        (index: number, blockInput: BlockInput) => {
            if (isReadOnly) {
                return
            }
            const addedBlock = notebook.insertBlockAtIndex(index, blockInput)
            if (addedBlock.type === 'md') {
                notebook.runBlockById(addedBlock.id)
            }
            setSelectedBlockId(addedBlock.id)
            updateBlocks()

            props.telemetryService.log('SearchNotebookAddBlock', { type: addedBlock.type }, { type: addedBlock.type })
        },
        [notebook, isReadOnly, props.telemetryService, updateBlocks, setSelectedBlockId]
    )

    const onDeleteBlock = useCallback(
        (id: string) => {
            if (isReadOnly) {
                return
            }

            const block = notebook.getBlockById(id)
            const blockToFocusAfterDelete = notebook.getNextBlockId(id) ?? notebook.getPreviousBlockId(id)
            notebook.deleteBlockById(id)
            setSelectedBlockId(blockToFocusAfterDelete)
            updateBlocks()

            props.telemetryService.log('SearchNotebookDeleteBlock', { type: block?.type }, { type: block?.type })
        },
        [notebook, isReadOnly, props.telemetryService, setSelectedBlockId, updateBlocks]
    )

    const onMoveBlock = useCallback(
        (id: string, direction: BlockDirection) => {
            if (isReadOnly) {
                return
            }

            notebook.moveBlockById(id, direction)
            updateBlocks()

            props.telemetryService.log(
                'SearchNotebookMoveBlock',
                { type: notebook.getBlockById(id)?.type, direction },
                { type: notebook.getBlockById(id)?.type, direction }
            )
        },
        [notebook, isReadOnly, props.telemetryService, updateBlocks]
    )

    const onDuplicateBlock = useCallback(
        (id: string) => {
            if (isReadOnly) {
                return
            }

            const duplicateBlock = notebook.duplicateBlockById(id)
            if (duplicateBlock) {
                setSelectedBlockId(duplicateBlock.id)
            }
            if (duplicateBlock?.type === 'md') {
                notebook.runBlockById(duplicateBlock.id)
            }
            updateBlocks()

            props.telemetryService.log(
                'SearchNotebookDuplicateBlock',
                { type: duplicateBlock?.type },
                { type: duplicateBlock?.type }
            )
        },
        [notebook, isReadOnly, props.telemetryService, setSelectedBlockId, updateBlocks]
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
            if (!target?.closest('.block-wrapper') && !target?.closest('[data-reach-combobox-list]')) {
                setSelectedBlockId(null)
            }
        }
        const handleKeyDown = (event: KeyboardEvent): void => {
            const target = event.target as HTMLElement
            if (!selectedBlockId && event.key === 'ArrowDown') {
                setSelectedBlockId(notebook.getFirstBlockId())
            } else if (
                event.key === 'Escape' &&
                !isMonacoEditorDescendant(target) &&
                target.tagName.toLowerCase() !== 'input'
            ) {
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

    const sourcegraphSearchLanguageId = useQueryIntelligence(fetchStreamSuggestions, {
        patternType: SearchPatternType.literal,
        globbing: props.globbing,
        interpretComments: true,
    })

    // Register dummy onCompletionSelected handler to prevent console errors
    useEffect(() => {
        const disposable = Monaco.editor.registerCommand('completionItemSelected', noop)
        return () => disposable.dispose()
    }, [])

    const renderBlock = useCallback(
        (block: Block) => {
            const blockProps = {
                ...props,
                onSelectBlock,
                onRunBlock,
                onBlockInputChange,
                onMoveBlockSelection,
                onDeleteBlock,
                onMoveBlock,
                onDuplicateBlock,
                isReadOnly,
                isSelected: selectedBlockId === block.id,
                isOtherBlockSelected: selectedBlockId !== null && selectedBlockId !== block.id,
            }

            switch (block.type) {
                case 'md':
                    return <SearchNotebookMarkdownBlock {...block} {...blockProps} />
                case 'file':
                    return <SearchNotebookFileBlock {...block} {...blockProps} />
                case 'query':
                    return (
                        <SearchNotebookQueryBlock
                            {...block}
                            {...blockProps}
                            sourcegraphSearchLanguageId={sourcegraphSearchLanguageId}
                        />
                    )
            }
        },
        [
            isReadOnly,
            onBlockInputChange,
            onDeleteBlock,
            onDuplicateBlock,
            onMoveBlock,
            onMoveBlockSelection,
            onRunBlock,
            onSelectBlock,
            props,
            selectedBlockId,
            sourcegraphSearchLanguageId,
        ]
    )

    return (
        <div className={styles.searchNotebook}>
            <div className="pb-1">
                <Button
                    className="mr-2"
                    variant="primary"
                    size="sm"
                    onClick={runAllBlocks}
                    disabled={blocks.length === 0 || runningAllBlocks === LOADING}
                >
                    <PlayCircleOutlineIcon className="icon-inline mr-1" />
                    <span>{runningAllBlocks === LOADING ? 'Running...' : 'Run all blocks'}</span>
                </Button>
            </div>
            {blocks.map((block, blockIndex) => (
                <div key={block.id}>
                    {!isReadOnly ? (
                        <SearchNotebookAddBlockButtons onAddBlock={onAddBlock} index={blockIndex} />
                    ) : (
                        <div className="mb-2" />
                    )}
                    {renderBlock(block)}
                </div>
            ))}
            {!isReadOnly && (
                <SearchNotebookAddBlockButtons
                    onAddBlock={onAddBlock}
                    index={blocks.length}
                    className="mt-2"
                    alwaysVisible={true}
                />
            )}
        </div>
    )
}
