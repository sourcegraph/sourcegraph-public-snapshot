import classNames from 'classnames'
import { LocationDescriptor } from 'history'
import CloseIcon from 'mdi-react/CloseIcon'
import CodeBracketsIcon from 'mdi-react/CodeBracketsIcon'
import FileDocumentOutlineIcon from 'mdi-react/FileDocumentOutlineIcon'
import NotebookPlusIcon from 'mdi-react/NotebookPlusIcon'
import PencilIcon from 'mdi-react/PencilIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import TextBoxIcon from 'mdi-react/TextBoxIcon'
import TrashIcon from 'mdi-react/TrashCanIcon'
import React, { useCallback, useState, useMemo, MouseEvent, KeyboardEvent, SyntheticEvent } from 'react'
import { useHistory } from 'react-router-dom'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { IHighlightLineRange } from '@sourcegraph/shared/src/schema'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { appendContextFilter, updateFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { buildSearchURLQuery, toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'
import { Button, Link, TextArea } from '@sourcegraph/wildcard'

import { PageRoutes } from '../routes.constants'
import { useExperimentalFeatures } from '../stores'
import {
    useSearchStackState,
    restorePreviousSession,
    SearchEntry,
    SearchStackEntry,
    removeSearchStackEntry,
    removeAllSearchStackEntries,
    SearchStackEntryInput,
    addSearchStackEntry,
    setEntryAnnotation,
} from '../stores/searchStack'

import { BlockInput } from './notebook'
import { serializeBlocks } from './notebook/serialize'
import styles from './SearchStack.module.scss'

export const SearchStack: React.FunctionComponent<{ initialOpen?: boolean }> = ({ initialOpen = false }) => {
    const history = useHistory()

    const [open, setOpen] = useState(initialOpen)
    const [confirmRemoveAll, setConfirmRemoveAll] = useState(false)
    const addableEntry = useSearchStackState(state => state.addableEntry)
    const entries = useSearchStackState(state => state.entries)
    const canRestore = useSearchStackState(state => state.canRestoreSession)
    const enableSearchStack = useExperimentalFeatures(features => features.enableSearchStack)

    const reversedEntries = useMemo(() => [...entries].reverse(), [entries])

    const createNotebook = useCallback(() => {
        const blocks: BlockInput[] = []
        for (const entry of entries) {
            if (entry.annotation) {
                blocks.push({ type: 'md', input: entry.annotation })
            }
            switch (entry.type) {
                case 'search':
                    blocks.push({ type: 'query', input: toSearchQuery(entry) })
                    break
                case 'file':
                    blocks.push({
                        type: 'file',
                        input: {
                            repositoryName: entry.repo,
                            revision: entry.revision,
                            filePath: entry.path,
                            lineRange: entry.lineRange,
                        },
                    })
                    break
            }
        }

        const location = {
            pathname: PageRoutes.NotebookCreate,
            hash: serializeBlocks(blocks, window.location.origin),
        }
        history.push(location)
    }, [entries, history])

    if (!enableSearchStack || (reversedEntries.length === 0 && !addableEntry)) {
        return null
    }

    if (open) {
        return (
            <div className={classNames(styles.root, { [styles.open]: open })}>
                <div className={classNames(styles.header, 'd-flex align-items-center justify-content-between')}>
                    <Button
                        aria-label={`${open ? 'Close' : 'Open'} search session`}
                        variant="icon"
                        className="p-2"
                        onClick={() => setOpen(open => !open)}
                    >
                        <PencilIcon className="icon-inline" />
                        <h4 className={classNames(styles.openVisible, 'px-1')}>Notepad</h4>
                        <small>
                            ({reversedEntries.length} item{reversedEntries.length === 1 ? '' : 's'})
                        </small>
                    </Button>
                    <Button
                        aria-label="Close search session"
                        variant="icon"
                        className={classNames('pr-2', styles.closeButton, styles.openVisible)}
                        onClick={() => setOpen(false)}
                    >
                        <CloseIcon className="icon-inline" />
                    </Button>
                </div>
                <ul>
                    <li className="d-flex flex-column">{addableEntry && <AddEntryButton entry={addableEntry} />}</li>
                    {reversedEntries.map(entry => (
                        <li key={entry.id}>
                            <SearchStackEntryComponent entry={entry} />
                        </li>
                    ))}
                </ul>
                {confirmRemoveAll && (
                    <div className="p-2">
                        <p>Are you sure you want to delete all entries?</p>
                        <div className="d-flex justify-content-between">
                            <Button variant="secondary" onClick={() => setConfirmRemoveAll(false)}>
                                Cancel
                            </Button>
                            <Button
                                variant="danger"
                                onClick={() => {
                                    removeAllSearchStackEntries()
                                    setConfirmRemoveAll(false)
                                }}
                            >
                                Yes, delete
                            </Button>
                        </div>
                    </div>
                )}
                <div className="p-2">
                    {canRestore && (
                        <Button
                            className="w-100 mb-1"
                            onClick={restorePreviousSession}
                            outline={true}
                            variant="secondary"
                            size="sm"
                        >
                            Restore previous session
                        </Button>
                    )}
                    <div className="d-flex justify-content-between align-items-center">
                        <Button onClick={createNotebook} variant="primary" size="sm" disabled={entries.length === 0}>
                            <NotebookPlusIcon className="icon-inline" /> Create Notebook
                        </Button>
                        <Button
                            aria-label="Remove all entries"
                            title="Remove all entries"
                            variant="icon"
                            className="text-muted"
                            disabled={entries.length === 0}
                            onClick={() => setConfirmRemoveAll(true)}
                        >
                            <TrashIcon className="icon-inline" />
                        </Button>
                    </div>
                </div>
            </div>
        )
    }
    const openNotepad = (): void => {
        setOpen(true)
    }
    return (
        <div className={classNames(styles.root)}>
            {reversedEntries.length === 0 && addableEntry && <AddEntryButton entry={addableEntry} />}
            {reversedEntries.length > 0 ? (
                <SearchStackEntryComponent entry={reversedEntries[0]} onSelect={openNotepad} />
            ) : null}
        </div>
    )
}

interface AddEntryButtonProps {
    entry: SearchStackEntryInput
}

const AddEntryButton: React.FunctionComponent<AddEntryButtonProps> = ({ entry }) => {
    switch (entry.type) {
        case 'search':
            return (
                <Button
                    variant="primary"
                    size="sm"
                    title="Add search"
                    className={styles.button}
                    onClick={() => addSearchStackEntry(entry)}
                >
                    + <SearchIcon className="icon-inline" /> Search
                </Button>
            )
        case 'file':
            return (
                <span className={classNames(styles.button, 'd-flex mx-0')}>
                    <Button
                        variant="primary"
                        size="sm"
                        title="Add file"
                        className="flex-1 mx-1"
                        onClick={() => addSearchStackEntry(entry, 'file')}
                    >
                        + <FileDocumentOutlineIcon className="icon-inline" /> File
                    </Button>
                    {entry.lineRange && (
                        <Button
                            variant="primary"
                            size="sm"
                            title="Add line range"
                            className="flex-1 mx-1"
                            onClick={() => addSearchStackEntry(entry, 'range')}
                        >
                            + <CodeBracketsIcon className="icon-inline" /> Range (
                            {entry.lineRange.endLine - entry.lineRange.startLine})
                        </Button>
                    )}
                </span>
            )
    }
}

function stopPropagation(event: SyntheticEvent): void {
    event.stopPropagation()
}

interface SearchStackEntryComponentProps {
    entry: SearchStackEntry
    /**
     * If the notepad is minimized we don't want render a link for the item that
     * is visible, clicking it should open the notepad instead.
     */
    renderLink?: boolean
    onSelect?: (event: MouseEvent | KeyboardEvent, entry: SearchStackEntry) => void
}

const SearchStackEntryComponent: React.FunctionComponent<SearchStackEntryComponentProps> = ({
    entry,
    onSelect,
    renderLink = true,
}) => {
    const { icon, title, location } = getUIComponentsForEntry(entry)
    const [annotation, setAnnotation] = useState(entry.annotation ?? '')
    const [showAnnotationInput, setShowAnnotationInput] = useState(false)

    const titleWrapper = renderLink ? (
        <Link to={location} className="p-0">
            {title}
        </Link>
    ) : (
        title
    )

    function keyHandler(event: KeyboardEvent<HTMLDivElement>): void {
        if (event.key === 'enter') {
            onSelect?.(event, entry)
        }
    }

    return (
        <div
            className={styles.entry}
            onClick={event => onSelect?.(event, entry)}
            onKeyUp={keyHandler}
            role="button"
            tabIndex={0}
        >
            <div className="d-flex">
                <span className="flex-shrink-0 text-muted mr-1">{icon}</span>
                <span className="flex-1">{titleWrapper}</span>
                <span className="ml-1 d-flex">
                    <Button
                        aria-label="Add annotation"
                        title="Add annotation"
                        variant="icon"
                        className="text-muted"
                        onClick={event => {
                            event.stopPropagation()
                            setShowAnnotationInput(show => !show)
                        }}
                    >
                        <TextBoxIcon className="icon-inline" />
                    </Button>
                    <Button
                        aria-label="Remove entry"
                        title="Remove entry"
                        variant="icon"
                        className="ml-1 text-muted"
                        onClick={event => {
                            event.stopPropagation()
                            removeSearchStackEntry(entry)
                        }}
                    >
                        <CloseIcon className="icon-inline" />
                    </Button>
                </span>
            </div>
            {showAnnotationInput && (
                <TextArea
                    className="mt-1"
                    placeholder="Type to add annotation..."
                    value={annotation}
                    onBlur={() => setEntryAnnotation(entry, annotation)}
                    onChange={event => setAnnotation(event.currentTarget.value)}
                    onClick={stopPropagation}
                />
            )}
        </div>
    )
}

function getUIComponentsForEntry(
    entry: SearchStackEntry
): { icon: React.ReactElement; title: React.ReactElement; location: LocationDescriptor; body?: React.ReactElement } {
    switch (entry.type) {
        case 'search':
            return {
                icon: <SearchIcon className="icon-inline" />,
                title: <SyntaxHighlightedSearchQuery query={entry.query} />,
                location: {
                    pathname: '/search',
                    search: buildSearchURLQuery(
                        entry.query,
                        entry.patternType,
                        entry.caseSensitive,
                        entry.searchContext
                    ),
                },
            }
        case 'file':
            return {
                icon: entry.lineRange ? (
                    <CodeBracketsIcon className="icon-inline" />
                ) : (
                    <FileDocumentOutlineIcon className="icon-inline" />
                ),
                title: (
                    <span title={entry.path}>
                        {fileName(entry.path)}
                        {entry.lineRange ? ` ${formatLineRange(entry.lineRange)}` : ''}
                    </span>
                ),
                location: {
                    pathname: toPrettyBlobURL({
                        repoName: entry.repo,
                        revision: entry.revision,
                        filePath: entry.path,
                    }),
                },
            }
    }
}

function toSearchQuery(entry: SearchEntry): string {
    let { query } = entry
    if (entry.patternType !== SearchPatternType.literal) {
        query = updateFilter(entry.query, FilterType.patterntype, entry.patternType)
    }
    if (entry.caseSensitive) {
        query = updateFilter(query, FilterType.case, 'yes')
    }
    if (entry.searchContext) {
        query = appendContextFilter(query, entry.searchContext)
    }
    return query
}

function fileName(path: string): string {
    const parts = path.split('/')
    return parts[parts.length - 1]
}

function formatLineRange(lineRange: IHighlightLineRange): string {
    if (lineRange.startLine === lineRange.endLine - 1) {
        return `L${lineRange.startLine}`
    }
    return `L${lineRange.startLine}:${lineRange.endLine}`
}
