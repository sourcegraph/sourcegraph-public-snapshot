import classNames from 'classnames'
import { LocationDescriptor } from 'history'
import CloseIcon from 'mdi-react/CloseIcon'
import CodeBracketsIcon from 'mdi-react/CodeBracketsIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import SearchStackIcon from 'mdi-react/LayersSearchIcon'
import NotebookPlusIcon from 'mdi-react/NotebookPlusIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import TrashIcon from 'mdi-react/TrashCanIcon'
import React, { useCallback, useState, useMemo } from 'react'
import { useHistory } from 'react-router-dom'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { RepoIcon } from '@sourcegraph/shared/src/components/RepoIcon'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { IHighlightLineRange } from '@sourcegraph/shared/src/schema'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { appendContextFilter, updateFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { buildSearchURLQuery, toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'
import { Button, Link } from '@sourcegraph/wildcard'

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
        const location = {
            pathname: PageRoutes.NotebookCreate,
            hash: serializeBlocks(
                entries.map(
                    (entry): BlockInput => {
                        switch (entry.type) {
                            case 'search':
                                return { type: 'query', input: toSearchQuery(entry) }
                            case 'file':
                                return {
                                    type: 'file',
                                    input: {
                                        repositoryName: entry.repo,
                                        revision: entry.revision,
                                        filePath: entry.path,
                                        lineRange: entry.lineRange,
                                    },
                                }
                        }
                    }
                ),
                window.location.origin
            ),
        }
        history.push(location)
    }, [entries, history])

    if (!enableSearchStack) {
        return null
    }

    return (
        <div className={classNames(styles.root, { [styles.open]: open })}>
            <div className={classNames(styles.header, 'd-flex align-items-center justify-content-between')}>
                <Button
                    aria-label={`${open ? 'Close' : 'Open'} search session`}
                    variant="icon"
                    className="p-2"
                    onClick={() => setOpen(open => !open)}
                >
                    <SearchStackIcon className="icon-inline" />
                    <h4 className={classNames(styles.openVisible, 'pl-1')}>Search session</h4>
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
            {open && (
                <>
                    {addableEntry && <AddEntryButton entry={addableEntry} />}
                    <ul>
                        {reversedEntries.map(entry => (
                            <li key={entry.id}>{renderSearchEntry(entry)}</li>
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
                            <Button
                                onClick={createNotebook}
                                variant="primary"
                                size="sm"
                                disabled={entries.length === 0}
                            >
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
                </>
            )}
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
                    className="m-3"
                    onClick={() => addSearchStackEntry(entry)}
                >
                    + <SearchIcon className="icon-inline" /> Search
                </Button>
            )
        case 'file':
            return (
                <span className="d-flex m-2">
                    <Button
                        variant="primary"
                        size="sm"
                        title="Add file"
                        className="flex-1 m-1"
                        onClick={() => addSearchStackEntry(entry, 'file')}
                    >
                        + <FileDocumentIcon className="icon-inline" /> File
                    </Button>
                    {entry.lineRange && (
                        <Button
                            variant="primary"
                            size="sm"
                            title="Add line range"
                            className="flex-1 m-1"
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

interface SearchStackEntryComponentProps {
    entry: SearchStackEntry
    icon: React.ReactElement
    title: React.ReactElement
    location: LocationDescriptor<any>
    children?: React.ReactElement
}

const SearchStackEntryComponent: React.FunctionComponent<SearchStackEntryComponentProps> = ({
    icon,
    title,
    location,
    children,
    entry,
}) => (
    <div className={styles.entry}>
        <div className="d-flex">
            <span className="flex-shrink-0 text-muted mr-1">{icon}</span>
            <Link to={location} className={classNames(styles.entry, 'flex-1 p-0')}>
                {title}
            </Link>
            <span className="ml-1">
                <Button
                    aria-label="Remove entry"
                    title="Remove entry"
                    variant="icon"
                    className="text-muted"
                    onClick={() => removeSearchStackEntry(entry)}
                >
                    <CloseIcon className="icon-inline" />
                </Button>
            </span>
        </div>
        {children}
    </div>
)

function renderSearchEntry(entry: SearchStackEntry): React.ReactChild {
    switch (entry.type) {
        case 'search':
            return (
                <SearchStackEntryComponent
                    entry={entry}
                    icon={<SearchIcon className="icon-inline" />}
                    title={<SyntaxHighlightedSearchQuery query={entry.query} />}
                    location={{
                        pathname: '/search',
                        search: buildSearchURLQuery(
                            entry.query,
                            entry.patternType,
                            entry.caseSensitive,
                            entry.searchContext
                        ),
                    }}
                />
            )
        case 'file':
            return (
                <SearchStackEntryComponent
                    entry={entry}
                    icon={
                        entry.lineRange ? (
                            <CodeBracketsIcon className="icon-inline" />
                        ) : (
                            <FileDocumentIcon className="icon-inline" />
                        )
                    }
                    title={
                        <span title={entry.path}>
                            {fileName(entry.path)}
                            {entry.lineRange ? ` ${formatLineRange(entry.lineRange)}` : ''}
                        </span>
                    }
                    location={{
                        pathname: toPrettyBlobURL({
                            repoName: entry.repo,
                            revision: entry.revision,
                            filePath: entry.path,
                        }),
                    }}
                >
                    <small className="text-muted">
                        <RepoIcon repoName={entry.repo} className="icon-inline text-muted mr-1" />
                        {entry.repo}
                    </small>
                </SearchStackEntryComponent>
            )
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
