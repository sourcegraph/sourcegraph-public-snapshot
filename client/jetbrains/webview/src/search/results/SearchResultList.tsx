import React, { createRef, useCallback, useEffect, useMemo, useState } from 'react'

import { ContentMatch, SearchMatch } from '@sourcegraph/shared/src/search/stream'

import { FileSearchResult } from './FileSearchResult'
import { decodeLineId, getElementFromId, getFirstResultId, getIdForMatch, getSiblingResult } from './utils'

import styles from './SearchResultList.module.scss'

interface Props {
    onPreviewChange: (result: ContentMatch, lineMatchIndex: number) => void
    onPreviewClear: () => void
    onOpen: (result: ContentMatch, lineMatchIndex: number) => void
    results: SearchMatch[]
}

export const SearchResultList: React.FunctionComponent<Props> = ({
    results,
    onPreviewChange,
    onPreviewClear,
    onOpen,
}) => {
    const scrollViewReference = createRef<HTMLDivElement>()
    const [selectedResult, setSelectedResult] = useState<null | string>(null)

    const resultMap = useMemo((): Map<string, ContentMatch> => {
        const map = new Map<string, ContentMatch>()
        for (const result of results) {
            if (result.type === 'content') {
                map.set(getIdForMatch(result), result)
            }
        }
        return map
    }, [results])

    const selectResultFromId = useCallback(
        (id: null | string) => {
            if (id !== null) {
                getElementFromId(id)?.scrollIntoView({ block: 'nearest', inline: 'nearest' })
                const [matchId, lineMatchIndex] = decodeLineId(id)
                const match = resultMap.get(matchId)
                if (match) {
                    onPreviewChange(match, lineMatchIndex)
                }
            } else {
                onPreviewClear()
            }
            setSelectedResult(id)
        },
        [onPreviewChange, onPreviewClear, resultMap]
    )

    useEffect(() => {
        if (selectedResult === null) {
            selectResultFromId(getFirstResultId(results))
        }
    }, [selectedResult, results, selectResultFromId])

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

            if (selectedResult === null) {
                return
            }

            const currentElement = getElementFromId(selectedResult)
            if (currentElement === null) {
                return
            }

            if (event.key === 'Enter' && event.altKey) {
                const [matchId, lineMatchIndex] = decodeLineId(selectedResult)
                const match = resultMap.get(matchId)
                if (match) {
                    onOpen(match, lineMatchIndex)
                }
                return
            }

            if (event.key === 'ArrowDown') {
                const next = getSiblingResult(currentElement, 'next')
                if (next) {
                    selectResultFromId(next)
                }
                event.preventDefault()
                event.stopPropagation()
                return
            }

            if (event.key === 'ArrowUp') {
                const previous = getSiblingResult(currentElement, 'previous')
                if (previous) {
                    selectResultFromId(previous)
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
        [selectedResult, resultMap, onOpen, selectResultFromId, scrollViewReference]
    )

    useEffect(() => {
        document.addEventListener('keydown', onKeyDown, { capture: true })
        return () => document.removeEventListener('keydown', onKeyDown, { capture: true })
    })

    return (
        <div className={styles.list} ref={scrollViewReference}>
            {results.map((match: SearchMatch) =>
                match.type === 'content' ? (
                    <FileSearchResult
                        key={`${match.repository}-${match.path}`}
                        result={match}
                        selectResultFromId={selectResultFromId}
                        selectedResult={selectedResult}
                    />
                ) : null
            )}
        </div>
    )
}
