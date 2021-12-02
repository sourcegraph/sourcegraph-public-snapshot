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

import { SyntaxHighlightedSearchQuery } from '../components/SyntaxHighlightedSearchQuery'
import { PageRoutes } from '../routes.constants'
import { useExperimentalFeatures } from '../stores'
import { SearchEntry, FileEntry, useSearchStackState, restorePreviousSession } from '../stores/searchStack'

import { BlockInput } from './notebook'
import { serializeBlocks } from './notebook/helpers'
import styles from './SearchStack.module.scss'

export const SearchStack: React.FunctionComponent = () => {
    const history = useHistory()

    const [open, setOpen] = useState(false)
    const entries = useSearchStackState(state => state.entries)
    const canRestore = useSearchStackState(state => state.canRestoreSession)
    const enableSearchStack = useExperimentalFeatures(features => features.enableSearchStack)

    const createNotebook = useCallback(() => {
        // Show searches first
        const sortedEntries = [...entries].sort((entryA, entryB) => {
            if (entryA.type === entryB.type) {
                return 0
            }
            if (entryA.type === 'search') {
                return -1
            }
            return 1
        })
        const location = {
            pathname: PageRoutes.SearchNotebook,
            hash: serializeBlocks(
                sortedEntries
                    .map(
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
                                            lineRange: null,
                                        },
                                    }
                            }
                        }
                    )
                    .filter(Boolean)
            ),
        }
        history.push(location)
    }, [entries, history])

    if (!enableSearchStack || (entries.length === 0 && !canRestore)) {
        return null
    }

    const searches: SearchEntry[] = []
    const files: FileEntry[] = []

    for (const entry of entries) {
        switch (entry.type) {
            case 'search':
                searches.push(entry)
                break
            case 'file':
                files.push(entry)
                break
        }
    }

    return (
        <div className={classNames(styles.root, { [styles.open]: open })}>
            <div className={classNames(styles.header, 'd-flex align-items-center justify-content-between')}>
                <button
                    type="button"
                    aria-label="Close alert"
                    className={classNames('btn btn-icon p-2')}
                    onClick={() => setOpen(open => !open)}
                >
                    <SearchStackIcon className="icon-inline" />
                    <h4 className={classNames(styles.openVisible, 'pl-1')}>Search Stack</h4>
                </button>
                <button
                    type="button"
                    aria-label="Close search stack"
                    className={classNames('btn btn-icon pr-2', styles.closeButton, styles.openVisible)}
                    onClick={() => setOpen(false)}
                >
                    <CloseIcon className="icon-inline" />
                </button>
            </div>
            {open && (
                <>
                    {searches.length > 0 && (
                        <details open={true}>
                            <summary>Recent searches</summary>
                            <ul>
                                {searches.map((entry, index) => (
                                    <li key={index}>
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
                                        >
                                            <SearchIcon className="icon-inline text-muted" />{' '}
                                            <SyntaxHighlightedSearchQuery query={entry.query} />
                                        </Link>
                                    </li>
                                ))}
                            </ul>
                        </details>
                    )}
                    {files.length > 0 && (
                        <details open={true}>
                            <summary>Recent files</summary>
                            <ul>
                                {files.map((entry, index) => (
                                    <li key={index}>
                                        <div title={entry.path}>
                                            <Link
                                                to={{
                                                    pathname: toPrettyBlobURL({
                                                        repoName: entry.repo,
                                                        revision: entry.revision,
                                                        filePath: entry.path,
                                                    }),
                                                }}
                                            >
                                                <FileDocumentIcon className="icon-inline text-muted" />
                                                &nbsp;{shortenFilePath(entry.path)}
                                            </Link>
                                        </div>
                                        <small className="text-muted">
                                            <RepoIcon repoName={entry.repo} className="icon-inline text-muted" />
                                            {entry.repo}
                                        </small>
                                    </li>
                                ))}
                            </ul>
                        </details>
                    )}
                    {(canRestore || entries.length > 0) && (
                        <>
                            <div className="p-2">
                                {canRestore && (
                                    <button
                                        type="button"
                                        className="w-100 btn btn-sm btn-outline-secondary mb-1"
                                        onClick={restorePreviousSession}
                                    >
                                        Restore previous session
                                    </button>
                                )}
                                {entries.length > 0 && (
                                    <button
                                        type="button"
                                        className="w-100 btn btn-sm btn-outline-secondary"
                                        onClick={createNotebook}
                                    >
                                        Create Notebook
                                    </button>
                                )}
                            </div>
                        </>
                    )}
                </>
            )}
        </div>
    )
}

function toSearchQuery(entry: SearchEntry): string {
    let { query } = entry
    console.debug(entry)
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
