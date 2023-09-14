import { type ChangeEvent, forwardRef, type InputHTMLAttributes } from 'react'

import { Combobox, ComboboxInput, ComboboxPopover } from '@reach/combobox'

import { FlexTextArea, useDebounce } from '@sourcegraph/wildcard'

import { SuggestionsPanel } from './components/suggestion-panel/SuggestionPanel'
import { useRepoSuggestions } from './hooks/use-repo-suggestions'

import styles from './RepositoriesField.module.scss'

interface RepositoryFieldProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'value' | 'onChange'> {
    value: string

    /**
     * Change handler runs when user changed input value or picked element
     * from the suggestion panel.
     */
    onChange: (value: string) => void
}

/**
 * Renders single repository field with suggestions.
 */
export const RepositoryField = forwardRef<HTMLInputElement, RepositoryFieldProps>((props, reference) => {
    const { value, onChange, onBlur, ...otherProps } = props

    const debouncedSearchTerm = useDebounce(value, 500)
    const suggestions = useRepoSuggestions({
        search: debouncedSearchTerm,
        selectedRepositories: [],
    })

    const handleInputChange = (event: ChangeEvent<HTMLInputElement>): void => {
        onChange(event.target.value)
    }

    return (
        <Combobox openOnFocus={true} onSelect={onChange} className={styles.combobox}>
            <ComboboxInput
                {...otherProps}
                as={FlexTextArea}
                ref={reference}
                value={value}
                onChange={handleInputChange}
            />

            <ComboboxPopover className={styles.comboboxReachPopover}>
                <SuggestionsPanel
                    value={debouncedSearchTerm}
                    suggestions={suggestions.suggestions}
                    className={styles.popover}
                />
            </ComboboxPopover>
        </Combobox>
    )
})
