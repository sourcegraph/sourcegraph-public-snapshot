import classNames from 'classnames'
import LanguageTypescriptIcon from 'mdi-react/LanguageTypescriptIcon'
import React, { useEffect, useMemo, useState } from 'react'
import { useHistory } from 'react-router'
import { Subject } from 'rxjs'
import { throttleTime } from 'rxjs/operators'

import { Position, Range, WorkspaceRoot } from '@sourcegraph/extension-api-types'
import { useDebounce } from '@sourcegraph/wildcard'

import { LanguageIcon } from '../../../components/languageIcons'
import { HighlightedLink } from '../../../fuzzyFinder/components/HighlightedLink'
import { downloadFilenames, FuzzyFinderProps, FuzzyFSM } from '../../../fuzzyFinder/fsm'
import { FuzzyModalProps, FuzzyModalState, renderFuzzyResult } from '../../../fuzzyFinder/FuzzyModal'
import { getModeFromPath } from '../../../languages'
import { PlatformContextProps } from '../../../platform/context'
import { ParsedRepoURI, parseQueryAndHash, parseRepoURI } from '../../../util/url'
import { useObservable } from '../../../util/useObservable'

import { Message } from './Message'
import { NavigableList } from './NavigableList'
import styles from './NavigableList.module.scss'

interface FuzzyFinderResultProps extends PlatformContextProps<'requestGraphQL' | 'urlToFile' | 'clientApplication'> {
    value: string
    onClick: () => void
    workspaceRoot: WorkspaceRoot | undefined
}

// Note: filenames don't have to be in monospace font.

export const FuzzyFinderResult: React.FC<FuzzyFinderResultProps> = ({
    value,
    onClick,
    workspaceRoot,
    platformContext,
}) => {
    const history = useHistory()

    // The state machine of the fuzzy finder. See `FuzzyFSM` for more details
    // about the state transititions.
    const [fsm, setFsm] = useState<FuzzyFSM>({ key: 'empty' })

    // TODO: use for infinite scrolling list
    const [maxResults, setMaxResults] = useState(100)

    const repoUrl = useMemo(() => {
        if (!workspaceRoot?.uri) {
            return null
        }

        return platformContext.clientApplication === 'sourcegraph'
            ? parseBrowserRepoURL(location.pathname + location.search + location.hash)
            : parseRepoURI(workspaceRoot.uri)
    }, [workspaceRoot?.uri, platformContext.clientApplication])

    // TODO: run fuzzy search in web worker
    // const debouncedValue = useDebounce(value, 300)

    const valueUpdates = useMemo(() => new Subject<string>(), [])
    useEffect(() => {
        valueUpdates.next(value)
    }, [value, valueUpdates])
    const throttledValue =
        useObservable(
            useMemo(() => valueUpdates.pipe(throttleTime(300, undefined, { leading: true, trailing: true })), [
                valueUpdates,
            ])
        ) ?? value // initial value

    const fuzzyResult = useMemo(() => {
        const commitID = repoUrl?.commitID ?? workspaceRoot?.inputRevision ?? ''
        const repoName = repoUrl?.repoName ?? ''

        const fuzzyFinderProps: FuzzyFinderProps = {
            repoName,
            commitID,
            platformContext,
        }

        const state: FuzzyModalState = {
            query: throttledValue,
            maxResults,
        }
        const fuzzyModalProps: Pick<
            FuzzyModalProps,
            | 'fsm'
            | 'setFsm'
            | 'downloadFilenames'
            | 'parseRepoUrl'
            | 'onClose'
            | 'repoName'
            | 'commitID'
            | 'platformContext'
        > = {
            fsm,
            setFsm,
            parseRepoUrl: () => ({ ...repoUrl, commitID, repoName }), // TODO error handling,
            downloadFilenames: () => downloadFilenames(fuzzyFinderProps),
            commitID,
            repoName,
            onClose: onClick,
            platformContext,
        }
        return renderFuzzyResult(fuzzyModalProps, state)
    }, [throttledValue, onClick, fsm, platformContext, repoUrl, workspaceRoot?.inputRevision, maxResults])

    if (!workspaceRoot) {
        return <Message type="muted">Navigate to a repo to use fuzzy finder</Message>
    }

    const onFuzzyResultClick = (url?: string): void => {
        if (url) {
            if (platformContext.clientApplication === 'sourcegraph') {
                history.push(url)
            } else {
                window.location.href = 'https://sourcegraph.test:3443' + url
            }
        }
    }

    return (
        <>
            {fuzzyResult.element && <Message>{fuzzyResult.element}</Message>}
            <NavigableList items={fuzzyResult.linksToRender}>
                {(file, { active }) => (
                    <NavigableList.Item
                        key={file.text}
                        active={active}
                        onClick={() => {
                            onFuzzyResultClick(file.url)
                            file.onClick?.()
                        }}
                    >
                        <span role="link" data-href={file.url ?? ''} className={styles.itemContainer}>
                            <LanguageIcon
                                language={getModeFromPath(file.text)}
                                className={classNames(styles.itemIcon, 'icon-inline')}
                            />
                            <HighlightedLink
                                {...{
                                    ...file,
                                    // Opt out of link rendering
                                    url: undefined,
                                }}
                            />
                        </span>
                    </NavigableList.Item>
                )}
            </NavigableList>
        </>
    )
}

// function useFuzzyFSM() {}

// TODO move to util, copied from web
/**
 * Parses the properties of a blob URL.
 */
export function parseBrowserRepoURL(href: string): ParsedRepoURI & Pick<ParsedRepoRevision, 'rawRevision'> {
    const url = new URL(href, window.location.href)
    let pathname = url.pathname.slice(1) // trim leading '/'
    if (pathname.endsWith('/')) {
        pathname = pathname.slice(0, -1) // trim trailing '/'
    }

    const indexOfSeparator = pathname.indexOf('/-/')

    // examples:
    // - 'github.com/gorilla/mux'
    // - 'github.com/gorilla/mux@revision'
    // - 'foo/bar' (from 'sourcegraph.mycompany.com/foo/bar')
    // - 'foo/bar@revision' (from 'sourcegraph.mycompany.com/foo/bar@revision')
    // - 'foobar' (from 'sourcegraph.mycompany.com/foobar')
    // - 'foobar@revision' (from 'sourcegraph.mycompany.com/foobar@revision')
    let repoRevision: string
    if (indexOfSeparator === -1) {
        repoRevision = pathname // the whole string
    } else {
        repoRevision = pathname.slice(0, indexOfSeparator) // the whole string leading up to the separator (allows revision to be multiple path parts)
    }
    const { repoName, revision, rawRevision } = parseRepoRevision(repoRevision)
    if (!repoName) {
        throw new Error('unexpected repo url: ' + href)
    }
    const commitID = revision && /^[\da-f]{40}$/i.test(revision) ? revision : undefined

    let filePath: string | undefined
    let commitRange: string | undefined
    const treeSeparator = pathname.indexOf('/-/tree/')
    const blobSeparator = pathname.indexOf('/-/blob/')
    const comparisonSeparator = pathname.indexOf('/-/compare/')
    if (treeSeparator !== -1) {
        filePath = decodeURIComponent(pathname.slice(treeSeparator + '/-/tree/'.length))
    }
    if (blobSeparator !== -1) {
        filePath = decodeURIComponent(pathname.slice(blobSeparator + '/-/blob/'.length))
    }
    if (comparisonSeparator !== -1) {
        commitRange = pathname.slice(comparisonSeparator + '/-/compare/'.length)
    }
    let position: Position | undefined
    let range: Range | undefined

    const parsedHash = parseQueryAndHash(url.search, url.hash)
    if (parsedHash.line) {
        position = {
            line: parsedHash.line,
            character: parsedHash.character || 0,
        }
        if (parsedHash.endLine) {
            range = {
                start: position,
                end: {
                    line: parsedHash.endLine,
                    character: parsedHash.endCharacter || 0,
                },
            }
        }
    }
    return { repoName, revision, rawRevision, commitID, filePath, commitRange, position, range }
}

/** The results of parsing a repo-revision string like "my/repo@my/revision". */
export interface ParsedRepoRevision {
    repoName: string

    /** The URI-decoded revision (e.g., "my#branch" in "my/repo@my%23branch"). */
    revision?: string

    /** The raw revision (e.g., "my%23branch" in "my/repo@my%23branch"). */
    rawRevision?: string
}

/**
 * Parses a repo-revision string like "my/repo@my/revision" to the repo and revision components.
 */
export function parseRepoRevision(repoRevision: string): ParsedRepoRevision {
    const [repository, revision] = repoRevision.split('@', 2) as [string, string | undefined]
    return {
        repoName: decodeURIComponent(repository),
        revision: revision && decodeURIComponent(revision),
        rawRevision: revision,
    }
}
