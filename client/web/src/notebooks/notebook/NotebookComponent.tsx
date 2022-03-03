import { noop } from 'lodash'
import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import DownloadIcon from 'mdi-react/DownloadIcon'
import PlayCircleOutlineIcon from 'mdi-react/PlayCircleOutlineIcon'
import * as Monaco from 'monaco-editor'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useLocation } from 'react-router'
import { Redirect } from 'react-router-dom'
import { Observable, ReplaySubject } from 'rxjs'
import { catchError, delay, filter, map, startWith, switchMap, tap, withLatestFrom } from 'rxjs/operators'

import { createHoverifier } from '@sourcegraph/codeintellify'
import { asError, isDefined, isErrorLike, property } from '@sourcegraph/common'
import { StreamingSearchResultsListProps } from '@sourcegraph/search-ui'
import { useQueryIntelligence } from '@sourcegraph/search/src/useQueryIntelligence'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { HoverMerged } from '@sourcegraph/shared/src/api/client/types/hover'
import { Controller as ExtensionsController } from '@sourcegraph/shared/src/extensions/controller'
import { getHoverActions } from '@sourcegraph/shared/src/hover/actions'
import { HoverContext } from '@sourcegraph/shared/src/hover/HoverOverlay'
import { getModeFromPath } from '@sourcegraph/shared/src/languages'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, useEventObservable, useObservable } from '@sourcegraph/wildcard'

import { Block, BlockDirection, BlockInit, BlockInput, BlockType } from '..'
import { AuthenticatedUser } from '../../auth'
import { getHover, getDocumentHighlights } from '../../backend/features'
import { WebHoverOverlay } from '../../components/WebHoverOverlay'
import { NotebookFields } from '../../graphql-operations'
import { getLSPTextDocumentPositionParameters } from '../../repo/blob/Blob'
import { PageRoutes } from '../../routes.constants'
import { SearchStreamingProps } from '../../search'
import { NotebookFileBlock } from '../blocks/file/NotebookFileBlock'
import { FileBlockValidationFunctions } from '../blocks/file/useFileBlockInputValidation'
import { NotebookMarkdownBlock } from '../blocks/markdown/NotebookMarkdownBlock'
import { NotebookQueryBlock } from '../blocks/query/NotebookQueryBlock'
import { isMonacoEditorDescendant } from '../blocks/useBlockSelection'

import { NotebookAddBlockButtons } from './NotebookAddBlockButtons'
import styles from './NotebookComponent.module.scss'

import { Notebook, CopyNotebookProps } from '.'

export interface NotebookComponentProps
    extends SearchStreamingProps,
        ThemeProps,
        TelemetryProps,
        Omit<StreamingSearchResultsListProps, 'location' | 'allExpanded'>,
        FileBlockValidationFunctions {
    globbing: boolean
    isMacPlatform: boolean
    isReadOnly?: boolean
    blocks: BlockInit[]
    authenticatedUser: AuthenticatedUser | null
    extensionsController: Pick<ExtensionsController, 'extHostAPI' | 'executeCommand'>
    platformContext: Pick<PlatformContext, 'requestGraphQL' | 'urlToFile' | 'settings' | 'forceUpdateTooltip'>
    exportedFileName: string
    isEmbedded?: boolean
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

export const NotebookComponent: React.FunctionComponent<NotebookComponentProps> = ({
    onSerializeBlocks,
    onCopyNotebook,
    isReadOnly = false,
    extensionsController,
    exportedFileName,
    isEmbedded,
    authenticatedUser,
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
            updateBlocks(false)

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
                        updateBlocks(false)
                        props.telemetryService.log('SearchNotebookRunAllBlocks')
                    })
                ),
            [notebook, props.telemetryService, updateBlocks]
        )
    )

    const exportNotebook = useCallback(() => {
        const exportedMarkdown = notebook.exportToMarkdown(window.location.origin)
        downloadTextAsFile(exportedMarkdown, exportedFileName)
        props.telemetryService.log('SearchNotebookExportNotebook')
    }, [notebook, exportedFileName, props.telemetryService])

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
        props.telemetryService.log('SearchNotebookCopyNotebookButtonClick')
        copyNotebook({ namespace: authenticatedUser.id, blocks: notebook.getBlocks() })
    }, [authenticatedUser, copyNotebook, notebook, props.telemetryService])

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

    // Element reference subjects passed to `hoverifier`
    const notebookElements = useMemo(() => new ReplaySubject<HTMLElement | null>(1), [])
    const nextNotebookElement = useCallback((blockElement: HTMLElement | null) => notebookElements.next(blockElement), [
        notebookElements,
    ])

    const hoverOverlayElements = useMemo(() => new ReplaySubject<HTMLElement | null>(1), [])
    const nextOverlayElement = useCallback(
        (overlayElement: HTMLElement | null) => hoverOverlayElements.next(overlayElement),
        [hoverOverlayElements]
    )

    // Subject that emits on every render. Source for `hoverOverlayRerenders`, used to
    // reposition hover overlay if needed when `SearchNotebook` rerenders
    const rerenders = useMemo(() => new ReplaySubject(1), [])
    useEffect(() => {
        rerenders.next()
    })

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
                getActions: context =>
                    getHoverActions({ extensionsController, platformContext: props.platformContext }, context),
                tokenize: false,
            }),
        [
            // None of these dependencies are likely to change
            extensionsController,
            props.platformContext,
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
                    return <NotebookMarkdownBlock {...block} {...blockProps} />
                case 'file':
                    return (
                        <NotebookFileBlock
                            {...block}
                            {...blockProps}
                            hoverifier={hoverifier}
                            extensionsController={extensionsController}
                        />
                    )
                case 'query':
                    return (
                        <NotebookQueryBlock
                            {...block}
                            {...blockProps}
                            authenticatedUser={authenticatedUser}
                            hoverifier={hoverifier}
                            sourcegraphSearchLanguageId={sourcegraphSearchLanguageId}
                            extensionsController={extensionsController}
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
            extensionsController,
            hoverifier,
            authenticatedUser,
        ]
    )

    const location = useLocation()

    if (copiedNotebookOrError && !isErrorLike(copiedNotebookOrError) && copiedNotebookOrError !== LOADING) {
        return <Redirect to={PageRoutes.Notebook.replace(':id', copiedNotebookOrError.id)} />
    }

    return (
        <div className={styles.searchNotebook} ref={nextNotebookElement}>
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
                {!isEmbedded && (
                    <Button
                        className="mr-2"
                        variant="secondary"
                        size="sm"
                        onClick={exportNotebook}
                        data-testid="export-notebook-markdown-button"
                    >
                        <DownloadIcon className="icon-inline mr-1" />
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
                        <ContentCopyIcon className="icon-inline mr-1" />
                        <span>{copiedNotebookOrError === LOADING ? 'Copying...' : 'Copy to My Notebooks'}</span>
                    </Button>
                )}
            </div>
            {blocks.map((block, blockIndex) => (
                <div key={block.id}>
                    {!isReadOnly ? (
                        <NotebookAddBlockButtons onAddBlock={onAddBlock} index={blockIndex} />
                    ) : (
                        <div className="mb-2" />
                    )}
                    {renderBlock(block)}
                </div>
            ))}
            {!isReadOnly && (
                <NotebookAddBlockButtons
                    onAddBlock={onAddBlock}
                    index={blocks.length}
                    className="mt-2"
                    alwaysVisible={true}
                />
            )}
            {hoverState.hoverOverlayProps && (
                <WebHoverOverlay
                    {...props}
                    {...hoverState.hoverOverlayProps}
                    hoveredTokenElement={hoverState.hoveredTokenElement}
                    hoverRef={nextOverlayElement}
                    extensionsController={extensionsController}
                    location={location}
                    telemetryService={props.telemetryService}
                    isLightTheme={props.isLightTheme}
                />
            )}
        </div>
    )
}
