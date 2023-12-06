import React, { useState, useCallback, useMemo, useEffect } from 'react'

import type { EditorView } from '@codemirror/view'
import { mdiOpenInNew, mdiFileDocument, mdiCheck, mdiPencil } from '@mdi/js'
import { debounce } from 'lodash'
import { of } from 'rxjs'
import { startWith } from 'rxjs/operators'

import { CodeExcerpt } from '@sourcegraph/branded'
import { isErrorLike } from '@sourcegraph/common'
import { getRepositoryUrl } from '@sourcegraph/shared/src/search/stream'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { codeCopiedEvent } from '@sourcegraph/shared/src/tracking/event-log-creators'
import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'
import { LoadingSpinner, useObservable, Icon, Alert } from '@sourcegraph/wildcard'

import type { BlockProps, FileBlock, FileBlockInput } from '../..'
import type { HighlightLineRange } from '../../../graphql-operations'
import { focusEditor } from '../../codemirror-utils'
import { parseFileBlockInput } from '../../serialize'
import type { BlockMenuAction } from '../menu/NotebookBlockMenu'
import { useCommonBlockMenuActions } from '../menu/useCommonBlockMenuActions'
import { NotebookBlock } from '../NotebookBlock'
import { RepoFileSymbolLink } from '../RepoFileSymbolLink'
import { useModifierKeyLabel } from '../useModifierKeyLabel'

import { NotebookFileBlockInputs } from './NotebookFileBlockInputs'

import styles from './NotebookFileBlock.module.scss'

interface NotebookFileBlockProps extends BlockProps<FileBlock>, TelemetryProps, TelemetryV2Props {
    isSourcegraphDotCom: boolean
}

const LOADING = 'loading' as const

export const NotebookFileBlock: React.FunctionComponent<React.PropsWithChildren<NotebookFileBlockProps>> = React.memo(
    ({
        id,
        input,
        output,
        telemetryService,
        telemetryRecorder,
        isSelected,
        showMenu,
        isReadOnly,
        onRunBlock,
        onBlockInputChange,
        ...props
    }) => {
        const [editor, setEditor] = useState<EditorView | undefined>()
        const [showInputs, setShowInputs] = useState(input.repositoryName.length === 0 && input.filePath.length === 0)
        const [fileQueryInput, setFileQueryInput] = useState(input.initialQueryInput ?? '')
        const debouncedSetFileQueryInput = useMemo(() => debounce(setFileQueryInput, 300), [setFileQueryInput])

        const onFileSelected = useCallback(
            (input: FileBlockInput) => {
                onBlockInputChange(id, { type: 'file', input })
                onRunBlock(id)
            },
            [id, onBlockInputChange, onRunBlock]
        )

        const onLineRangeChange = useCallback(
            (lineRange: HighlightLineRange | null) => {
                onFileSelected({
                    repositoryName: input.repositoryName,
                    revision: input.revision,
                    filePath: input.filePath,
                    lineRange,
                })
            },
            [input.filePath, input.repositoryName, input.revision, onFileSelected]
        )

        const focusInput = useCallback(() => {
            if (editor) {
                focusEditor(editor)
            }
        }, [editor])

        const hideInputs = useCallback(() => setShowInputs(false), [setShowInputs])

        const isFileSelected = input.repositoryName.length > 0 && input.filePath.length > 0
        const blobLines = useObservable(useMemo(() => output?.pipe(startWith(LOADING)) ?? of(undefined), [output]))
        const commonMenuActions = useCommonBlockMenuActions({ id, isReadOnly, ...props })
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
                    icon: <Icon aria-hidden={true} svgPath={mdiOpenInNew} />,
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
                    icon: <Icon aria-hidden={true} svgPath={showInputs ? mdiCheck : mdiPencil} />,
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

        const logEventOnCopy = useCallback(() => {
            telemetryService.log(...codeCopiedEvent('notebook-file-block'))
            telemetryRecorder.recordEvent('notebook-file-block', 'copied')
        }, [telemetryService, telemetryRecorder])

        return (
            <NotebookBlock
                className={styles.block}
                id={id}
                aria-label="Notebook file block"
                isSelected={isSelected}
                showMenu={showMenu}
                isReadOnly={isReadOnly}
                isInputVisible={showInputs}
                setIsInputVisible={setShowInputs}
                focusInput={focusInput}
                actions={isSelected ? menuActions : linkMenuAction}
                {...props}
            >
                <div className={styles.header} data-testid="file-block-header">
                    {isFileSelected ? <NotebookFileBlockHeader {...input} fileURL={fileURL} /> : <>No file selected.</>}
                </div>
                {showInputs && (
                    <NotebookFileBlockInputs
                        id={id}
                        onEditorCreated={setEditor}
                        lineRange={input.lineRange}
                        onLineRangeChange={onLineRangeChange}
                        queryInput={fileQueryInput}
                        setQueryInput={debouncedSetFileQueryInput}
                        onRunBlock={hideInputs}
                        onFileSelected={onFileSelected}
                        {...props}
                    />
                )}
                {blobLines && blobLines === LOADING && (
                    <div className="d-flex justify-content-center py-3">
                        <LoadingSpinner inline={false} />
                    </div>
                )}
                {blobLines && blobLines !== LOADING && !isErrorLike(blobLines) && (
                    <div>
                        <CodeExcerpt
                            className={styles.code}
                            repoName={input.repositoryName}
                            commitID={input.revision}
                            filePath={input.filePath}
                            blobLines={blobLines}
                            highlightRanges={[]}
                            startLine={input.lineRange?.startLine ?? 0}
                            endLine={input.lineRange?.endLine ?? 1}
                            fetchHighlightedFileRangeLines={() => of([])}
                            onCopy={logEventOnCopy}
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
)

const NotebookFileBlockHeader: React.FunctionComponent<
    React.PropsWithChildren<FileBlockInput & { fileURL: string }>
> = ({ repositoryName, filePath, revision, fileURL }) => {
    const repoAtRevisionURL = getRepositoryUrl(repositoryName, [revision])
    return (
        <>
            <Icon aria-hidden={true} svgPath={mdiFileDocument} />
            <div className={styles.separator} />
            <RepoFileSymbolLink
                repoName={repositoryName}
                repoURL={repoAtRevisionURL}
                filePath={filePath}
                fileURL={fileURL}
            />
        </>
    )
}
