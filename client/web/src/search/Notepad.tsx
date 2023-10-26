import React, {
    useCallback,
    useState,
    useMemo,
    type KeyboardEvent,
    type SyntheticEvent,
    type MouseEvent,
    useRef,
    useEffect,
    useLayoutEffect,
} from 'react'

import {
    mdiBookPlusOutline,
    mdiChevronUp,
    mdiDelete,
    mdiMagnify,
    mdiFileDocumentOutline,
    mdiCodeBrackets,
    mdiTextBox,
} from '@mdi/js'
import classNames from 'classnames'
import type { LocationDescriptorObject } from 'history'
import { useNavigate } from 'react-router-dom'
import * as uuid from 'uuid'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded'
import { isMacPlatform, logger } from '@sourcegraph/common'
import { type HighlightLineRange, SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { appendContextFilter, updateFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { buildSearchURLQuery, toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'
import { Button, Link, TextArea, Icon, H2, H3, Text, createLinkUrl, useMatchMedia } from '@sourcegraph/wildcard'

import { useSidebarSize } from '../cody/sidebar/useSidebarSize'
import type { BlockInput } from '../notebooks'
import { createNotebook } from '../notebooks/backend'
import { blockToGQLInput } from '../notebooks/serialize'
import { EnterprisePageRoutes } from '../routes.constants'
import {
    addNotepadEntry,
    type NotepadEntry,
    type NotepadEntryID,
    type NotepadEntryInput,
    removeAllNotepadEntries,
    removeFromNotepad,
    restorePreviousSession,
    type SearchEntry,
    setEntryAnnotation,
    useNotepadState,
} from '../stores/notepad'

import styles from './Notepad.module.scss'

const NOTEPAD_ID = 'search:notepad'

function isMacMetaKey(event: KeyboardEvent, isMacPlatform: boolean): boolean {
    return isMacPlatform && event.metaKey
}

function isMetaKey(event: KeyboardEvent, isMacPlatform: boolean): boolean {
    return isMacMetaKey(event, isMacPlatform) || (!isMacPlatform && event.ctrlKey)
}

/**
 * This handler is used on mousedown to prevent text selection when multiple
 * entries are selected with Shift+click.
 * (tested in Firefox and Chromium)
 */
function preventTextSelection(event: MouseEvent | KeyboardEvent): void {
    if (event.shiftKey) {
        event.preventDefault()
    }
}

/**
 * Helper hook to determine whether a new entry has been added to the notepad.
 * Whenever the number of entries increases we have a new entry. It assumes that
 * the newest entry is the first element in the input array.
 */
function useHasNewEntry(entries: NotepadEntry[]): boolean {
    const previousLength = useRef<number>()

    const previous = previousLength.current
    previousLength.current = entries.length

    return previous !== undefined && previous < entries.length
}

export const NotepadIcon: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Icon aria-hidden={true} svgPath={mdiBookPlusOutline} />
)

export interface NotepadContainerProps {
    initialOpen?: boolean
    userId?: string
    isRepositoryRelatedPage?: boolean
}

export const NotepadContainer: React.FunctionComponent<React.PropsWithChildren<NotepadContainerProps>> = ({
    initialOpen,
    userId,
    isRepositoryRelatedPage,
}) => {
    const newEntry = useNotepadState(state => state.addableEntry)
    const entries = useNotepadState(state => state.entries)
    const canRestore = useNotepadState(state => state.canRestoreSession)
    const [enableNotepad] = useTemporarySetting('search.notepad.enabled')
    // Taken from global-styles/breakpoints.css , $viewport-md
    const isWideScreen = useMatchMedia('(min-width: 768px)')

    if (enableNotepad && isWideScreen) {
        return (
            <Notepad
                className={styles.fixed}
                initialOpen={initialOpen}
                newEntry={newEntry}
                entries={entries}
                userId={userId}
                restorePreviousSession={canRestore ? restorePreviousSession : undefined}
                addEntry={addNotepadEntry}
                removeEntry={removeFromNotepad}
                isRepositoryRelatedPage={isRepositoryRelatedPage}
            />
        )
    }

    return null
}
export interface NotepadProps {
    className?: string
    initialOpen?: boolean
    newEntry?: NotepadEntryInput | null
    entries: NotepadEntry[]
    addEntry: typeof addNotepadEntry
    removeEntry: (ids: NotepadEntryID[] | NotepadEntryID) => void
    restorePreviousSession?: () => void
    // This is only used in our CTA to prevent notes from being rendered as
    // selected
    selectable?: boolean
    userId?: string
    isRepositoryRelatedPage?: boolean
}

export const Notepad: React.FunctionComponent<React.PropsWithChildren<NotepadProps>> = ({
    className,
    initialOpen = false,
    entries,
    restorePreviousSession,
    addEntry,
    removeEntry,
    newEntry,
    selectable = true,
    userId,
    isRepositoryRelatedPage,
}) => {
    const navigate = useNavigate()

    const [open, setOpen] = useState(initialOpen)
    const [selectedEntries, setSelectedEntries] = useState<number[]>([])
    const isMacPlatform_ = useMemo(isMacPlatform, [])

    const reversedEntries = useMemo(() => [...entries].reverse(), [entries])
    const hasNewEntry = useHasNewEntry(reversedEntries)

    useLayoutEffect(() => {
        if (hasNewEntry && selectable) {
            // Always select the new entry. This is also avoids problems with
            // getting the selection index out of sync.
            setSelectedEntries([0])
        }
    }, [hasNewEntry, selectable])

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
            removeFromNotepad(entryIDs)
            // Clear selection for now.
            setSelectedEntries([])
        }
    }, [reversedEntries, selectedEntries, setSelectedEntries])

    const deleteEntry = useCallback(
        (toDelete: NotepadEntry) => {
            if (selectedEntries.length > 0) {
                const entryPosition = reversedEntries.findIndex(entry => entry.id === toDelete.id)
                setSelectedEntries(selection => adjustSelection(selection, entryPosition))
            }
            removeEntry([toDelete.id])
        },
        [reversedEntries, selectedEntries, setSelectedEntries, removeEntry]
    )

    const handleCreateNotebook = useCallback(() => {
        if (!userId) {
            return
        }

        const blocks: BlockInput[] = []
        for (const entry of entries) {
            if (entry.annotation) {
                blocks.push({ type: 'md', input: { text: entry.annotation } })
            }
            switch (entry.type) {
                case 'search': {
                    blocks.push({ type: 'query', input: { query: toSearchQuery(entry) } })
                    break
                }
                case 'file': {
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
        }

        createNotebook({
            notebook: {
                title: 'New Notebook',
                blocks: blocks.map(block => blockToGQLInput({ id: uuid.v4(), ...block })),
                public: false,
                namespace: userId,
            },
        })
            .toPromise()
            .then(createdNotebook => {
                navigate(EnterprisePageRoutes.Notebook.replace(':id', createdNotebook.id))
            })
            .catch(logger.error)
    }, [entries, userId, navigate])

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
            const hasMacMeta = isMacMetaKey(event, isMacPlatform_)
            const hasMeta = isMetaKey(event, isMacPlatform_)

            if (document.activeElement && document.activeElement.tagName === 'TEXTAREA') {
                // Ignore any events originating from an annotations input
                return
            }

            switch (event.key) {
                // Select all entries
                case 'a': {
                    if (hasMeta) {
                        // This prevents text selection
                        event.preventDefault()
                        setSelectedEntries(reversedEntries.map((_value, index) => index))
                    }
                    break
                }
                // Clear selection
                case 'Escape': {
                    if (selectedEntries.length > 0) {
                        setSelectedEntries([])
                    }
                    break
                }
                // Delete selected entries
                case 'Delete': {
                    if (selectedEntries.length > 0) {
                        deleteSelectedEntries()
                    }
                    break
                }
                // On macOS we also support CMD+Backpace for deletion
                case 'Backspace': {
                    if (hasMacMeta && selectedEntries.length > 0) {
                        deleteSelectedEntries()
                    }
                    break
                }
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
                                wrapPosition(selection.at(-1)! + (key === 'ArrowDown' ? 1 : -1), reversedEntries.length)
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

    // Focus the cancel button when the remove all confirmation box is shown
    const [confirmRemoveAll, setConfirmRemoveAll] = useState(false)
    const cancelRemoveAll = useRef<HTMLButtonElement>(null)
    useEffect(() => {
        if (confirmRemoveAll) {
            cancelRemoveAll.current?.focus()
        }
    }, [confirmRemoveAll])

    // Focus the remove all button when the remove all confirmation box is hidden.
    // If the remove all button is now disabled (because there are no entries left),
    // focus the top-level notepad button.
    const removeAllButton = useRef<HTMLButtonElement>(null)
    const rootButton = useRef<HTMLButtonElement>(null)
    const onRemoveAllClosed = useCallback((removeAll: boolean) => {
        setConfirmRemoveAll(false)

        if (removeAll) {
            removeAllNotepadEntries()
            rootButton.current?.focus()
        } else {
            removeAllButton.current?.focus()
        }
    }, [])

    // HACK: This is temporary fix for the overlapping Notepad icon until we either disable notepad
    //       or move Cody to the top level and mount the Notepad entrypoint inside it
    const { sidebarSize: codySidebarWidth } = useSidebarSize()

    return (
        <aside
            className={classNames(styles.root, className, { [styles.open]: open })}
            id={NOTEPAD_ID}
            aria-labelledby={`${NOTEPAD_ID}-button`}
            // eslint-disable-next-line react/forbid-dom-props
            style={{
                marginRight: isRepositoryRelatedPage ? codySidebarWidth : 0,
            }}
        >
            <Button
                variant="icon"
                className={classNames(styles.header, 'p-2 d-flex align-items-center justify-content-between')}
                onClick={toggleOpen}
                aria-controls={NOTEPAD_ID}
                aria-expanded="true"
                id={`${NOTEPAD_ID}-button`}
                ref={rootButton}
            >
                <span>
                    <NotepadIcon />
                    <H2 className="px-1 d-inline">Notepad</H2>
                    <small>
                        ({reversedEntries.length} note{reversedEntries.length === 1 ? '' : 's'})
                    </small>
                </span>
                <span className={styles.toggleIcon}>
                    <Icon aria-label={(open ? 'Close' : 'Open') + ' Notepad'} svgPath={mdiChevronUp} />
                </span>
            </Button>
            {open && (
                <>
                    {newEntry && (
                        <div className={classNames(styles.newNote, 'p-2')}>
                            <H3>Create new note from current {newEntry.type === 'file' ? 'file' : 'search'}:</H3>
                            <AddEntryButton entry={newEntry} addEntry={addEntry} />
                        </div>
                    )}
                    <H3 className="p-2">
                        Notes <small>({reversedEntries.length})</small>
                    </H3>

                    {/* This should be a role="listbox" and the entries should be role="option", but that doesn't work with the
                        design because of the nested buttons. Leave this as a normal list with some interaction (arrow keys, etc)
                        so that screen readers can navigate the nested controls.
                    */}
                    {/* eslint-disable-next-line jsx-a11y/no-noninteractive-element-interactions */}
                    <ol
                        onKeyDown={handleKey}
                        // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
                        tabIndex={0}
                        aria-label="Notepad entries. Use arrow keys to move selection. Use shift key to select multiple items."
                    >
                        {reversedEntries.map((entry, index) => {
                            const selected = selectedEntries.includes(index)
                            return (
                                // eslint-disable-next-line jsx-a11y/no-noninteractive-element-interactions
                                <li
                                    key={entry.id}
                                    data-notepad-entry-index={index}
                                    onClick={event => toggleSelectedEntry(index, event)}
                                    onKeyDown={event => {
                                        if (document.activeElement === event.currentTarget && event.key === ' ') {
                                            event.stopPropagation()
                                            toggleSelectedEntry(index, event)
                                        }
                                    }}
                                    onMouseDown={preventTextSelection}
                                    // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
                                    tabIndex={0}
                                    aria-label={getLabel(entry, selected)}
                                >
                                    <NotepadEntryComponent
                                        entry={entry}
                                        focus={hasNewEntry && index === 0}
                                        selected={selected}
                                        onDelete={deleteEntry}
                                        index={index}
                                    />
                                </li>
                            )
                        })}
                    </ol>
                    {confirmRemoveAll && (
                        <div className="p-2" role="alert">
                            <Text>Are you sure you want to delete all entries?</Text>
                            <div className="d-flex justify-content-between">
                                <Button
                                    variant="secondary"
                                    onClick={() => onRemoveAllClosed(false)}
                                    ref={cancelRemoveAll}
                                >
                                    Cancel
                                </Button>
                                <Button variant="danger" onClick={() => onRemoveAllClosed(true)}>
                                    Yes, delete
                                </Button>
                            </div>
                        </div>
                    )}
                    <div className="p-2 d-flex align-items-center">
                        <Button
                            onClick={handleCreateNotebook}
                            variant="primary"
                            size="sm"
                            disabled={entries.length === 0}
                            className="flex-1 mr-2"
                        >
                            Create Notebook
                        </Button>
                        {restorePreviousSession && (
                            <Button
                                className="mr-2"
                                onClick={restorePreviousSession}
                                outline={true}
                                variant="secondary"
                                size="sm"
                            >
                                Restore last session
                            </Button>
                        )}
                        <Button
                            aria-label="Remove all notes"
                            title="Remove all notes"
                            variant="icon"
                            className="text-muted"
                            disabled={entries.length === 0}
                            onClick={() => setConfirmRemoveAll(true)}
                            ref={removeAllButton}
                        >
                            <Icon aria-hidden={true} svgPath={mdiDelete} />
                        </Button>
                    </div>
                </>
            )}
        </aside>
    )
}

interface AddEntryButtonProps {
    entry: NotepadEntryInput
    addEntry: typeof addNotepadEntry
}

const AddEntryButton: React.FunctionComponent<React.PropsWithChildren<AddEntryButtonProps>> = ({ entry, addEntry }) => {
    let button: React.ReactElement
    switch (entry.type) {
        case 'search': {
            button = (
                <Button
                    outline={true}
                    variant="primary"
                    size="sm"
                    title="Add search"
                    className="w-100"
                    onClick={event => {
                        event.stopPropagation()
                        addEntry(entry)
                    }}
                >
                    <Icon aria-hidden={true} svgPath={mdiMagnify} /> Add search
                </Button>
            )
            break
        }
        case 'file': {
            button = (
                <span className="d-flex mx-0">
                    <Button
                        outline={true}
                        variant="primary"
                        size="sm"
                        title="Add file"
                        className={classNames({ 'flex-1': true, 'mr-1': !!entry.lineRange })}
                        onClick={event => {
                            event.stopPropagation()
                            addEntry(entry, 'file')
                        }}
                    >
                        <Icon aria-hidden={true} svgPath={mdiFileDocumentOutline} /> Add as file
                    </Button>
                    {entry.lineRange && (
                        <Button
                            outline={true}
                            variant="primary"
                            size="sm"
                            title="Add line range"
                            className="flex-1 ml-1"
                            onClick={event => {
                                event.stopPropagation()
                                addEntry(entry, 'range')
                            }}
                        >
                            <Icon aria-hidden={true} svgPath={mdiCodeBrackets} /> Add as range{' '}
                            {formatLineRange(entry.lineRange)}
                        </Button>
                    )}
                </span>
            )
        }
    }

    const { title } = getUIComponentsForEntry(entry)

    return (
        <>
            <div className={classNames(styles.entry, 'p-0 py-2')}>{title}</div>
            {button}
        </>
    )
}

function stopPropagation(event: SyntheticEvent): void {
    event.stopPropagation()
}

interface NotepadEntryComponentProps {
    entry: NotepadEntry
    /**
     * If set to true, show and focus the annotations input.
     */
    focus: boolean
    selected: boolean
    onDelete: (entry: NotepadEntry) => void
    index: number
}

const NotepadEntryComponent: React.FunctionComponent<React.PropsWithChildren<NotepadEntryComponentProps>> = ({
    entry,
    focus = false,
    selected,
    onDelete,
    index,
}) => {
    const { icon, title, location } = getUIComponentsForEntry(entry)
    const [annotation, setAnnotation] = useState(entry.annotation ?? '')
    const [showAnnotationInput, setShowAnnotationInput] = useState(focus)
    const textarea = useRef<HTMLTextAreaElement | null>(null)

    // Focus annotation input when the whenever it is opened.
    useEffect(() => {
        if (showAnnotationInput) {
            textarea.current?.focus()
        }
    }, [showAnnotationInput])

    // Focus entry when selected.
    useEffect(() => {
        if (selected) {
            const element = document.querySelector(`[data-notepad-entry-index="${index}"]`) as HTMLElement
            element?.focus()
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [selected])

    const toggleAnnotationInput = useCallback(
        (show: boolean) => {
            setShowAnnotationInput(show)
            if (!show) {
                // Persist the entry annotation when hiding the annotation input.
                setEntryAnnotation(entry, annotation)
            }
        },
        [entry, annotation, setShowAnnotationInput]
    )

    return (
        <div className={classNames(styles.entry, { [styles.selected]: selected })}>
            <div className="d-flex">
                <span className="sr-only">{selected ? 'Selected, ' : ''}</span>
                <span className="flex-shrink-0 text-muted mr-1">{icon}</span>
                <span className="flex-1">
                    <Link
                        to={typeof location === 'string' ? location : createLinkUrl(location)}
                        className="text-monospace search-query-link"
                    >
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
                            toggleAnnotationInput(!showAnnotationInput)
                        }}
                    >
                        <Icon aria-hidden={true} svgPath={mdiTextBox} />
                    </Button>
                    <Button
                        aria-label="Remove entry"
                        title="Remove entry"
                        variant="icon"
                        className="ml-1 text-muted"
                        onClick={event => {
                            event.stopPropagation()
                            onDelete(entry)
                        }}
                    >
                        <Icon aria-hidden={true} svgPath={mdiDelete} />
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
                    onKeyDown={event => {
                        switch (event.key) {
                            case 'Escape': {
                                event.currentTarget.blur()
                                break
                            }
                            case 'Enter': {
                                if (isMetaKey(event, isMacPlatform())) {
                                    toggleAnnotationInput(false)
                                }
                                break
                            }
                        }
                    }}
                />
            )}
        </div>
    )
}

function getUIComponentsForEntry(entry: NotepadEntry | NotepadEntryInput): {
    icon: React.ReactElement
    title: React.ReactElement
    location: LocationDescriptorObject | string
    body?: React.ReactElement
} {
    switch (entry.type) {
        case 'search': {
            return {
                icon: <Icon aria-label="Search" svgPath={mdiMagnify} />,
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
        }
        case 'file': {
            return {
                icon: (
                    <Icon
                        aria-label={entry.lineRange ? 'Line range' : 'File'}
                        svgPath={entry.lineRange ? mdiCodeBrackets : mdiFileDocumentOutline}
                    />
                ),
                title: (
                    <span title={entry.path}>
                        {fileName(entry.path)}
                        {entry.lineRange ? ` ${formatLineRange(entry.lineRange)}` : ''}
                    </span>
                ),
                location: toPrettyBlobURL({
                    repoName: entry.repo,
                    revision: entry.revision,
                    filePath: entry.path,
                    range: entry.lineRange
                        ? {
                              start: { line: entry.lineRange.startLine + 1, character: 0 },
                              end: { line: entry.lineRange.endLine + 1, character: 0 },
                          }
                        : undefined,
                }),
            }
        }
    }
}

function getLabel(entry: NotepadEntry, selected: boolean): string {
    const selectedText = selected ? 'Selected, ' : ''
    switch (entry.type) {
        case 'search': {
            return `${selectedText}search: ${toSearchQuery(entry)}`
        }
        case 'file': {
            if (entry.lineRange) {
                return `${selectedText}line range: ${fileName(entry.path)}${formatLineRange(entry.lineRange)}`
            }
            return `${selectedText}file: ${fileName(entry.path)}`
        }
    }
}

function toSearchQuery(entry: SearchEntry): string {
    let { query } = entry
    if (entry.patternType !== SearchPatternType.standard) {
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
    return parts.at(-1)!
}

function formatLineRange(lineRange: HighlightLineRange): string {
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

    const lastSelectedPosition = newSelection.at(-1)!
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
    let nextPosition = wrapPosition(selection.at(-1)! + delta, total)

    // Did we change direction and "deselected" the last position?
    // (it's enough to look at the penultimate selected position)
    if (selection.length > 1 && selection.at(-2) === nextPosition) {
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
