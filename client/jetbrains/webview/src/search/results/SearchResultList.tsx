import React, { createRef, useCallback, useEffect, useMemo, useState } from 'react'

import { SearchMatch } from '@sourcegraph/shared/src/search/stream'

import { CommitSearchResult } from './CommitSearchResult'
import { FileSearchResult } from './FileSearchResult'
import { PathSearchResult } from './PathSearchResult'
import { RepoSearchResult } from './RepoSearchResult'
import {
    getFirstResultId,
    getLineMatchIndexForContentMatch,
    getMatchId,
    getMatchIdForResult,
    getSearchResultElement,
    getSiblingResultElement,
} from './utils'

import styles from './SearchResultList.module.scss'

interface Props {
    onPreviewChange: (match: SearchMatch, lineMatchIndex?: number) => void
    onPreviewClear: () => void
    onOpen: (result: SearchMatch, lineMatchIndex?: number) => void
    matches: SearchMatch[]
}

export const SearchResultList: React.FunctionComponent<Props> = ({
    matches,
    onPreviewChange,
    onPreviewClear,
    onOpen,
}) => {
    const scrollViewReference = createRef<HTMLDivElement>()
    const [selectedResultId, setSelectedResultId] = useState<null | string>(null)

    const matchIdToMatchMap = useMemo((): Map<string, SearchMatch> => {
        const map = new Map<string, SearchMatch>()
        for (const match of matches) {
            if (['commit', 'content', 'path', 'repo'].includes(match.type)) {
                map.set(getMatchId(match), match)
            }
        }
        return map
    }, [matches])

    const selectResult = useCallback(
        (resultId: null | string) => {
            if (resultId !== null) {
                getSearchResultElement(resultId)?.scrollIntoView({ block: 'nearest', inline: 'nearest' })
                const matchId = getMatchIdForResult(resultId)
                const match = matchIdToMatchMap.get(matchId)
                if (match) {
                    onPreviewChange(
                        match,
                        match.type === 'content' ? getLineMatchIndexForContentMatch(resultId) : undefined
                    )
                } else {
                    console.log(`No match found for result id: ${resultId}`)
                }
            } else {
                onPreviewClear()
            }
            setSelectedResultId(resultId)
        },
        [onPreviewChange, onPreviewClear, matchIdToMatchMap]
    )

    useEffect(() => {
        if (selectedResultId === null) {
            selectResult(getFirstResultId(matches))
        }
    }, [selectedResultId, matches, selectResult])

    const onKeyDown = useCallback(
        (event: KeyboardEvent) => {
            const target = event.target as HTMLElement

            // We only want to handle keydown events on the search box
            if (
                (target.nodeName !== 'TEXTAREA' || !target.className.includes('inputarea')) &&
                target.nodeName !== 'BODY'
            ) {
                return
            }

            // Ignore events when the autocomplete dropdown is open
            const isAutocompleteOpen = document.querySelector('.monaco-list.element-focused') !== null
            if (isAutocompleteOpen) {
                return
            }

            if (selectedResultId === null) {
                return
            }

            const currentElement = getSearchResultElement(selectedResultId)
            if (currentElement === null) {
                return
            }

            if (event.key === 'Enter' && event.altKey) {
                const matchId = getMatchIdForResult(selectedResultId)
                const match = matchIdToMatchMap.get(matchId)
                if (match) {
                    onOpen(
                        match,
                        match.type === 'content' ? getLineMatchIndexForContentMatch(selectedResultId) : undefined
                    )
                }
                return
            }

            if (event.key === 'ArrowDown') {
                const nextElement = getSiblingResultElement(currentElement, 'next')
                if (nextElement) {
                    selectResult(nextElement.id.replace('search-result-list-item-', ''))
                }
                event.preventDefault()
                event.stopPropagation()
                return
            }

            if (event.key === 'ArrowUp') {
                const previousElement = getSiblingResultElement(currentElement, 'previous')
                if (previousElement) {
                    selectResult(previousElement.id.replace('search-result-list-item-', ''))
                } else if (scrollViewReference.current) {
                    // If we're already at the first item, we want to set the scroll view to the top
                    // so that the user can see the header.
                    scrollViewReference.current.scrollTop = 0
                }
                event.preventDefault()
                event.stopPropagation()
                return
            }
        },
        [selectedResultId, matchIdToMatchMap, onOpen, selectResult, scrollViewReference]
    )

    useEffect(() => {
        document.addEventListener('keydown', onKeyDown, { capture: true })
        return () => document.removeEventListener('keydown', onKeyDown, { capture: true })
    })

    return (
        <div className={styles.list} ref={scrollViewReference}>
            {matches.map((match: SearchMatch) => {
                switch (match.type) {
                    case 'commit':
                        return (
                            <CommitSearchResult
                                key={`${match.repository}-${match.url}`}
                                match={match}
                                selectedResult={selectedResultId}
                                selectResult={selectResult}
                            />
                        )
                    case 'content':
                        return (
                            <FileSearchResult
                                key={`${match.repository}-${match.path}`}
                                match={match}
                                selectedResult={selectedResultId}
                                selectResult={selectResult}
                            />
                        )
                    case 'repo':
                        return (
                            <RepoSearchResult
                                key={`${match.repository}`}
                                match={match}
                                selectedResult={selectedResultId}
                                selectResult={selectResult}
                            />
                        )
                    case 'path':
                        return (
                            <PathSearchResult
                                key={`${match.repository}-${match.path}`}
                                match={match}
                                selectedResult={selectedResultId}
                                selectResult={selectResult}
                            />
                        )
                    // TODO: Add more types
                    default:
                        console.log('Unknown search result type:', match.type)
                        return null
                }
            })}
        </div>
    )
}
