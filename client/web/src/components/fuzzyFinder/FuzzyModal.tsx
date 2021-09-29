import Dialog from '@reach/dialog'
import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useState } from 'react'

import { HighlightedLink } from '@sourcegraph/shared/src/fuzzyFinder/components/HighlightedLink'
import { FuzzyModalProps, FuzzyResultsSummary, renderFuzzyResult } from '@sourcegraph/shared/src/fuzzyFinder/FuzzyModal'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import styles from './FuzzyModal.module.scss'

const FUZZY_MODAL_TITLE = 'fuzzy-modal-title'
const FUZZY_MODAL_RESULTS = 'fuzzy-modal-results'

// The number of results to jump by on PageUp/PageDown keyboard shortcuts.
const PAGE_DOWN_INCREMENT = 10

interface FuzzyModalState {
    query: string
    setQuery: (query: string) => void

    focusIndex: number
    setFocusIndex: (focusIndex: number) => void

    maxResults: number
    increaseMaxResults: () => void
}

/**
 * Component that interactively displays filenames in the open repository when given fuzzy queries.
 *
 * Similar to "Go to file" in VS Code or the "t" keyboard shortcut on github.com
 */
export const FuzzyModal: React.FunctionComponent<FuzzyModalProps> = props => {
    // NOTE: the query is cached in local storage to mimic the file pickers in
    // IntelliJ (by default) and VS Code (when "Workbench > Quick Open >
    // Preserve Input" is enabled).
    const [query, setQuery] = useLocalStorage(`fuzzy-modal.query.${props.repoName}`, props.initialQuery)

    // The "focus index" is the index of the file result that the user has
    // select with up/down arrow keys. The focused item is highlighted and the
    // window.location is moved to that URL when the user presses the enter key.
    const [focusIndex, setFocusIndex] = useState(0)

    // The maximum number of results to display in the fuzzy finder. For large
    // repositories, a generic query like "src" may return thousands of results
    // making DOM rendering slow.  The user can increase this number by clicking
    // on a button at the bottom of the result list.
    const [maxResults, setMaxResults] = useState(props.initialMaxResults)

    const state: FuzzyModalState = {
        query,
        setQuery,
        focusIndex,
        setFocusIndex,
        maxResults,
        increaseMaxResults: () => {
            setMaxResults(maxResults + props.initialMaxResults)
        },
    }

    const fuzzyResult = renderFuzzyResult(props, state)

    // Sets the new "focus index" so that it's rounded by the number of
    // displayed filenames.  Cycles so that the user can press-hold the down
    // arrow and it goes all the way down and back up to the top result.
    function setRoundedFocusIndex(increment: number): void {
        const newNumber = state.focusIndex + increment
        const index = newNumber % fuzzyResult.resultsCount
        const nextIndex = index < 0 ? fuzzyResult.resultsCount + index : index
        state.setFocusIndex(nextIndex)
        document.querySelector(`#fuzzy-modal-result-${nextIndex}`)?.scrollIntoView(false)
    }

    function onInputKeyDown(event: React.KeyboardEvent): void {
        switch (event.key) {
            case 'Escape':
                props.onClose()
                break
            case 'ArrowDown':
                event.preventDefault() // Don't move the cursor to the end of the input.
                setRoundedFocusIndex(1)
                break
            case 'PageDown':
                setRoundedFocusIndex(PAGE_DOWN_INCREMENT)
                break
            case 'ArrowUp':
                event.preventDefault() // Don't move the cursor to the start of input.
                setRoundedFocusIndex(-1)
                break
            case 'PageUp':
                setRoundedFocusIndex(-PAGE_DOWN_INCREMENT)
                break
            case 'Enter':
                if (state.focusIndex < fuzzyResult.resultsCount) {
                    const fileAnchor = document.querySelector<HTMLAnchorElement>(
                        `#fuzzy-modal-result-${state.focusIndex} a`
                    )
                    fileAnchor?.click()
                    props.onClose()
                }
                break
            default:
        }
    }

    const renderedLinks = (
        <ul id={FUZZY_MODAL_RESULTS} className={styles.results} role="listbox" aria-label="Fuzzy finder results">
            {fuzzyResult.linksToRender.map((file, fileIndex) => (
                <li
                    id={fuzzyResultId(fileIndex)}
                    key={file.text}
                    role="option"
                    aria-selected={fileIndex === state.focusIndex}
                    className={classNames('p-1', fileIndex === state.focusIndex && styles.focused)}
                >
                    <HighlightedLink {...file} />
                </li>
            ))}
        </ul>
    )

    return (
        <Dialog
            className={classNames(styles.modal, 'modal-body p-4 rounded border')}
            onDismiss={() => props.onClose()}
            aria-labelledby={FUZZY_MODAL_TITLE}
        >
            <div className={styles.content}>
                <div className={styles.header}>
                    <h3 className="mb-0" id={FUZZY_MODAL_TITLE}>
                        Find file
                    </h3>
                    <button type="button" className="btn btn-icon" onClick={() => props.onClose()} aria-label="Close">
                        <CloseIcon className={classNames('icon-inline', styles.closeIcon)} />
                    </button>
                </div>
                <input
                    autoComplete="off"
                    spellCheck="false"
                    role="combobox"
                    aria-autocomplete="list"
                    aria-controls={FUZZY_MODAL_RESULTS}
                    aria-owns={FUZZY_MODAL_RESULTS}
                    aria-expanded={props.fsm.key !== 'downloading'}
                    aria-activedescendant={fuzzyResultId(state.focusIndex)}
                    id="fuzzy-modal-input"
                    className={classNames('form-control py-1', styles.input)}
                    placeholder="Enter a partial file path or name"
                    value={state.query}
                    onChange={event => {
                        state.setQuery(event.target.value)
                        state.setFocusIndex(0)
                    }}
                    type="text"
                    onKeyDown={onInputKeyDown}
                />
                <div className={styles.summary}>
                    <FuzzyResultsSummary fsm={props.fsm} files={fuzzyResult} />
                </div>
                {fuzzyResult.element}
                {renderedLinks}
                {!fuzzyResult.isComplete && (
                    <button
                        className={classNames('btn btn-secondary', styles.showMore)}
                        type="button"
                        onClick={() => state.increaseMaxResults()}
                    >
                        Show more
                    </button>
                )}
            </div>
        </Dialog>
    )
}

function fuzzyResultId(id: number): string {
    return `fuzzy-modal-result-${id}`
}
