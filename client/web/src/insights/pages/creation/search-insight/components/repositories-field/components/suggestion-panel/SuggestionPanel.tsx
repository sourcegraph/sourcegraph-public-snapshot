import { ComboboxList, ComboboxOption, ComboboxOptionText } from '@reach/combobox'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon';
import React, { ReactElement } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ErrorAlert } from '../../../../../../../../components/alerts'

import styles from './SuggestionPanel.module.scss'

interface SuggestionsPanelProps {
    suggestions?: Error | RepositorySuggestion[]
}

interface RepositorySuggestion {
    name: string
}

/**
 * Renders suggestion panel for repositories combobox component.
 */
export function SuggestionsPanel(props: SuggestionsPanelProps): ReactElement {
    const { suggestions } = props

    if (suggestions === undefined) {
        return (
            <div className={styles.loadingPanel}>
                <LoadingSpinner />
            </div>
        )
    }

    if (isErrorLike(suggestions)) {
        return <ErrorAlert className="m-1" error={suggestions} data-testid='repository-suggestions-error'/>
    }

    return (
        <ComboboxList className={styles.suggestionsList}>
            {suggestions.map(suggestion => (
                <ComboboxOption className={styles.suggestionsListItem} key={suggestion.name} value={suggestion.name}>
                    <SourceBranchIcon className="mr-1" size='1rem'/>
                    <ComboboxOptionText />
                </ComboboxOption>
            ))}

            {!suggestions.length && (
                // eslint-disable-next-line react/forbid-dom-props
                <span style={{ display: 'block', margin: 8 }}>No results found</span>
            )}
        </ComboboxList>
    )
}
