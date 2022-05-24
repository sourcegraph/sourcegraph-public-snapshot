import React, { createRef, useCallback, useEffect, useMemo, useState } from 'react'

import { SearchMatch } from '@sourcegraph/shared/src/search/stream'

import { FileSearchResult } from './FileSearchResult'
import {
    getContentMatchId,
    getFirstResultId,
    getSearchResultElement,
    getSiblingResultElement,
    splitResultIdForContentMatch,
} from './utils'

import styles from './SearchResultList.module.scss'

interface Props {
    onPreviewChange: (result: SearchMatch, lineMatchIndex: number) => void
    onPreviewClear: () => void
    onOpen: (result: SearchMatch, lineMatchIndex: number) => void
    matches: SearchMatch[]
}

export const SearchResultList: React.FunctionComponent<Props> = ({ matches, onPreviewChange, onPreviewClear, onOpen }) => {
    const scrollViewReference = createRef<HTMLDivElement>()
    const [selectedResultId, setSelectedResultId] = useState<null | string>(null)

    const matchIdToMatchMap = useMemo((): Map<string, SearchMatch> => {
        const map = new Map<string, SearchMatch>()
        for (const match of matches) {
            if (match.type === 'content') {
                map.set(getContentMatchId(match), match)
            }
        }
        return map
    }, [matches])

    const selectResult = useCallback(
        (id: null | string) => {
            if (id !== null) {
                getSearchResultElement(id)?.scrollIntoView({ block: 'nearest', inline: 'nearest' })
                const [matchId, lineMatchIndex] = splitResultIdForContentMatch(id)
                const match = matchIdToMatchMap.get(matchId)
                if (match) {
                    onPreviewChange(match, lineMatchIndex)
                }
            } else {
                onPreviewClear()
            }
            setSelectedResultId(id)
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
                const [matchId, lineMatchIndex] = splitResultIdForContentMatch(selectedResultId)
                const match = matchIdToMatchMap.get(matchId)
                if (match) {
                    onOpen(match, lineMatchIndex)
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
                    case 'content':
                        return (
                            <FileSearchResult
                                key={`${match.repository}-${match.path}`}
                                match={match}
                                selectResult={selectResult}
                                selectedResult={selectedResultId}
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
