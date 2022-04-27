import React, { useCallback, useEffect, useState } from 'react'

import classNames from 'classnames'

import { ContentMatch, SearchMatch } from '@sourcegraph/shared/src/search/stream'

import styles from './SearchResultList.module.scss'

interface Props {
    results: SearchMatch[]
}

export const SearchResultList: React.FunctionComponent<Props> = ({ results }) => {
    const [selectedResult, setSelectedResult] = useState<null | string>(null)

    useEffect(() => {
        if (selectedResult === null) {
            setSelectedResult(getFirstResultKey(results))
        }
    }, [selectedResult, results])

    const onKeyDown = useCallback(
        (event: KeyboardEvent) => {
            const target = event.target as HTMLElement

            // We only want to handle keydown events on the search box
            if (target.nodeName !== 'TEXTAREA' || !target.className.includes('inputarea')) {
                return
            }

            // Ignore events when the autocomplete dropdown is open
            const isAutocompleteOpen = document.querySelector('.monaco-list.element-focused') !== null
            if (isAutocompleteOpen) {
                return
            }

            if ((event.key === 'ArrowDown' || event.key === 'ArrowUp') && selectedResult !== null) {
                // eslint-disable-next-line unicorn/prefer-query-selector
                const currentMatch = document.getElementById(`search-result-list-item-${selectedResult}`)
                if (currentMatch === null) {
                    return
                }

                if (
                    event.key === 'ArrowDown' &&
                    currentMatch.nextElementSibling &&
                    currentMatch.nextElementSibling.id
                ) {
                    setSelectedResult(currentMatch.nextElementSibling.id.replace('search-result-list-item-', ''))
                    currentMatch.nextElementSibling.scrollIntoView(false)
                    event.preventDefault()
                    event.stopPropagation()
                    return
                }

                if (
                    event.key === 'ArrowUp' &&
                    currentMatch.previousElementSibling &&
                    currentMatch.previousElementSibling.id
                ) {
                    setSelectedResult(currentMatch.previousElementSibling.id.replace('search-result-list-item-', ''))
                    currentMatch.previousElementSibling.scrollIntoView(false)
                    event.preventDefault()
                    event.stopPropagation()
                    return
                }
            }
        },
        [selectedResult]
    )

    useEffect(() => {
        document.addEventListener('keydown', onKeyDown, { capture: true })
        return () => document.removeEventListener('keydown', onKeyDown, { capture: true })
    })

    return (
        <div className={styles.list}>
            {results.map((match: SearchMatch) =>
                match.type === 'content'
                    ? match.lineMatches.map(line => {
                          const key = getKeyForLine(match, line)
                          const onClick = (): void => setSelectedResult(key)
                          return (
                              // The below element's accessibility is handled via a document level event listener.
                              //
                              // eslint-disable-next-line jsx-a11y/click-events-have-key-events,jsx-a11y/no-static-element-interactions
                              <div
                                  id={`search-result-list-item-${key}`}
                                  className={classNames(styles.listItem, {
                                      [styles.listItemActive]: key === selectedResult,
                                  })}
                                  onClick={onClick}
                                  key={key}
                              >
                                  {line.line} <small>{match.path}</small>
                              </div>
                          )
                      })
                    : null
            )}
        </div>
    )
}

function getFirstResultKey(results: SearchMatch[]): string | null {
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    const firstContentMatch: null | ContentMatch = results.find(result => result.type === 'content')
    if (firstContentMatch) {
        return getKeyForLine(firstContentMatch, firstContentMatch.lineMatches[0])
    }
    return null
}

function getKeyForLine(match: ContentMatch, line: ContentMatch['lineMatches'][0]): string {
    return `${match.repository}-${match.path}-${match.lineMatches.indexOf(line)}`
}
