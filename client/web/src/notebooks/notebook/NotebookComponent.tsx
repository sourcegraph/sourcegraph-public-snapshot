import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { mdiPlayCircleOutline, mdiDownload, mdiContentCopy } from '@mdi/js'
import classNames from 'classnames'
import { debounce } from 'lodash'
import { Navigate, useLocation } from 'react-router-dom'
import type { Observable } from 'rxjs'
import { catchError, delay, startWith, switchMap, tap } from 'rxjs/operators'

import type { StreamingSearchResultsListProps } from '@sourcegraph/branded'
import { asError, isErrorLike } from '@sourcegraph/common'
import type { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, useEventObservable, Icon } from '@sourcegraph/wildcard'

import type { Block, BlockDirection, BlockInit, BlockInput, BlockType } from '..'
import type { AuthenticatedUser } from '../../auth'
import type { NotebookFields } from '../../graphql-operations'
import type { OwnConfigProps } from '../../own/OwnConfigProps'
import { EnterprisePageRoutes } from '../../routes.constants'
import type { SearchStreamingProps } from '../../search'
import { NotebookFileBlock } from '../blocks/file/NotebookFileBlock'
import { NotebookMarkdownBlock } from '../blocks/markdown/NotebookMarkdownBlock'
import { NotebookQueryBlock } from '../blocks/query/NotebookQueryBlock'
import { NotebookSymbolBlock } from '../blocks/symbol/NotebookSymbolBlock'

import { Notebook, type CopyNotebookProps } from '.'
import { NotebookCommandPaletteInput } from './NotebookCommandPaletteInput'
import { NotebookOutline } from './NotebookOutline'
import { focusBlockElement, useNotebookEventHandlers } from './useNotebookEventHandlers'

import styles from './NotebookComponent.module.scss'

export interface NotebookComponentProps
    extends SearchStreamingProps,
        TelemetryProps,
        Omit<StreamingSearchResultsListProps, 'location' | 'allExpanded' | 'executedQuery' | 'enableOwnershipSearch'>,
        OwnConfigProps {
    isReadOnly?: boolean
    blocks: BlockInit[]
    authenticatedUser: AuthenticatedUser | null
    platformContext: Pick<PlatformContext, 'sourcegraphURL' | 'requestGraphQL' | 'urlToFile' | 'settings'>
    exportedFileName: string
    isEmbedded?: boolean
    outlineContainerElement?: HTMLElement | null
    onSerializeBlocks: (blocks: Block[]) => void
    onCopyNotebook: (props: Omit<CopyNotebookProps, 'title'>) => Observable<NotebookFields>
}

const LOADING = 'LOADING' as const

type BlockCounts = { [blockType in BlockType]: number }

function countBlockTypes(blocks: Block[]): BlockCounts {
    return blocks.reduce((aggregate, block) => ({ ...aggregate, [block.type]: aggregate[block.type] + 1 }), {
        md: 0,
        file: 0,
        query: 0,
        compute: 0,
        symbol: 0,
    })
}

function downloadTextAsFile(text: string, fileName: string): void {
    const blob = new Blob([text], { type: 'text/plain' })
    const blobURL = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.style.display = 'none'
    link.href = blobURL
    link.download = fileName
    document.body.append(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(blobURL)
}

export const NotebookComponent: React.FunctionComponent<React.PropsWithChildren<NotebookComponentProps>> = React.memo(
    ({
        onSerializeBlocks,
        onCopyNotebook,
        isReadOnly = false,
        exportedFileName,
        isEmbedded,
        authenticatedUser,
        telemetryService,
        telemetryRecorder,
        isSourcegraphDotCom,
        platformContext,
        blocks: initialBlocks,
        fetchHighlightedFileLineRanges,
        searchContextsEnabled,
        ownEnabled,
        settingsCascade,
        outlineContainerElement,
    }) => {
        const notebook = useMemo(
            () =>
                new Notebook(initialBlocks, {
                    fetchHighlightedFileLineRanges,
                }),
            [initialBlocks, fetchHighlightedFileLineRanges]
        )

        const notebookElement = useRef<HTMLDivElement | null>(null)
        const [selectedBlockId, setSelectedBlockId] = useState<string | null>(null)
        const [blockInserterIndex, setBlockInserterIndex] = useState<number>(-1)

        const [blocks, setBlocks] = useState<Block[]>(notebook.getBlocks())
        const commandPaletteInputReference = useRef<HTMLInputElement>(null)
        const floatingCommandPaletteInputReference = useRef<HTMLInputElement>(null)
        const debouncedOnSerializeBlocks = useMemo(() => debounce(onSerializeBlocks, 400), [onSerializeBlocks])

        const updateBlocks = useCallback(
            (serialize = true) => {
                const blocks = notebook.getBlocks()
                setBlocks(blocks)
                if (serialize) {
                    debouncedOnSerializeBlocks(blocks)
                }
                const blockCountsPerType = countBlockTypes(blocks)
                telemetryService.log(
                    'SearchNotebookBlocksUpdated',
                    { blocksCount: blocks.length, blockCountsPerType },
                    { blocksCount: blocks.length, blockCountsPerType }
                )
            },
            [notebook, setBlocks, debouncedOnSerializeBlocks, telemetryService]
        )

        const selectBlock = useCallback(
            (blockId: string | null) => {
                if (!isReadOnly) {
                    setSelectedBlockId(blockId)
                }
            },
            [isReadOnly, setSelectedBlockId]
        )

        const focusBlock = useCallback(
            (blockId: string) => {
                focusBlockElement(blockId, isReadOnly)
            },
            [isReadOnly]
        )

        // Update the blocks if the notebook instance changes (when new initializer blocks are provided)
        useEffect(() => setBlocks(notebook.getBlocks()), [notebook])

        const onRunBlock = useCallback(
            (id: string) => {
                notebook.runBlockById(id)
                updateBlocks(false)

                telemetryService.log(
                    'SearchNotebookRunBlock',
                    { type: notebook.getBlockById(id)?.type },
                    { type: notebook.getBlockById(id)?.type }
                )
            },
            [notebook, telemetryService, updateBlocks]
        )

        const [runAllBlocks, runningAllBlocks] = useEventObservable(
            useCallback(
                (click: Observable<React.MouseEvent>) =>
                    click.pipe(
                        switchMap(() => notebook.runAllBlocks().pipe(startWith(LOADING))),
                        tap(() => {
                            updateBlocks(false)
                            telemetryService.log('SearchNotebookRunAllBlocks')
                        })
                    ),
                [notebook, telemetryService, updateBlocks]
            )
        )

        const [exportNotebook] = useEventObservable(
            useCallback(
                (event: Observable<React.MouseEvent<HTMLButtonElement>>) =>
                    event.pipe(
                        switchMap(() => notebook.exportToMarkdown(window.location.origin)),
                        tap(exportedMarkdown => {
                            downloadTextAsFile(exportedMarkdown, exportedFileName)
                            telemetryService.log('SearchNotebookExportNotebook')
                        })
                    ),
                [notebook, exportedFileName, telemetryService]
            )
        )

        const [copyNotebook, copiedNotebookOrError] = useEventObservable(
            useCallback(
                (input: Observable<Omit<CopyNotebookProps, 'title'>>) =>
                    input.pipe(
                        // The delay is added to make the copy loading state more obvious. Otherwise, the copy
                        // happens too fast, and it can seem like nothing happened.
                        switchMap(props => onCopyNotebook(props).pipe(delay(400), startWith(LOADING))),
                        catchError(error => [asError(error)])
                    ),
                [onCopyNotebook]
            )
        )

        const onCopyNotebookButtonClick = useCallback(() => {
            if (!authenticatedUser) {
                return
            }
            telemetryService.log('SearchNotebookCopyNotebookButtonClick')
            copyNotebook({ namespace: authenticatedUser.id, blocks: notebook.getBlocks() })
        }, [authenticatedUser, copyNotebook, notebook, telemetryService])

        const onBlockInputChange = useCallback(
            (id: string, blockInput: BlockInput) => {
                notebook.setBlockInputById(id, blockInput)
                updateBlocks()
            },
            [notebook, updateBlocks]
        )

        const onNewBlock = useCallback(
            (id: string) => {
                if (isReadOnly) {
                    return
                }
                const idx = notebook.getBlockIndex(id)
                setBlockInserterIndex(idx + 1)
                selectBlock(null)
            },
            [isReadOnly, notebook, selectBlock]
        )

        const dismissNewBlockPalette = useCallback(() => {
            if (isReadOnly) {
                return
            }
            if (blocks.length === 0) {
                return
            }
            const blockToSelectIndex = blockInserterIndex - 1
            selectBlock(blocks[blockToSelectIndex].id)
            setBlockInserterIndex(-1)
        }, [isReadOnly, blocks, selectBlock, blockInserterIndex])

        const onAddBlock = useCallback(
            (index: number, blockInput: BlockInput) => {
                if (isReadOnly) {
                    return
                }
                const addedBlock = notebook.insertBlockAtIndex(index, blockInput)
                if (
                    addedBlock.type === 'md' ||
                    (addedBlock.type === 'file' && addedBlock.input.repositoryName && addedBlock.input.filePath)
                ) {
                    notebook.runBlockById(addedBlock.id)
                }
                selectBlock(addedBlock.id)
                focusBlock(addedBlock.id)
                setBlockInserterIndex(-1)
                updateBlocks()

                telemetryService.log('SearchNotebookAddBlock', { type: addedBlock.type }, { type: addedBlock.type })
            },
            [isReadOnly, notebook, selectBlock, focusBlock, updateBlocks, telemetryService]
        )

        const onDeleteBlock = useCallback(
            (id: string) => {
                if (isReadOnly) {
                    return
                }

                const block = notebook.getBlockById(id)
                const blockToFocusAfterDelete = notebook.getNextBlockId(id) ?? notebook.getPreviousBlockId(id)
                notebook.deleteBlockById(id)
                selectBlock(blockToFocusAfterDelete)
                if (blockToFocusAfterDelete) {
                    focusBlock(blockToFocusAfterDelete)
                }
                updateBlocks()

                telemetryService.log('SearchNotebookDeleteBlock', { type: block?.type }, { type: block?.type })
            },
            [notebook, isReadOnly, telemetryService, selectBlock, updateBlocks, focusBlock]
        )

        const onMoveBlock = useCallback(
            (id: string, direction: BlockDirection) => {
                if (isReadOnly) {
                    return
                }

                notebook.moveBlockById(id, direction)
                focusBlock(id)
                updateBlocks()

                telemetryService.log(
                    'SearchNotebookMoveBlock',
                    { type: notebook.getBlockById(id)?.type, direction },
                    { type: notebook.getBlockById(id)?.type, direction }
                )
            },
            [notebook, isReadOnly, telemetryService, updateBlocks, focusBlock]
        )

        const onDuplicateBlock = useCallback(
            (id: string) => {
                if (isReadOnly) {
                    return
                }

                const duplicateBlock = notebook.duplicateBlockById(id)
                if (duplicateBlock) {
                    selectBlock(duplicateBlock.id)
                    focusBlock(duplicateBlock.id)
                }
                if (duplicateBlock?.type === 'md') {
                    notebook.runBlockById(duplicateBlock.id)
                }
                updateBlocks()

                telemetryService.log(
                    'SearchNotebookDuplicateBlock',
                    { type: duplicateBlock?.type },
                    { type: duplicateBlock?.type }
                )
            },
            [notebook, isReadOnly, telemetryService, selectBlock, updateBlocks, focusBlock]
        )

        const onFocusLastBlock = useCallback(() => {
            const lastBlockId = notebook.getLastBlockId()
            if (lastBlockId) {
                selectBlock(lastBlockId)
                focusBlock(lastBlockId)
            }
        }, [notebook, selectBlock, focusBlock])

        const notebookEventHandlersProps = useMemo(
            () => ({
                notebook,
                selectedBlockId,
                commandPaletteInputReference,
                floatingCommandPaletteInputReference,
                isReadOnly,
                selectBlock,
                onMoveBlock,
                onRunBlock,
                onDeleteBlock,
                onDuplicateBlock,
                onNewBlock,
            }),
            [
                notebook,
                onDeleteBlock,
                onDuplicateBlock,
                onNewBlock,
                onMoveBlock,
                onRunBlock,
                selectedBlockId,
                selectBlock,
                isReadOnly,
            ]
        )
        useNotebookEventHandlers(notebookEventHandlersProps)

        const renderBlock = useCallback(
            (block: Block) => {
                const isSelected = selectedBlockId === block.id
                const isSomethingElseSelected =
                    (selectedBlockId !== null && selectedBlockId !== block.id) || blockInserterIndex !== -1
                const blockProps = {
                    onRunBlock,
                    onBlockInputChange,
                    onDeleteBlock,
                    onNewBlock,
                    onAddBlock,
                    onMoveBlock,
                    onDuplicateBlock,
                    isReadOnly,
                    isSelected,
                    showMenu: isSelected || !isSomethingElseSelected,
                }

                switch (block.type) {
                    case 'md':
                        return <NotebookMarkdownBlock {...block} {...blockProps} isEmbedded={isEmbedded} />
                    case 'file':
                        return (
                            <NotebookFileBlock
                                {...block}
                                {...blockProps}
                                telemetryService={telemetryService}
                                telemetryRecorder={telemetryRecorder}
                                isSourcegraphDotCom={isSourcegraphDotCom}
                            />
                        )
                    case 'query':
                        return (
                            <NotebookQueryBlock
                                {...block}
                                {...blockProps}
                                isSourcegraphDotCom={isSourcegraphDotCom}
                                fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                                searchContextsEnabled={searchContextsEnabled}
                                ownEnabled={ownEnabled}
                                settingsCascade={settingsCascade}
                                telemetryService={telemetryService}
                                telemetryRecorder={telemetryRecorder}
                                platformContext={platformContext}
                                authenticatedUser={authenticatedUser}
                            />
                        )
                    case 'symbol':
                        return (
                            <NotebookSymbolBlock
                                {...block}
                                {...blockProps}
                                isSourcegraphDotCom={isSourcegraphDotCom}
                                telemetryService={telemetryService}
                                telemetryRecorder={telemetryRecorder}
                                platformContext={platformContext}
                            />
                        )
                }
            },
            [
                selectedBlockId,
                blockInserterIndex,
                onRunBlock,
                onBlockInputChange,
                onDeleteBlock,
                onNewBlock,
                onAddBlock,
                onMoveBlock,
                onDuplicateBlock,
                isReadOnly,
                isEmbedded,
                telemetryService,
                isSourcegraphDotCom,
                fetchHighlightedFileLineRanges,
                searchContextsEnabled,
                ownEnabled,
                settingsCascade,
                platformContext,
                authenticatedUser,
                telemetryRecorder,
            ]
        )

        const location = useLocation()

        useEffect(() => {
            const headingId = location.hash.slice(1)
            if (!headingId) {
                return
            }
            const heading = document.querySelector(`[id="${headingId}"]`)
            heading?.scrollIntoView()
            // Scroll to heading only on mount.
            // eslint-disable-next-line react-hooks/exhaustive-deps
        }, [])

        if (copiedNotebookOrError && !isErrorLike(copiedNotebookOrError) && copiedNotebookOrError !== LOADING) {
            return (
                <Navigate to={EnterprisePageRoutes.Notebook.replace(':id', copiedNotebookOrError.id)} replace={true} />
            )
        }

        return (
            <div
                className={classNames(styles.searchNotebook, isReadOnly && 'is-read-only-notebook')}
                ref={notebookElement}
            >
                <div className={classNames(styles.header, 'pb-1', 'px-3')}>
                    <Button
                        className="mr-2"
                        variant="primary"
                        size="sm"
                        onClick={runAllBlocks}
                        disabled={blocks.length === 0 || runningAllBlocks === LOADING}
                    >
                        <Icon aria-hidden={true} className="mr-1" svgPath={mdiPlayCircleOutline} />
                        <span>{runningAllBlocks === LOADING ? 'Running...' : 'Run all blocks'}</span>
                    </Button>
                    {!isEmbedded && (
                        <Button
                            className="mr-2"
                            variant="secondary"
                            size="sm"
                            onClick={exportNotebook}
                            data-testid="export-notebook-markdown-button"
                        >
                            <Icon aria-hidden={true} className="mr-1" svgPath={mdiDownload} />
                            <span>Export as Markdown</span>
                        </Button>
                    )}
                    {!isEmbedded && authenticatedUser && (
                        <Button
                            className="mr-2"
                            variant="secondary"
                            size="sm"
                            onClick={onCopyNotebookButtonClick}
                            data-testid="copy-notebook-button"
                            disabled={copiedNotebookOrError === LOADING}
                        >
                            <Icon aria-hidden={true} className="mr-1" svgPath={mdiContentCopy} />
                            <span>{copiedNotebookOrError === LOADING ? 'Copying...' : 'Copy to My Notebooks'}</span>
                        </Button>
                    )}
                </div>
                {blocks.map((block, blockIndex) => (
                    <div key={block.id}>
                        {blockInserterIndex === blockIndex && (
                            <NotebookCommandPaletteInput
                                hasFocus={true}
                                ref={floatingCommandPaletteInputReference}
                                index={blockIndex}
                                onAddBlock={onAddBlock}
                                onShouldDismiss={dismissNewBlockPalette}
                            />
                        )}
                        {renderBlock(block)}
                    </div>
                ))}
                {!isReadOnly && (
                    <NotebookCommandPaletteInput
                        hasFocus={blockInserterIndex === blocks.length}
                        ref={commandPaletteInputReference}
                        index={blocks.length}
                        onAddBlock={onAddBlock}
                        onFocusPreviousBlock={onFocusLastBlock}
                    />
                )}
                {notebookElement.current && outlineContainerElement && (
                    <NotebookOutline
                        notebookElement={notebookElement.current}
                        outlineContainerElement={outlineContainerElement}
                        blocks={blocks}
                    />
                )}
            </div>
        )
    }
)

NotebookComponent.displayName = 'NotebookComponent'
