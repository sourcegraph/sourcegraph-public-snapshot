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
import React, {
    useCallback,
    useState,
    useMemo,
    KeyboardEvent,
    SyntheticEvent,
    MouseEvent,
    useRef,
    useEffect,
} from 'react'
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
    removeFromSearchStack,
    removeAllSearchStackEntries,
    SearchStackEntryInput,
    addSearchStackEntry,
    setEntryAnnotation,
    SearchStackEntryID,
} from '../stores/searchStack'

import { BlockInput } from './notebook'
import { serializeBlocksToURL } from './notebook/serialize'
import styles from './SearchStack.module.scss'

const SEARCH_STACK_ID = 'search:search-stack'

/**
 * This handler is used on mousedown to prevent text selection when multiple
 * stack entries are selected with Shift+click.
 * (tested in Firefox and Chromium)
 */
function preventTextSelection(event: MouseEvent): void {
    if (event.shiftKey) {
        event.preventDefault()
    }
}

/**
 * Helper hook to determine whether a new entry has been added to the search
 * stack. Whenever the number of entries increases we have a new entry. It
 * assumes that the newest entry is the first element in the input array.
 */
function useHasNewEntry(entries: SearchStackEntry[]): boolean {
    const previousLength = useRef<number>()

    useEffect(() => {
        previousLength.current = entries.length
    }, [entries])

    return previousLength.current !== undefined && previousLength.current < entries.length
}

export const SearchStack: React.FunctionComponent<{ initialOpen?: boolean }> = ({ initialOpen = false }) => {
    const history = useHistory()

    const [open, setOpen] = useState(initialOpen)
    const [confirmRemoveAll, setConfirmRemoveAll] = useState(false)
    const addableEntry = useSearchStackState(state => state.addableEntry)
    const entries = useSearchStackState(state => state.entries)
    const canRestore = useSearchStackState(state => state.canRestoreSession)
    const enableSearchStack = useExperimentalFeatures(features => features.enableSearchStack)
    const [selectedEntries, setSelectedEntries] = useState<SearchStackEntryID[]>([])

    const reversedEntries = useMemo(() => [...entries].reverse(), [entries])
    const hasNewEntry = useHasNewEntry(reversedEntries)

    const toggleSelectedEntry = useCallback(
        (entry: SearchStackEntry, event: MouseEvent | KeyboardEvent) => {
            const { ctrlKey, metaKey, shiftKey } = event

            setSelectedEntries(selectedEntries => {
                const index = selectedEntries.indexOf(entry.id)

                if (ctrlKey || metaKey) {
                    // Add or remove item to selection
                    if (index > -1) {
                        const copy = selectedEntries.slice()
                        copy.splice(index, 1)
                        return copy
                    }
                    return [...selectedEntries, entry.id]
                }

                if (shiftKey) {
                    // Select range. This will always add items.
                    // The range of entries is always computed from the last
                    // selected entry.
                    if (selectedEntries.length === 0) {
                        return [entry.id]
                    }

                    const newSelectedEntries = [...selectedEntries]

                    const lastSelectedID = selectedEntries[selectedEntries.length - 1]
                    // Find all entries between this one the selected entry
                    const indexA = reversedEntries.findIndex(entry => entry.id === lastSelectedID)
                    const indexB = reversedEntries.indexOf(entry)
                    const direction = indexA > indexB ? -1 : 1
                    for (let index_ = indexA; index_ !== indexB + direction; index_ += direction) {
                        // Re-arrange selected entries as necessary
                        const existingSelectionIndex = newSelectedEntries.indexOf(reversedEntries[index_].id)
                        if (existingSelectionIndex > -1) {
                            newSelectedEntries.splice(existingSelectionIndex, 1)
                        }
                        newSelectedEntries.push(reversedEntries[index_].id)
                    }
                    return newSelectedEntries
                }

                // Normal (de)selection
                if (index > -1) {
                    // If multiple entries are selected then selecting
                    // (without ctrl/cmd/shift) an already selected entry will
                    // just select that entry.
                    if (selectedEntries.length > 1) {
                        return [entry.id]
                    }
                    // Otherwise we delesect it.
                    return []
                }

                return [entry.id]
            })
        },
        [reversedEntries, setSelectedEntries]
    )

    const deleteSelectedEntries = useCallback(() => {
        removeFromSearchStack(selectedEntries)
        setSelectedEntries([])
    }, [selectedEntries, setSelectedEntries])

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
                            // Notebooks expect the end line to be exclusive
                            lineRange: entry.lineRange
                                ? { ...entry.lineRange, endLine: entry.lineRange?.endLine + 1 }
                                : null,
                        },
                    })
                    break
            }
        }

        const location = {
            pathname: PageRoutes.NotebookCreate,
            hash: serializeBlocksToURL(blocks, window.location.origin),
        }
        history.push(location)
    }, [entries, history])

    const toggleOpen = useCallback(() => {
        setOpen(open => {
            if (open) {
                // clear selected entries on close
                setSelectedEntries([])
            }
            return !open
        })
    }, [setSelectedEntries, setOpen])

    if (!enableSearchStack || (reversedEntries.length === 0 && !addableEntry)) {
        return null
    }

    if (open) {
        return (
            <section className={classNames(styles.root, { [styles.open]: open })} id={SEARCH_STACK_ID} role="dialog">
                <div className={classNames(styles.header, 'd-flex align-items-center justify-content-between')}>
                    <Button
                        aria-label="Close search session"
                        variant="icon"
                        className="p-2"
                        onClick={toggleOpen}
                        aria-controls={SEARCH_STACK_ID}
                        aria-expanded="true"
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
                        onClick={toggleOpen}
                        aria-controls={SEARCH_STACK_ID}
                        aria-expanded="true"
                    >
                        <CloseIcon className="icon-inline" />
                    </Button>
                </div>
                <ul role="listbox">
                    <li className="d-flex flex-column">{addableEntry && <AddEntryButton entry={addableEntry} />}</li>
                    {reversedEntries.map((entry, index) => {
                        const selected = selectedEntries.includes(entry.id)
                        return (
                            <li
                                key={entry.id}
                                role="option"
                                onClick={event => toggleSelectedEntry(entry, event)}
                                onKeyUp={event => {
                                    if (document.activeElement === event.currentTarget && event.key === ' ') {
                                        toggleSelectedEntry(entry, event)
                                    }
                                }}
                                aria-selected={selected}
                                aria-label={getLabel(entry)}
                                onMouseDown={preventTextSelection}
                                tabIndex={0}
                            >
                                <SearchStackEntryComponent
                                    entry={entry}
                                    focus={hasNewEntry && index === 0}
                                    selected={selected}
                                    onDelete={selected ? deleteSelectedEntries : () => removeFromSearchStack(entry.id)}
                                />
                            </li>
                        )
                    })}
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
            </section>
        )
    }

    const handleEnterKey = (event: KeyboardEvent<HTMLDivElement>): void => {
        if (event.key === 'enter') {
            toggleOpen()
        }
    }

    return (
        <section id={SEARCH_STACK_ID} className={classNames(styles.root)} aria-label="Notepad">
            <div
                role="button"
                aria-expanded="false"
                aria-controls={SEARCH_STACK_ID}
                onClick={toggleOpen}
                onKeyUp={handleEnterKey}
                aria-label="Open search session"
                tabIndex={0}
            >
                {reversedEntries.length === 0 && addableEntry && <AddEntryButton entry={addableEntry} />}
                {reversedEntries.length > 0 ? (
                    // `key` is necessary here to force new elemments being created
                    // when the top entry is deleted. Otherwise the annotations
                    // input isn't rendered correctly.
                    <SearchStackEntryComponent
                        key={reversedEntries[0].id}
                        entry={reversedEntries[0]}
                        focus={hasNewEntry}
                        onDelete={() => removeFromSearchStack(reversedEntries[0].id)}
                    />
                ) : null}
            </div>
        </section>
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
                    onClick={event => {
                        event.stopPropagation()
                        addSearchStackEntry(entry)
                    }}
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
                        onClick={event => {
                            event.stopPropagation()
                            addSearchStackEntry(entry, 'file')
                        }}
                    >
                        + <FileDocumentOutlineIcon className="icon-inline" /> File
                    </Button>
                    {entry.lineRange && (
                        <Button
                            variant="primary"
                            size="sm"
                            title="Add line range"
                            className="flex-1 mx-1"
                            onClick={event => {
                                event.stopPropagation()
                                addSearchStackEntry(entry, 'range')
                            }}
                        >
                            + <CodeBracketsIcon className="icon-inline" /> Range (
                            {entry.lineRange.endLine - entry.lineRange.startLine + 1})
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
     * If set to true, show and focus the annotations input.
     */
    focus?: boolean
    selected?: boolean
    onDelete: (event: MouseEvent | KeyboardEvent) => void
}

const SearchStackEntryComponent: React.FunctionComponent<SearchStackEntryComponentProps> = ({
    entry,
    focus = false,
    selected,
    onDelete,
}) => {
    const { icon, title, location } = getUIComponentsForEntry(entry)
    const [annotation, setAnnotation] = useState(entry.annotation ?? '')
    const [showAnnotationInput, setShowAnnotationInput] = useState(focus)
    const textarea = useRef<HTMLTextAreaElement | null>(null)

    useEffect(() => {
        textarea.current?.focus()
    }, [focus])

    const deletionLabel = selected ? 'Remove all selected entries' : 'Remove entry'

    return (
        <div className={classNames(styles.entry, { [styles.selected]: selected })}>
            <div className="d-flex">
                <span className="flex-shrink-0 text-muted mr-1">{icon}</span>
                <span className="flex-1">
                    <Link to={location} className="p-0">
                        {title}
                    </Link>
                </span>
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
                        aria-label={deletionLabel}
                        title={deletionLabel}
                        variant="icon"
                        className="ml-1 text-muted"
                        onClick={event => {
                            event.stopPropagation()
                            onDelete(event)
                        }}
                    >
                        <CloseIcon className="icon-inline" />
                    </Button>
                </span>
            </div>
            {showAnnotationInput && (
                <TextArea
                    ref={textarea}
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

function getLabel(entry: SearchStackEntry): string {
    switch (entry.type) {
        case 'search':
            return `search: ${toSearchQuery(entry)}`
        case 'file':
            if (entry.lineRange) {
                return `line range: ${fileName(entry.path)}${formatLineRange(entry.lineRange)}`
            }
            return `file: ${fileName(entry.path)}`
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
    if (lineRange.startLine === lineRange.endLine) {
        return `L${lineRange.startLine + 1}`
    }
    return `L${lineRange.startLine + 1}:${lineRange.endLine + 1}`
}
