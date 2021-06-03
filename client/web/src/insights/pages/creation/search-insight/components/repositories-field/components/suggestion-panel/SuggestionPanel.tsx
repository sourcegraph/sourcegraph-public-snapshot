import { ComboboxList, ComboboxOption } from '@reach/combobox';
import React, { ReactElement } from 'react';

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner';
import { isErrorLike } from '@sourcegraph/shared/src/util/errors';

import { ErrorMessage } from '../../../../../../../../components/alerts';

import styles from './SuggestionPanel.module.scss';

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
        return (
            <ErrorMessage error={suggestions} />
        )
    }

    return (
        <ComboboxList>
            {suggestions.map(suggestion =>
                <ComboboxOption
                    key={suggestion.name}
                    value={suggestion.name} />
            )}

            {
                !suggestions.length &&
                // eslint-disable-next-line react/forbid-dom-props
                <span style={{ display: 'block', margin: 8 }}>
                        No results found
                    </span>
            }
        </ComboboxList>
    )
}
