import { ChangeEvent, forwardRef, Ref, useImperativeHandle, useRef } from 'react'

import { Combobox, ComboboxInput, ComboboxPopover } from '@reach/combobox'

import { FlexTextArea } from '@sourcegraph/wildcard'

import { getSanitizedRepositories } from '../../creation-ui-kit'

import { SuggestionsPanel } from './components/suggestion-panel/SuggestionPanel'
import { useRepoSuggestions } from './hooks/use-repo-suggestions'
import { RepositoryFieldProps } from './types'

import styles from './RepositoriesField.module.scss'

/**
 * Renders single repository field with suggestions.
 */
export const RepositoryField = forwardRef((props: RepositoryFieldProps, reference: Ref<HTMLInputElement | null>) => {
    const { value, onChange, onBlur, ...otherProps } = props

    const inputReference = useRef<HTMLInputElement>(null)

    const { searchValue, suggestions } = useRepoSuggestions({
        search: getSanitizedRepositories(value)[0],
    })

    // Support top level reference prop
    useImperativeHandle(reference, () => inputReference.current)

    const handleInputChange = (event: ChangeEvent<HTMLInputElement>): void => {
        onChange(event.target.value)
    }

    return (
        <Combobox openOnFocus={true} onSelect={onChange} className={styles.combobox}>
            <ComboboxInput
                {...otherProps}
                as={FlexTextArea}
                ref={inputReference}
                value={value}
                onChange={handleInputChange}
            />

            <ComboboxPopover className={styles.comboboxReachPopover}>
                <SuggestionsPanel value={searchValue} suggestions={suggestions} className={styles.popover} />
            </ComboboxPopover>
        </Combobox>
    )
})
