import classNames from 'classnames'
import CheckIcon from 'mdi-react/CheckIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import MinusIcon from 'mdi-react/MinusIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import PencilIcon from 'mdi-react/PencilIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import React, { useState, useCallback, useRef, useMemo, useEffect } from 'react'
import { of } from 'rxjs'
import { startWith } from 'rxjs/operators'

import { isErrorLike } from '@sourcegraph/common'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { CodeExcerpt } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { BlockProps, FileBlock, FileBlockInput } from '..'
import blockStyles from '../SearchNotebookBlock.module.scss'
import { BlockMenuAction, SearchNotebookBlockMenu } from '../SearchNotebookBlockMenu'
import { isSingleLineRange, serializeLineRange } from '../serialize'
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
        TelemetryProps {
    isMacPlatform: boolean
    isSourcegraphDotCom: boolean
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
        if (!areInputsValid) {
            return
        }
        onRunBlock(id)
    }, [id, input, areInputsValid, onRunBlock])

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
                        <LoadingSpinner />
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
                        />
                    </div>
                )}
                {blobLines && blobLines !== LOADING && isErrorLike(blobLines) && (
                    <div className="alert alert-danger m-3">{blobLines.message}</div>
                )}
            </div>
            {(isSelected || !isOtherBlockSelected) && (
                <SearchNotebookBlockMenu id={id} actions={isSelected ? menuActions : linkMenuAction} />
            )}
        </div>
    )
}
