import React, { useState, useCallback, useMemo, useEffect } from 'react'

import classNames from 'classnames'
import { debounce } from 'lodash'
import CheckIcon from 'mdi-react/CheckIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import PencilIcon from 'mdi-react/PencilIcon'
import { of } from 'rxjs'
import { startWith } from 'rxjs/operators'

import { Hoverifier } from '@sourcegraph/codeintellify'
import { isErrorLike } from '@sourcegraph/common'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { HoverMerged } from '@sourcegraph/shared/src/api/client/types/hover'
import { CodeExcerpt } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { HoverContext } from '@sourcegraph/shared/src/hover/HoverOverlay'
import { IHighlightLineRange } from '@sourcegraph/shared/src/schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'
import { useCodeIntelViewerUpdates } from '@sourcegraph/shared/src/util/useCodeIntelViewerUpdates'
import { LoadingSpinner, useObservable, Link, Alert } from '@sourcegraph/wildcard'

import { BlockProps, FileBlock, FileBlockInput } from '../..'
import { parseFileBlockInput, serializeLineRange } from '../../serialize'
import { BlockMenuAction } from '../menu/NotebookBlockMenu'
import { useCommonBlockMenuActions } from '../menu/useCommonBlockMenuActions'
import { NotebookBlock } from '../NotebookBlock'
import { useModifierKeyLabel } from '../useModifierKeyLabel'

import { NotebookFileBlockInputs } from './NotebookFileBlockInputs'

import styles from './NotebookFileBlock.module.scss'

interface NotebookFileBlockProps
    extends BlockProps<FileBlock>,
        TelemetryProps,
        ExtensionsControllerProps<'extHostAPI' | 'executeCommand'>,
        ThemeProps {
    sourcegraphSearchLanguageId: string
    isSourcegraphDotCom: boolean
    hoverifier?: Hoverifier<HoverContext, HoverMerged, ActionItemAction>
}

const LOADING = 'loading' as const

export const NotebookFileBlock: React.FunctionComponent<NotebookFileBlockProps> = ({
    id,
    input,
    output,
    telemetryService,
    isSelected,
    isOtherBlockSelected,
    isReadOnly,
    hoverifier,
    extensionsController,
    onRunBlock,
    onSelectBlock,
    onBlockInputChange,
    ...props
}) => {
    const [showInputs, setShowInputs] = useState(input.repositoryName.length === 0 && input.filePath.length === 0)
    const [isInputFocused, setIsInputFocused] = useState(false)
    const [fileQueryInput, setFileQueryInput] = useState('')
    const debouncedSetFileQueryInput = useMemo(() => debounce(setFileQueryInput, 300), [setFileQueryInput])

    const onFileSelected = useCallback(
        (input: FileBlockInput) => {
            onBlockInputChange(id, { type: 'file', input })
            onRunBlock(id)
        },
        [id, onBlockInputChange, onRunBlock]
    )

    const onLineRangeChange = useCallback(
        (lineRange: IHighlightLineRange | null) => {
            onFileSelected({
                repositoryName: input.repositoryName,
                revision: input.revision,
                filePath: input.filePath,
                lineRange,
            })
        },
        [input.filePath, input.repositoryName, input.revision, onFileSelected]
    )

    const onEnterBlock = useCallback(() => {
        if (!isReadOnly) {
            setShowInputs(true)
        }
    }, [isReadOnly, setShowInputs])

    const hideInputs = useCallback(() => {
        setShowInputs(false)
        setIsInputFocused(false)
    }, [setShowInputs, setIsInputFocused])

    const isFileSelected = input.repositoryName.length > 0 && input.filePath.length > 0
    const blobLines = useObservable(useMemo(() => output?.pipe(startWith(LOADING)) ?? of(undefined), [output]))
    const commonMenuActions = useCommonBlockMenuActions({ isInputFocused, isReadOnly, ...props })
    const fileURL = useMemo(
        () =>
            toPrettyBlobURL({
                repoName: input.repositoryName,
                revision: input.revision,
                filePath: input.filePath,
                range: input.lineRange
                    ? {
                          start: { line: input.lineRange.startLine + 1, character: 0 },
                          end: { line: input.lineRange.endLine, character: 0 },
                      }
                    : undefined,
            }),
        [input]
    )
    const linkMenuAction: BlockMenuAction[] = useMemo(
        () => [
            {
                type: 'link',
                label: 'Open in new tab',
                icon: <OpenInNewIcon className="icon-inline" />,
                url: fileURL,
            },
        ],
        [fileURL]
    )
    const modifierKeyLabel = useModifierKeyLabel()
    const toggleEditMenuAction: BlockMenuAction[] = useMemo(
        () => [
            {
                type: 'button',
                label: showInputs ? 'Save' : 'Edit',
                icon: showInputs ? <CheckIcon className="icon-inline" /> : <PencilIcon className="icon-inline" />,
                onClick: () => setShowInputs(!showInputs),
                keyboardShortcutLabel: showInputs ? `${modifierKeyLabel} + ↵` : '↵',
            },
        ],
        [setShowInputs, modifierKeyLabel, showInputs]
    )

    const menuActions = useMemo(
        () => (!isReadOnly ? toggleEditMenuAction : []).concat(linkMenuAction).concat(commonMenuActions),
        [isReadOnly, toggleEditMenuAction, linkMenuAction, commonMenuActions]
    )

    const onFileURLPaste = useCallback(
        (event: ClipboardEvent) => {
            if (!isSelected || !showInputs || !event.clipboardData) {
                return
            }
            const value = event.clipboardData.getData('text')
            const parsedFileInput = parseFileBlockInput(value)
            if (parsedFileInput.repositoryName.length === 0 || parsedFileInput.filePath.length === 0) {
                return
            }
            onFileSelected(parsedFileInput)
        },
        [isSelected, showInputs, onFileSelected]
    )

    useEffect(() => {
        // We need to add a global paste handler due to focus issues when adding a new block.
        // When a new block is added, we focus it programmatically, but it does not receive the paste events.
        // The user would have to click it manually before copying the file URL. That would result in a weird UX, so we
        // need to handle the paste action globally.
        document.addEventListener('paste', onFileURLPaste)
        return () => document.removeEventListener('paste', onFileURLPaste)
    }, [isSelected, onFileURLPaste])

    const codeIntelViewerUpdatesProps = useMemo(() => ({ extensionsController, ...input }), [
        extensionsController,
        input,
    ])
    const viewerUpdates = useCodeIntelViewerUpdates(codeIntelViewerUpdatesProps)

    return (
        <NotebookBlock
            className={styles.block}
            id={id}
            isReadOnly={isReadOnly}
            isInputFocused={isInputFocused}
            aria-label="Notebook file block"
            onEnterBlock={onEnterBlock}
            isSelected={isSelected}
            isOtherBlockSelected={isOtherBlockSelected}
            onRunBlock={hideInputs}
            onBlockInputChange={onBlockInputChange}
            onSelectBlock={onSelectBlock}
            actions={isSelected ? menuActions : linkMenuAction}
            {...props}
        >
            <div className={styles.header} data-testid="file-block-header">
                {isFileSelected ? <NotebookFileBlockHeader {...input} fileURL={fileURL} /> : <>No file selected.</>}
            </div>
            {showInputs && (
                <NotebookFileBlockInputs
                    id={id}
                    lineRange={input.lineRange}
                    onLineRangeChange={onLineRangeChange}
                    queryInput={fileQueryInput}
                    setQueryInput={setFileQueryInput}
                    debouncedSetQueryInput={debouncedSetFileQueryInput}
                    onRunBlock={hideInputs}
                    setIsInputFocused={setIsInputFocused}
                    onSelectBlock={onSelectBlock}
                    onFileSelected={onFileSelected}
                    {...props}
                />
            )}
            {blobLines && blobLines === LOADING && (
                <div className={classNames('d-flex justify-content-center py-3', styles.highlightedFileWrapper)}>
                    <LoadingSpinner inline={false} />
                </div>
            )}
            {blobLines && blobLines !== LOADING && !isErrorLike(blobLines) && (
                <div className={styles.highlightedFileWrapper}>
                    <CodeExcerpt
                        repoName={input.repositoryName}
                        commitID={input.revision}
                        filePath={input.filePath}
                        blobLines={blobLines}
                        highlightRanges={[]}
                        startLine={input.lineRange?.startLine ?? 0}
                        endLine={input.lineRange?.endLine ?? 1}
                        isFirst={false}
                        fetchHighlightedFileRangeLines={() => of([])}
                        hoverifier={hoverifier}
                        viewerUpdates={viewerUpdates}
                    />
                </div>
            )}
            {blobLines && blobLines !== LOADING && isErrorLike(blobLines) && (
                <Alert className="m-3" variant="danger">
                    {blobLines.message}
                </Alert>
            )}
        </NotebookBlock>
    )
}

const NotebookFileBlockHeader: React.FunctionComponent<FileBlockInput & { fileURL: string }> = ({
    repositoryName,
    filePath,
    revision,
    lineRange,
    fileURL,
}) => (
    <>
        <div className="mr-2">
            <FileDocumentIcon className="icon-inline" />
        </div>
        <div className="d-flex flex-column">
            <div className="mb-1 d-flex align-items-center">
                <Link className={styles.headerFileLink} to={fileURL}>
                    {filePath}
                    {lineRange && <>#{serializeLineRange(lineRange)}</>}
                </Link>
            </div>
            <small className="text-muted">
                {repositoryName}
                {revision && <>@{revision}</>}
            </small>
        </div>
    </>
)
