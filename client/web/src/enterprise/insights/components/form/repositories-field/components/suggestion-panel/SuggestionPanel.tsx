import React from 'react'

import { ComboboxList, ComboboxOption, ComboboxOptionText } from '@reach/combobox'
import classNames from 'classnames'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isErrorLike } from '@sourcegraph/common'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import styles from './SuggestionPanel.module.scss'

interface SuggestionsPanelProps {
    value: string | null
    suggestions?: Error | RepositorySuggestion[]
    className?: string
}

interface RepositorySuggestion {
    id: string
    name: string
}

/**
 * Renders suggestion panel for repositories combobox component.
 */
export const SuggestionsPanel: React.FunctionComponent<React.PropsWithChildren<SuggestionsPanelProps>> = props => {
    const { value, suggestions, className } = props

    if (suggestions === undefined) {
        return (
            <div className={classNames(styles.loadingPanel, classNames)}>
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
                <ComboboxOption className={styles.suggestionsListItem} key={suggestion.id} value={suggestion.name}>
                    <SourceRepositoryIcon className="mr-1" size="1rem" />
                    <ComboboxOptionText />
                </ComboboxOption>
            ))}

            {!isValueEmpty && !suggestions.length && (
                <span className={styles.suggestionsListItem}>No results found</span>
            )}
        </ComboboxList>
    )
}
