import { ComboboxList, ComboboxOption, ComboboxOptionText } from '@reach/combobox'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import React, { ReactElement } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../../../../../../components/alerts'

import styles from './SuggestionPanel.module.scss'

interface SuggestionsPanelProps {
    value: string | null
    suggestions?: Error | RepositorySuggestion[]
}

interface RepositorySuggestion {
    id: string
    name: string
}

/**
 * Renders suggestion panel for repositories combobox component.
 */
export function SuggestionsPanel(props: SuggestionsPanelProps): ReactElement {
    const { value, suggestions } = props

    if (suggestions === undefined) {
        return (
            <div className={styles.loadingPanel}>
                <LoadingSpinner />
            </div>
        )
    }

    if (isErrorLike(suggestions)) {
        return <ErrorAlert className="m-1" error={suggestions} data-testid="repository-suggestions-error" />
    }

    const searchValue = value ?? ''
    const isValueEmpty = searchValue.trim() === ''

    return (
        <ComboboxList className={styles.suggestionsList}>
            {suggestions.map(suggestion => (
                <ComboboxOption className={styles.suggestionsListItem} key={suggestion.id} value={suggestion.name}>
                    <SourceBranchIcon className="mr-1" size="1rem" />
                    <ComboboxOptionText />
                </ComboboxOption>
            ))}

            {isValueEmpty && <span className={styles.suggestionsListItem}>Start entering the value</span>}

            {!isValueEmpty && !suggestions.length && (
                <span className={styles.suggestionsListItem}>No results found</span>
            )}
        </ComboboxList>
    )
}
