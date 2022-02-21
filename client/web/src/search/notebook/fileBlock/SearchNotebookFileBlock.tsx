import classNames from 'classnames'
import CheckIcon from 'mdi-react/CheckIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import MinusIcon from 'mdi-react/MinusIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import PencilIcon from 'mdi-react/PencilIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import React, { useState, useCallback, useRef, useMemo, useEffect } from 'react'
import { of, ReplaySubject } from 'rxjs'
import { startWith } from 'rxjs/operators'

import { Hoverifier } from '@sourcegraph/codeintellify'
import { isErrorLike } from '@sourcegraph/common'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { HoverMerged } from '@sourcegraph/shared/src/api/client/types/hover'
import { ViewerId } from '@sourcegraph/shared/src/api/viewerTypes'
import { CodeExcerpt } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { HoverContext } from '@sourcegraph/shared/src/hover/HoverOverlay'
import { getModeFromPath } from '@sourcegraph/shared/src/languages'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { toPrettyBlobURL, toURIWithPath } from '@sourcegraph/shared/src/util/url'
import { LoadingSpinner, useObservable, Link, Alert } from '@sourcegraph/wildcard'

import { BlockProps, FileBlock, FileBlockInput } from '..'
import blockStyles from '../SearchNotebookBlock.module.scss'
import { BlockMenuAction, SearchNotebookBlockMenu } from '../SearchNotebookBlockMenu'
import { isSingleLineRange, parseFileBlockInput, serializeLineRange } from '../serialize'
import { useBlockSelection } from '../useBlockSelection'
import { useBlockShortcuts } from '../useBlockShortcuts'
import { useCommonBlockMenuActions } from '../useCommonBlockMenuActions'

import styles from './SearchNotebookFileBlock.module.scss'
import { SearchNotebookFileBlockInputs } from './SearchNotebookFileBlockInputs'
import { FileBlockValidationFunctions, useFileBlockInputValidation } from './useFileBlockInputValidation'

interface SearchNotebookFileBlockProps
    extends BlockProps,
        Omit<FileBlock, 'type'>,
        FileBlockValidationFunctions,
        TelemetryProps,
        ExtensionsControllerProps<'extHostAPI' | 'executeCommand'> {
    isMacPlatform: boolean
    isSourcegraphDotCom: boolean
    hoverifier: Hoverifier<HoverContext, HoverMerged, ActionItemAction>
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

export const SearchNotebookFileBlock: React.FunctionComponent<SearchNotebookFileBlockProps> = ({
    id,
    input,
    output,
    telemetryService,
    isSelected,
    isOtherBlockSelected,
    isMacPlatform,
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
    const blockElement = useRef<HTMLDivElement>(null)
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

    const { onSelect } = useBlockSelection({
        id,
        blockElement: blockElement.current,
        isSelected,
        isInputFocused,
        onSelectBlock,
        ...props,
    })

    const { onKeyDown } = useBlockShortcuts({
        id,
        isMacPlatform,
        onEnterBlock: () => setShowInputs(true),
        onRunBlock: () => {
            setShowInputs(false)
            setIsInputFocused(false)
        },
        ...props,
    })

    const blobLines = useObservable(useMemo(() => output?.pipe(startWith(LOADING)) ?? of(undefined), [output]))

    const areInputsValid =
        isRepositoryNameValid === true &&
        isFilePathValid === true &&
        isRevisionValid !== false &&
        isLineRangeValid !== false

    const modifierKeyLabel = isMacPlatform ? '⌘' : 'Ctrl'
    const commonMenuActions = useCommonBlockMenuActions({
        modifierKeyLabel,
        isInputFocused,
        isMacPlatform,
        isReadOnly,
        ...props,
    })

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

    // Inform the extension host about the file (if we have code to render). CodeExcerpt will call `hoverifier.hoverify`.
    const viewerUpdates = useMemo(() => new ReplaySubject<{ viewerId: ViewerId } & HoverContext>(1), [])
    useEffect(() => {
        let previousViewerId: ViewerId | undefined
        const commitID = input.revision || 'HEAD'
        const uri = toURIWithPath({ repoName: input.repositoryName, filePath: input.filePath, commitID })
        const languageId = getModeFromPath(input.filePath)
        // HACK: code intel extensions don't depend on the `text` field.
        // Fix to support other hover extensions on search results (likely too expensive).
        const text = ''

        extensionsController.extHostAPI
            .then(extensionHostAPI =>
                Promise.all([
                    // This call should be made before adding viewer, but since
                    // messages to web worker are handled in order, we can use Promise.all
                    extensionHostAPI.addTextDocumentIfNotExists({
                        uri,
                        languageId,
                        text,
                    }),
                    extensionHostAPI.addViewerIfNotExists({
                        type: 'CodeEditor' as const,
                        resource: uri,
                        selections: [],
                        isActive: true,
                    }),
                ])
            )
            .then(([, viewerId]) => {
                previousViewerId = viewerId
                viewerUpdates.next({
                    viewerId,
                    repoName: input.repositoryName,
                    revision: commitID,
                    commitID,
                    filePath: input.filePath,
                })
            })
            .catch(error => {
                console.error('Extension host API error', error)
            })

        return () => {
            // Remove from extension host
            extensionsController.extHostAPI
                .then(extensionHostAPI => previousViewerId && extensionHostAPI.removeViewer(previousViewerId))
                .catch(error => console.error('Error removing viewer from extension host', error))
        }
    }, [input, viewerUpdates, extensionsController])

    return (
        <div className={classNames('block-wrapper', blockStyles.blockWrapper)} data-block-id={id}>
            {/* See the explanation for the disable in markdown and query blocks. */}
            {/* eslint-disable-next-line jsx-a11y/no-static-element-interactions */}
            <div
                className={classNames(
                    blockStyles.block,
                    styles.block,
                    isSelected && !isInputFocused && blockStyles.selected,
                    isSelected && isInputFocused && blockStyles.selectedNotFocused
                )}
                onClick={onSelect}
                onKeyDown={onKeyDown}
                onFocus={onSelect}
                // A tabIndex is necessary to make the block focusable.
                // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
                tabIndex={0}
                aria-label="Notebook file block"
                ref={blockElement}
            >
                {showInputs ? (
                    <SearchNotebookFileBlockInputs
                        id={id}
                        {...input}
                        lineRangeInput={lineRangeInput}
                        isRepositoryNameValid={isRepositoryNameValid}
                        isFilePathValid={isFilePathValid}
                        isRevisionValid={isRevisionValid}
                        isLineRangeValid={isLineRangeValid}
                        showRevisionInput={showRevisionInput}
                        showLineRangeInput={showLineRangeInput}
                        isMacPlatform={isMacPlatform}
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
            </div>
            {(isSelected || !isOtherBlockSelected) && (
                <SearchNotebookBlockMenu id={id} actions={isSelected ? menuActions : linkMenuAction} />
            )}
        </div>
    )
}
