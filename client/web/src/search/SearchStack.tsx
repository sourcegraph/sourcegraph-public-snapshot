import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import SearchStackIcon from 'mdi-react/LayersSearchIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import React, { useCallback, useState } from 'react'
import { Link, useHistory } from 'react-router-dom'

import { RepoIcon } from '@sourcegraph/shared/src/components/RepoIcon'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { appendContextFilter, updateFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { buildSearchURLQuery, toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'
import { Button } from '@sourcegraph/wildcard'

import { SyntaxHighlightedSearchQuery } from '../components/SyntaxHighlightedSearchQuery'
import { PageRoutes } from '../routes.constants'
import { useExperimentalFeatures } from '../stores'
import { useSearchStackState, restorePreviousSession, SearchEntry, SearchStackEntry } from '../stores/searchStack'

import { BlockInput } from './notebook'
import { serializeBlocks } from './notebook/serialize'
import styles from './SearchStack.module.scss'

export const SearchStack: React.FunctionComponent<{ initialOpen?: boolean }> = ({ initialOpen = false }) => {
    const history = useHistory()

    const [open, setOpen] = useState(initialOpen)
    const entries = useSearchStackState(state => state.entries)
    const canRestore = useSearchStackState(state => state.canRestoreSession)
    const enableSearchStack = useExperimentalFeatures(features => features.enableSearchStack)

    const createNotebook = useCallback(() => {
        const location = {
            pathname: PageRoutes.SearchNotebook,
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

    if (!enableSearchStack || (entries.length === 0 && !canRestore)) {
        return null
    }

    return (
        <div className={classNames(styles.root, { [styles.open]: open })}>
            <div className={classNames(styles.header, 'd-flex align-items-center justify-content-between')}>
                <Button
                    aria-label={`${open ? 'Close' : 'Open'} search session`}
                    className={classNames('btn-icon p-2')}
                    onClick={() => setOpen(open => !open)}
                >
                    <SearchStackIcon className="icon-inline" />
                    <h4 className={classNames(styles.openVisible, 'pl-1')}>Search session</h4>
                </Button>
                <Button
                    aria-label="Close search session"
                    className={classNames('btn-icon pr-2', styles.closeButton, styles.openVisible)}
                    onClick={() => setOpen(false)}
                >
                    <CloseIcon className="icon-inline" />
                </Button>
            </div>
            {open && (
                <>
                    <ul>
                        {entries.map((entry, index) => (
                            <li key={index}>{renderSearchEntry(entry)}</li>
                        ))}
                    </ul>
                    {(canRestore || entries.length > 0) && (
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
                            {entries.length > 0 && (
                                <Button
                                    className="w-100"
                                    onClick={createNotebook}
                                    outline={true}
                                    variant="secondary"
                                    size="sm"
                                >
                                    Create Notebook
                                </Button>
                            )}
                        </div>
                    )}
                </>
            )}
        </div>
    )
}

function renderSearchEntry(entry: SearchStackEntry): React.ReactChild {
    switch (entry.type) {
        case 'search':
            return (
                <Link
                    to={{
                        pathname: '/search',
                        search: buildSearchURLQuery(
                            entry.query,
                            entry.patternType,
                            entry.caseSensitive,
                            entry.searchContext
                        ),
                    }}
                    className={styles.entry}
                >
                    <SearchIcon className="icon-inline text-muted mr-1" />
                    <SyntaxHighlightedSearchQuery query={entry.query} />
                </Link>
            )
        case 'file':
            return (
                <Link
                    to={{
                        pathname: toPrettyBlobURL({
                            repoName: entry.repo,
                            revision: entry.revision,
                            filePath: entry.path,
                        }),
                    }}
                    className={styles.entry}
                >
                    <div>
                        <FileDocumentIcon className="icon-inline text-muted mr-1" />
                        <span title={entry.path}>{shortenFilePath(entry.path)}</span>
                    </div>
                    <small className="text-muted">
                        <RepoIcon repoName={entry.repo} className="icon-inline text-muted mr-1" />
                        {entry.repo}
                    </small>
                </Link>
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

/**
 * This function takes a file path and shortens any path segment to the first
 * character, except for the first and last segment and any segment that
 * contains less than five characters.
 *
 * Example: path/to/deeply/nested/file => path/to/d/n/file
 */
function shortenFilePath(path: string): string {
    const parts = path.split('/')
    if (parts.length === 1) {
        return path
    }
    return [parts[0]]
        .concat(
            parts.slice(1, -1).map(part => (part.length < 5 ? part : part[0])),
            parts[parts.length - 1]
        )
        .join('/')
}
