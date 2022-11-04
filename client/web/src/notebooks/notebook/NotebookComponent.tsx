import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { mdiPlayCircleOutline, mdiDownload, mdiContentCopy } from '@mdi/js'
import classNames from 'classnames'
import { debounce } from 'lodash'
import { useLocation } from 'react-router'
import { Redirect } from 'react-router-dom'
import { Observable, ReplaySubject } from 'rxjs'
import { catchError, delay, filter, map, startWith, switchMap, tap, withLatestFrom } from 'rxjs/operators'

import { HoverMerged } from '@sourcegraph/client-api'
import { createHoverifier } from '@sourcegraph/codeintellify'
import { asError, isDefined, isErrorLike, property } from '@sourcegraph/common'
import { StreamingSearchResultsListProps } from '@sourcegraph/search-ui'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { Controller as ExtensionsController } from '@sourcegraph/shared/src/extensions/controller'
import { getHoverActions } from '@sourcegraph/shared/src/hover/actions'
import { HoverContext } from '@sourcegraph/shared/src/hover/HoverOverlay'
import { getModeFromPath } from '@sourcegraph/shared/src/languages'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, useEventObservable, Icon, useObservable } from '@sourcegraph/wildcard'

import { Block, BlockDirection, BlockInit, BlockInput, BlockType } from '..'
import { AuthenticatedUser } from '../../auth'
import { getHover, getDocumentHighlights } from '../../backend/features'
import { WebHoverOverlay } from '../../components/WebHoverOverlay'
import { NotebookFields } from '../../graphql-operations'
import { getLSPTextDocumentPositionParameters } from '../../repo/blob/Blob'
import { PageRoutes } from '../../routes.constants'
import { SearchStreamingProps } from '../../search'
import { useExperimentalFeatures } from '../../stores'
import { NotebookComputeBlock } from '../blocks/compute/NotebookComputeBlock'
import { NotebookFileBlock } from '../blocks/file/NotebookFileBlock'
import { NotebookMarkdownBlock } from '../blocks/markdown/NotebookMarkdownBlock'
import { NotebookQueryBlock } from '../blocks/query/NotebookQueryBlock'
import { NotebookSymbolBlock } from '../blocks/symbol/NotebookSymbolBlock'

import { NotebookBlockSeparator } from './NotebookBlockSeparator'
import { NotebookCommandPaletteInput } from './NotebookCommandPaletteInput'
import { NotebookOutline } from './NotebookOutline'
import { focusBlockElement, useNotebookEventHandlers } from './useNotebookEventHandlers'

import { Notebook, CopyNotebookProps } from '.'

import styles from './NotebookComponent.module.scss'

export interface NotebookComponentProps
    extends SearchStreamingProps,
        ThemeProps,
        TelemetryProps,
        Omit<StreamingSearchResultsListProps, 'location' | 'allExpanded' | 'executedQuery'> {
    globbing: boolean
    isReadOnly?: boolean
    blocks: BlockInit[]
    authenticatedUser: AuthenticatedUser | null
    extensionsController: Pick<ExtensionsController, 'extHostAPI' | 'executeCommand'> | null
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
        extensionsController,
        exportedFileName,
        isEmbedded,
        authenticatedUser,
        isLightTheme,
        telemetryService,
        isSourcegraphDotCom,
        platformContext,
        blocks: initialBlocks,
        fetchHighlightedFileLineRanges,
        globbing,
        searchContextsEnabled,
        settingsCascade,
        outlineContainerElement,
    }) => {
        const enableGoImportsSearchQueryTransform = useExperimentalFeatures(
            features => features.enableGoImportsSearchQueryTransform
        )
        const notebook = useMemo(
            () =>
                new Notebook(initialBlocks, {
                    extensionHostAPI: extensionsController !== null ? extensionsController.extHostAPI : null,
                    fetchHighlightedFileLineRanges,
                    enableGoImportsSearchQueryTransform,
                }),
            [initialBlocks, fetchHighlightedFileLineRanges, extensionsController, enableGoImportsSearchQueryTransform]
        )

        const notebookElement = useRef<HTMLDivElement | null>(null)
        const [selectedBlockId, setSelectedBlockId] = useState<string | null>(null)
        const [blocks, setBlocks] = useState<Block[]>(notebook.getBlocks())
        const commandPaletteInputReference = useRef<HTMLInputElement>(null)
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

        const focusBlock = useCallback((blockId: string) => focusBlockElement(blockId, isReadOnly), [isReadOnly])

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
                updateBlocks()

                telemetryService.log('SearchNotebookAddBlock', { type: addedBlock.type }, { type: addedBlock.type })
            },
            [notebook, isReadOnly, telemetryService, updateBlocks, selectBlock]
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
                isReadOnly,
                selectBlock,
                onMoveBlock,
                onRunBlock,
                onDeleteBlock,
                onDuplicateBlock,
            }),
            [
                notebook,
                onDeleteBlock,
                onDuplicateBlock,
                onMoveBlock,
                onRunBlock,
                selectedBlockId,
                selectBlock,
                isReadOnly,
            ]
        )
        useNotebookEventHandlers(notebookEventHandlersProps)

        // Element reference subjects passed to `hoverifier`
        const notebookElements = useMemo(() => new ReplaySubject<HTMLElement | null>(1), [])
        useEffect(() => notebookElements.next(notebookElement.current), [notebookElement, notebookElements])
        const hoverOverlayElements = useMemo(() => new ReplaySubject<HTMLElement | null>(1), [])
        const nextOverlayElement = useCallback(
            (overlayElement: HTMLElement | null) => hoverOverlayElements.next(overlayElement),
            [hoverOverlayElements]
        )

        // Subject that emits on every render. Source for `hoverOverlayRerenders`, used to
        // reposition hover overlay if needed when `SearchNotebook` rerenders
        const rerenders = useMemo(() => new ReplaySubject(1), [])
        useEffect(() => rerenders.next())

        // Create hoverifier.
        const hoverifier = useMemo(
            () =>
                createHoverifier<HoverContext, HoverMerged, ActionItemAction>({
                    hoverOverlayElements,
                    hoverOverlayRerenders: rerenders.pipe(
                        withLatestFrom(hoverOverlayElements, notebookElements),
                        map(([, hoverOverlayElement, blockElement]) => ({
                            hoverOverlayElement,
                            relativeElement: blockElement,
                        })),
                        filter(property('relativeElement', isDefined)),
                        // Can't reposition HoverOverlay if it wasn't rendered
                        filter(property('hoverOverlayElement', isDefined))
                    ),
                    getHover: context =>
                        getHover(getLSPTextDocumentPositionParameters(context, getModeFromPath(context.filePath)), {
                            extensionsController,
                        }),
                    getDocumentHighlights: context =>
                        getDocumentHighlights(
                            getLSPTextDocumentPositionParameters(context, getModeFromPath(context.filePath)),
                            { extensionsController }
                        ),
                    getActions: context => getHoverActions({ extensionsController, platformContext }, context),
                    tokenize: false,
                }),
            [
                // None of these dependencies are likely to change
                extensionsController,
                platformContext,
                hoverOverlayElements,
                notebookElements,
                rerenders,
            ]
        )

        // Passed to HoverOverlay
        const hoverState = useObservable(hoverifier.hoverStateUpdates) || {}

        // Dispose hoverifier or change/unmount.
        useEffect(() => () => hoverifier.unsubscribe(), [hoverifier])

        const renderBlock = useCallback(
            (block: Block) => {
                const blockProps = {
                    onRunBlock,
                    onBlockInputChange,
                    onDeleteBlock,
                    onMoveBlock,
                    onDuplicateBlock,
                    isLightTheme,
                    isReadOnly,
                    isSelected: selectedBlockId === block.id,
                    isOtherBlockSelected: selectedBlockId !== null && selectedBlockId !== block.id,
                }

                switch (block.type) {
                    case 'md':
                        return <NotebookMarkdownBlock {...block} {...blockProps} isEmbedded={isEmbedded} />
                    case 'file':
                        return (
                            <NotebookFileBlock
                                {...block}
                                {...blockProps}
                                hoverifier={hoverifier}
                                extensionsController={extensionsController}
                                telemetryService={telemetryService}
                                isSourcegraphDotCom={isSourcegraphDotCom}
                                globbing={globbing}
                            />
                        )
                    case 'query':
                        return (
                            <NotebookQueryBlock
                                {...block}
                                {...blockProps}
                                isSourcegraphDotCom={isSourcegraphDotCom}
                                globbing={globbing}
                                fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                                searchContextsEnabled={searchContextsEnabled}
                                settingsCascade={settingsCascade}
                                telemetryService={telemetryService}
                                platformContext={platformContext}
                                authenticatedUser={authenticatedUser}
                                hoverifier={hoverifier}
                                extensionsController={extensionsController}
                            />
                        )
                    case 'compute':
                        return <NotebookComputeBlock platformContext={platformContext} {...block} {...blockProps} />
                    case 'symbol':
                        return (
                            <NotebookSymbolBlock
                                {...block}
                                {...blockProps}
                                isSourcegraphDotCom={isSourcegraphDotCom}
                                globbing={globbing}
                                telemetryService={telemetryService}
                                platformContext={platformContext}
                                hoverifier={hoverifier}
                                extensionsController={extensionsController}
                            />
                        )
                }
            },
            [
                onRunBlock,
                onBlockInputChange,
                onDeleteBlock,
                onMoveBlock,
                onDuplicateBlock,
                isEmbedded,
                isLightTheme,
                isReadOnly,
                selectedBlockId,
                hoverifier,
                extensionsController,
                telemetryService,
                isSourcegraphDotCom,
                globbing,
                fetchHighlightedFileLineRanges,
                searchContextsEnabled,
                settingsCascade,
                platformContext,
                authenticatedUser,
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
            return <Redirect to={PageRoutes.Notebook.replace(':id', copiedNotebookOrError.id)} />
        }

        return (
            <div
                className={classNames(styles.searchNotebook, isReadOnly && 'is-read-only-notebook')}
                ref={notebookElement}
            >
                <div className="pb-1 px-3">
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
                        <NotebookBlockSeparator isReadOnly={isReadOnly} index={blockIndex} onAddBlock={onAddBlock} />
                        {renderBlock(block)}
                    </div>
                ))}
                {!isReadOnly && (
                    <NotebookCommandPaletteInput
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
                {hoverState.hoverOverlayProps && extensionsController !== null && (
                    <WebHoverOverlay
                        {...hoverState.hoverOverlayProps}
                        platformContext={platformContext}
                        settingsCascade={settingsCascade}
                        hoveredTokenElement={hoverState.hoveredTokenElement}
                        hoverRef={nextOverlayElement}
                        extensionsController={extensionsController}
                        location={location}
                        telemetryService={telemetryService}
                        isLightTheme={isLightTheme}
                    />
                )}
            </div>
        )
    }
)
