import React, { createRef, useCallback, useEffect, useState } from 'react'

import { SearchMatch } from '@sourcegraph/shared/src/search/stream'

import { FileSearchResult } from './FileSearchResult'
import { getElementFromId, getFirstResultId, getNextResult, getPreviousResult } from './utils'

import styles from './SearchResultList.module.scss'

interface Props {
    results: SearchMatch[]
}

export const SearchResultList: React.FunctionComponent<Props> = ({ results }) => {
    const scrollViewReference = createRef<HTMLDivElement>()
    const [selectedResult, setSelectedResult] = useState<null | string>(null)

    console.log({ selectedResult })
    useEffect(() => {
        if (selectedResult === null) {
            setSelectedResult(getFirstResultId(results))
        }
    }, [selectedResult, results])

    const selectResultFromId = useCallback((id: string) => {
        setSelectedResult(id)
        getElementFromId(id)?.scrollIntoView({ block: 'nearest', inline: 'nearest' })
    }, [])

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

            if (event.key === 'ArrowDown') {
                const next = getNextResult(currentElement)
                if (next) {
                    selectResultFromId(next)
                }
                event.preventDefault()
                event.stopPropagation()
                return
            }

            if (event.key === 'ArrowUp') {
                const previous = getPreviousResult(currentElement)
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
        [selectResultFromId, selectedResult, scrollViewReference]
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
