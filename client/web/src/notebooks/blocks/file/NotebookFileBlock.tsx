import classNames from 'classnames'
import CheckIcon from 'mdi-react/CheckIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import MinusIcon from 'mdi-react/MinusIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import PencilIcon from 'mdi-react/PencilIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import React, { useState, useCallback, useMemo, useEffect } from 'react'
import { of } from 'rxjs'
import { startWith } from 'rxjs/operators'

import { Hoverifier } from '@sourcegraph/codeintellify'
import { isErrorLike } from '@sourcegraph/common'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { HoverMerged } from '@sourcegraph/shared/src/api/client/types/hover'
import { CodeExcerpt } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { HoverContext } from '@sourcegraph/shared/src/hover/HoverOverlay'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'
import { useCodeIntelViewerUpdates } from '@sourcegraph/shared/src/util/useCodeIntelViewerUpdates'
import { LoadingSpinner, useObservable, Link, Alert } from '@sourcegraph/wildcard'

import { BlockProps, FileBlock, FileBlockInput } from '../..'
import { isSingleLineRange, parseFileBlockInput, serializeLineRange } from '../../serialize'
import { BlockMenuAction } from '../menu/NotebookBlockMenu'
import { useCommonBlockMenuActions } from '../menu/useCommonBlockMenuActions'
import { NotebookBlock } from '../NotebookBlock'
import { useModifierKeyLabel } from '../useModifierKeyLabel'

import styles from './NotebookFileBlock.module.scss'
import { NotebookFileBlockInputs } from './NotebookFileBlockInputs'
import { FileBlockValidationFunctions, useFileBlockInputValidation } from './useFileBlockInputValidation'

interface NotebookFileBlockProps
    extends BlockProps<FileBlock>,
        FileBlockValidationFunctions,
        TelemetryProps,
        ExtensionsControllerProps<'extHostAPI' | 'executeCommand'> {
    isSourcegraphDotCom: boolean
    hoverifier?: Hoverifier<HoverContext, HoverMerged, ActionItemAction>
}

const LOADING = 'loading' as const

function getFileHeader(input: FileBlockInput): string {
    const repositoryName = input.repositoryName.trim()
    const filePath = repositoryName ? `/${input.filePath}` : input.filePath
    const revision = input.revision ? `@${input.revision}` : ''
    const lineRange = serializeLineRange(input.lineRange)
    const lines = isSingleLineRange(input.lineRange) ? 'line' : 'lines'
    const lineRangeSummary = lineRange ? `, ${lines} ${lineRange}` : ''
    return `${repositoryName}${filePath}${revision}${lineRangeSummary}`
}

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
    fetchHighlightedFileLineRanges,
    resolveRevision,
    fetchRepository,
    onRunBlock,
    onSelectBlock,
    onBlockInputChange,
    ...props
}) => {
    const [showInputs, setShowInputs] = useState(input.repositoryName.length === 0 && input.filePath.length === 0)
    const [isInputFocused, setIsInputFocused] = useState(false)
    const [lineRangeInput, setLineRangeInput] = useState(serializeLineRange(input.lineRange))
    const [showRevisionInput, setShowRevisionInput] = useState(input.revision.trim().length > 0)
    const [showLineRangeInput, setShowLineRangeInput] = useState(!!input.lineRange)

    const setFileInput = useCallback(
        (newInput: Partial<FileBlockInput>) =>
            onBlockInputChange(id, { type: 'file', input: { ...input, ...newInput } }),
        [id, input, onBlockInputChange]
    )

    const { isRepositoryNameValid, isFilePathValid, isRevisionValid, isLineRangeValid } = useFileBlockInputValidation(
        input,
        lineRangeInput,
        {
            fetchHighlightedFileLineRanges,
            resolveRevision,
            fetchRepository,
        }
    )

    const onEnterBlock = useCallback(() => setShowInputs(true), [setShowInputs])

    const hideInputs = useCallback(() => {
        setShowInputs(false)
        setIsInputFocused(false)
    }, [setShowInputs, setIsInputFocused])

    const blobLines = useObservable(useMemo(() => output?.pipe(startWith(LOADING)) ?? of(undefined), [output]))

    const areInputsValid =
        isRepositoryNameValid === true &&
        isFilePathValid === true &&
        isRevisionValid !== false &&
        isLineRangeValid !== false

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
                isDisabled: !areInputsValid,
            },
        ],
        [fileURL, areInputsValid]
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

    const toggleOptionalInputsMenuActions: BlockMenuAction[] = useMemo(
        () => [
            {
                type: 'button',
                label: showRevisionInput ? 'Remove revision' : 'Add revision',
                icon: showRevisionInput ? <MinusIcon className="icon-inline" /> : <PlusIcon className="icon-inline" />,
                onClick: () => {
                    setFileInput({ revision: '' })
                    setShowRevisionInput(!showRevisionInput)
                },
            },
            {
                type: 'button',
                label: showLineRangeInput ? 'Remove line range' : 'Add line range',
                icon: showLineRangeInput ? <MinusIcon className="icon-inline" /> : <PlusIcon className="icon-inline" />,
                onClick: () => {
                    setLineRangeInput('')
                    setFileInput({ lineRange: null })
                    setShowLineRangeInput(!showLineRangeInput)
                },
            },
        ],
        [setFileInput, showLineRangeInput, showRevisionInput]
    )

    const menuActions = useMemo(
        () =>
            (!isReadOnly ? toggleEditMenuAction : [])
                .concat(showInputs ? toggleOptionalInputsMenuActions : [])
                .concat(linkMenuAction)
                .concat(commonMenuActions),
        [
            isReadOnly,
            toggleEditMenuAction,
            showInputs,
            toggleOptionalInputsMenuActions,
            linkMenuAction,
            commonMenuActions,
        ]
    )

    // Automatically fetch the highlighted file on each input change, if all inputs are valid
    useEffect(() => {
        if (!showInputs || !areInputsValid) {
            return
        }
        onRunBlock(id)
    }, [id, input, showInputs, areInputsValid, onRunBlock])

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
            setShowRevisionInput(parsedFileInput.revision.length > 0)
            setShowLineRangeInput(!!parsedFileInput.lineRange)
            if (parsedFileInput.lineRange) {
                setLineRangeInput(serializeLineRange(parsedFileInput.lineRange))
            }
            setFileInput(parsedFileInput)
        },
        [showInputs, isSelected, setFileInput, setShowRevisionInput, setShowLineRangeInput, setLineRangeInput]
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
            {showInputs ? (
                <NotebookFileBlockInputs
                    id={id}
                    {...input}
                    lineRangeInput={lineRangeInput}
                    isRepositoryNameValid={isRepositoryNameValid}
                    isFilePathValid={isFilePathValid}
                    isRevisionValid={isRevisionValid}
                    isLineRangeValid={isLineRangeValid}
                    showRevisionInput={showRevisionInput}
                    showLineRangeInput={showLineRangeInput}
                    setIsInputFocused={setIsInputFocused}
                    onSelectBlock={onSelectBlock}
                    setFileInput={setFileInput}
                    setLineRangeInput={setLineRangeInput}
                />
            ) : (
                <div className={styles.header} data-testid="file-block-header">
                    <FileDocumentIcon className="icon-inline mr-2" />
                    {areInputsValid ? (
                        <Link className={styles.headerFileLink} to={fileURL}>
                            {getFileHeader(input)}
                        </Link>
                    ) : (
                        <span>{getFileHeader(input)}</span>
                    )}
                </div>
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
