import { Combobox, ComboboxInput, ComboboxPopover } from '@reach/combobox'
import React, { ChangeEvent, forwardRef, Ref, useImperativeHandle, useRef } from 'react'

import { getSanitizedRepositories } from '../../../pages/creation/search-insight/utils/insight-sanitizer'

import { FlexTextArea } from './components/flex-textarea/FlexTextArea'
import { SuggestionsPanel } from './components/suggestion-panel/SuggestionPanel'
import { useRepoSuggestions } from './hooks/use-repo-suggestions'
import styles from './RepositoriesField.module.scss'
import { RepositoryFieldProps } from './types'

/**
 * Renders single repository field with suggestions.
 */
export const RepositoryField = forwardRef((props: RepositoryFieldProps, reference: Ref<HTMLInputElement | null>) => {
    const { value, onChange, onBlur, ...otherProps } = props

    const inputReference = useRef<HTMLInputElement>(null)

    const { searchValue, suggestions } = useRepoSuggestions({
        excludedItems: [value],
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

            <ComboboxPopover className={styles.comboboxPopover}>
                <SuggestionsPanel value={searchValue} suggestions={suggestions} />
            </ComboboxPopover>
        </Combobox>
    )
})
