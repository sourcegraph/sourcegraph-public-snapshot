import React from 'react'

import { mdiSourceRepository } from '@mdi/js'
import { ComboboxList, ComboboxOption, ComboboxOptionText } from '@reach/combobox'
import classNames from 'classnames'

import { isErrorLike } from '@sourcegraph/common'
import { LoadingSpinner, Icon, ErrorAlert } from '@sourcegraph/wildcard'

import styles from './SuggestionPanel.module.scss'

interface SuggestionsPanelProps {
    value: string | null
    suggestions: string[]
    className?: string
}

/**
 * Renders suggestion panel for repositories combobox component.
 */
export const SuggestionsPanel: React.FunctionComponent<React.PropsWithChildren<SuggestionsPanelProps>> = props => {
    const { value, suggestions, className } = props

    if (suggestions === undefined) {
        return (
            <div className={classNames(styles.loadingPanel, className)}>
                <LoadingSpinner inline={false} />
            </div>
        )
    }

    if (isErrorLike(suggestions)) {
        return <ErrorAlert className="m-1" error={suggestions} data-testid="repository-suggestions-error" />
    }

    if (suggestions.length === 0) {
        return null
    }

    const searchValue = value ?? ''
    const isValueEmpty = searchValue.trim() === ''

    return (
        <ComboboxList className={classNames(styles.suggestionsList, className)}>
            {suggestions.map(suggestion => (
                <ComboboxOption className={styles.suggestionsListItem} key={suggestion} value={suggestion}>
                    <Icon
                        className="mr-1"
                        svgPath={mdiSourceRepository}
                        inline={false}
                        aria-hidden={true}
                        height="1rem"
                        width="1rem"
                    />
                    <ComboboxOptionText />
                </ComboboxOption>
            ))}

            {!isValueEmpty && !suggestions.length && (
                <span className={styles.suggestionsListItem}>No results found</span>
            )}
        </ComboboxList>
    )
}
