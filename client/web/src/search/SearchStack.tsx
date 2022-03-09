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
    useLayoutEffect,
} from 'react'
import { useHistory } from 'react-router-dom'

import { isMacPlatform } from '@sourcegraph/common'
import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { IHighlightLineRange } from '@sourcegraph/shared/src/schema'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { appendContextFilter, updateFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { buildSearchURLQuery, toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'
import { Button, Link, TextArea } from '@sourcegraph/wildcard'

import { BlockInput } from '../notebooks'
import { serializeBlocksToURL } from '../notebooks/serialize'
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
} from '../stores/searchStack'

import styles from './SearchStack.module.scss'

const SEARCH_STACK_ID = 'search:search-stack'

/**
 * This handler is used on mousedown to prevent text selection when multiple
 * stack entries are selected with Shift+click.
 * (tested in Firefox and Chromium)
 */
function preventTextSelection(event: MouseEvent | KeyboardEvent): void {
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

    const previous = previousLength.current
    previousLength.current = entries.length

    return previous !== undefined && previous < entries.length
}

export const SearchStack: React.FunctionComponent<{ initialOpen?: boolean }> = ({ initialOpen = false }) => {
    const history = useHistory()

    const [open, setOpen] = useState(initialOpen)
    const [confirmRemoveAll, setConfirmRemoveAll] = useState(false)
    const addableEntry = useSearchStackState(state => state.addableEntry)
    const entries = useSearchStackState(state => state.entries)
    const canRestore = useSearchStackState(state => state.canRestoreSession)
    const enableSearchStack = useExperimentalFeatures(features => features.enableSearchStack)
    const [selectedEntries, setSelectedEntries] = useState<number[]>([])
    const isMacPlatform_ = useMemo(isMacPlatform, [])

    const reversedEntries = useMemo(() => [...entries].reverse(), [entries])
    const hasNewEntry = useHasNewEntry(reversedEntries)

    useLayoutEffect(() => {
        if (hasNewEntry) {
            // Always select the new entry. This is also avoids problems with
            // getting the selection index out of sync.
            setSelectedEntries([0])
        }
    }, [hasNewEntry])

    const toggleSelectedEntry = useCallback(
        (position: number, event: MouseEvent | KeyboardEvent) => {
            const { ctrlKey, metaKey, shiftKey } = event

            setSelectedEntries(selectedEntries => {
                if (shiftKey) {
                    // Select range. The range of entries is always computed
                    // from the last selected entry.
                    return extendSelection(selectedEntries, position)
                }

                // Normal (de)selection, taking into account modifier keys for
                // multiple selection.
                // If multiple entries are selected then selecting
                // (without ctrl/cmd/shift) an already selected entry will
                // just select that entry.
                return toggleSelection(
                    selectedEntries,
                    position,
                    (isMacPlatform_ && metaKey) || (!isMacPlatform_ && ctrlKey)
                )
            })
        },
        [setSelectedEntries, isMacPlatform_]
    )

    const deleteSelectedEntries = useCallback(() => {
        if (selectedEntries.length > 0) {
            const entryIDs = selectedEntries.map(index => reversedEntries[index].id)
            removeFromSearchStack(entryIDs)
            // Clear selection for now.
            setSelectedEntries([])
        }
    }, [reversedEntries, selectedEntries, setSelectedEntries])

    const deleteEntry = useCallback(
        (toDelete: SearchStackEntry) => {
            if (selectedEntries.length > 0) {
                const entryPosition = reversedEntries.findIndex(entry => entry.id === toDelete.id)
                setSelectedEntries(selection => adjustSelection(selection, entryPosition))
            }
            removeFromSearchStack([toDelete.id])
        },
        [reversedEntries, selectedEntries, setSelectedEntries]
    )

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

    // Handles key events on the whole list
    const handleKey = useCallback(
        (event: KeyboardEvent): void => {
            const hasMeta = (isMacPlatform_ && event.metaKey) || (!isMacPlatform_ && event.ctrlKey)

            switch (event.key) {
                // Select all entries
                case 'a':
                    if (hasMeta) {
                        // This prevents text selection
                        event.preventDefault()
                        setSelectedEntries(reversedEntries.map((_value, index) => index))
                    }
                    break
                // Clear selection
                case 'Escape':
                    if (selectedEntries.length > 0) {
                        setSelectedEntries([])
                    }
                    break
                // Delete selected entries
                case 'Delete':
                case 'Backspace':
                    if (selectedEntries.length > 0) {
                        deleteSelectedEntries()
                    }
                    break
                // Select "next" entry
                case 'ArrowUp':
                case 'ArrowDown': {
                    const { shiftKey, key } = event

                    if (shiftKey) {
                        // This prevents text selection
                        event.preventDefault()
                    }

                    setSelectedEntries(selection => {
                        if (shiftKey || hasMeta) {
                            // Extend (or shrink) selected entries range
                            // Shift and ctrl modifier are equivalent in this scenario
                            return growOrShrinkSelection(
                                selection,
                                key === 'ArrowDown' ? 'DOWN' : 'UP',
                                reversedEntries.length
                            )
                        }
                        if (selection.length > 0) {
                            // Select next entry
                            return toggleSelection(
                                selection,
                                wrapPosition(
                                    selection[selection.length - 1] + (key === 'ArrowDown' ? 1 : -1),
                                    reversedEntries.length
                                )
                            )
                        }
                        if (reversedEntries.length > 0) {
                            // Select default (bottom or top) entry
                            return [key === 'ArrowDown' ? 0 : reversedEntries.length - 1]
                        }

                        return selection
                    })
                    break
                }
            }
        },
        [reversedEntries, selectedEntries, deleteSelectedEntries, isMacPlatform_]
    )

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
                <ul role="listbox" aria-multiselectable={true} onKeyDown={handleKey} tabIndex={0}>
                    <li className="d-flex flex-column">{addableEntry && <AddEntryButton entry={addableEntry} />}</li>
                    {reversedEntries.map((entry, index) => {
                        const selected = selectedEntries.includes(index)
                        return (
                            <li
                                key={entry.id}
                                role="option"
                                onClick={event => toggleSelectedEntry(index, event)}
                                onKeyDown={event => {
                                    if (document.activeElement === event.currentTarget && event.key === ' ') {
                                        event.stopPropagation()
                                        toggleSelectedEntry(index, event)
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
                                    onDelete={selected ? deleteSelectedEntries : deleteEntry}
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
    onDelete: (entry: SearchStackEntry) => void
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
                            onDelete(entry)
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
                    onKeyUp={event => {
                        // This is used mainly to prevent deletion of the entry
                        // when Delete or Backspace are pressed (one of the
                        // ancestors listens to keyup events to handle
                        // keybindings)
                        event.stopPropagation()
                    }}
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

// Helper functions for working with "selections", an ordered list of indexes
type Selection = number[]

/**
 * Adds or removes a position from a selection. If `multiple` is false (default)
 * but the selection contains multiple elements the new position will always be
 * added.
 *
 * @param selection The selection to operate on
 * @param position The position to add or remove
 * @param multiple Whether to allow multiple selected items or not.
 */
function toggleSelection(selection: Selection, position: number, multiple: boolean = false): Selection {
    const index = selection.indexOf(position)

    if (multiple) {
        if (index === -1) {
            return [...selection, position]
        }

        const newSelection = [...selection]
        newSelection.splice(index, 1)
        return newSelection
    }
    return index === -1 || selection.length > 1 ? [position] : []
}

/**
 * Extends a given selection to contain all positions between the last one and
 * the new newly added one. The new selection will always contain the new
 * position. This will rearrange existing selected positions.
 *
 * ([1,2,3], 5) => [1,2,3,4,5]
 * ([1,2,3], 3) => [1,2,3]
 * ([1,2,3], 2) => [1,3,2]
 * ([1,2,3], 0) => [3,2,1,0]
 *
 * @param selection The selection to operate on
 * @param newPosition The position to extend the selection to
 */
function extendSelection(selection: Selection, newPosition: number): Selection {
    if (selection.length === 0) {
        return [newPosition]
    }

    const newSelection = [...selection]

    const lastSelectedPosition = newSelection[newSelection.length - 1]
    const direction = lastSelectedPosition > newPosition ? -1 : 1
    for (let position = lastSelectedPosition; position !== newPosition + direction; position += direction) {
        // Re-arrange selection as necessary
        const existingSelectionIndex = newSelection.indexOf(position)
        if (existingSelectionIndex > -1) {
            newSelection.splice(existingSelectionIndex, 1)
        }
        newSelection.push(position)
    }
    return newSelection
}

/**
 * This function is supposed to be used when reacting to shift+arrow_up/down
 * events. In particular it
 * - selects the next unselected position
 * - deselects a previously selected position if the direction changes
 *
 * This behavior is different enough from shift+click to warrant its own
 * function.
 *
 * @param selection The selection to operator on
 * @param direction The direction in which to change the selection
 * @param total The total number of entries in the list (to handle
 * wrapping around)
 */
function growOrShrinkSelection(selection: Selection, direction: 'UP' | 'DOWN', total: number): Selection {
    // Select top/bottom element if selection is empty
    if (selection.length === 0) {
        return [direction === 'UP' ? total - 1 : 0]
    }

    const delta = direction === 'UP' ? -1 : 1
    let nextPosition = wrapPosition(selection[selection.length - 1] + delta, total)

    // Did we change direction and "deselected" the last position?
    // (it's enough to look at the penultimate selected position)
    if (selection.length > 1 && selection[selection.length - 2] === nextPosition) {
        return selection.slice(0, -1)
    }

    // Otherwise select the next unselected position (and rearrange positions
    // accordingly)
    const selectionCopy = [...selection]
    let index = selectionCopy.indexOf(nextPosition)
    while (index !== -1) {
        selectionCopy.splice(index, 1)
        selectionCopy.push(nextPosition)
        nextPosition += delta
        index = selectionCopy.indexOf(nextPosition)
    }
    selectionCopy.push(nextPosition)
    return selectionCopy
}

/**
 * Adjusts all indexes in the selection which are above the removed position.
 */
function adjustSelection(selection: Selection, removedPosition: number): Selection {
    const result: number[] = []
    for (const position of selection) {
        if (position === removedPosition) {
            continue
        } else if (position > removedPosition) {
            result.push(position - 1)
            continue
        }
        result.push(position)
    }
    return result
}

/**
 * Helper function for properly wrapping a value between 0 and max (exclusive).
 * Basically modulo without negative numbers.
 */
function wrapPosition(position: number, max: number): number {
    if (position >= max) {
        return position % max
    }
    if (position < 0) {
        return max + position
    }
    return position
}
