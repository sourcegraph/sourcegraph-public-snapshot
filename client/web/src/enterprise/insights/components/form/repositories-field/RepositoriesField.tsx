import { ChangeEvent, FocusEvent, forwardRef, MouseEvent, Ref, useImperativeHandle, useRef, useState } from 'react'

import { Combobox, ComboboxInput, ComboboxPopover } from '@reach/combobox'

import { FlexTextArea, ForwardReferenceComponent } from '@sourcegraph/wildcard'

import { SuggestionsPanel } from './components/suggestion-panel/SuggestionPanel'
import { useRepoSuggestions } from './hooks/use-repo-suggestions'
import { RepositoryFieldProps } from './types'
import { getSuggestionsSearchTerm } from './utils/get-suggestions-search-term'

import styles from './RepositoriesField.module.scss'

/**
 * Renders multi repositories input with suggestions.
 */
const RepositoriesField = forwardRef((props: RepositoryFieldProps, reference: Ref<HTMLInputElement | null>) => {
    const { value, onChange, onBlur, ...otherProps } = props

    const inputReference = useRef<HTMLInputElement>(null)

    const [caretPosition, setCaretPosition] = useState<number | null>(null)
    const [panel, setPanel] = useState(false)

    const { repositories, value: search, index: searchTermIndex } = getSuggestionsSearchTerm({
        value,
        caretPosition,
    })
    const { searchValue, suggestions } = useRepoSuggestions({
        search,
        disable: !panel,
    })

    // Support top level reference prop
    useImperativeHandle(reference, () => inputReference.current)

    const handleInputChange = (event: ChangeEvent<HTMLInputElement>): void => {
        onChange(event.target.value)
        setCaretPosition(event.target.selectionStart)
        setPanel(true)
    }

    const handleSelect = (selectValue: string): void => {
        const separatorString = ', '

        if (searchTermIndex !== null) {
            const newRepositoriesSerializedValue =
                [
                    ...repositories.slice(0, searchTermIndex),
                    selectValue,
                    ...repositories.slice(searchTermIndex + 1),
                ].join(separatorString) + separatorString

            onChange(newRepositoriesSerializedValue)
            setPanel(false)

            /**
             * Setting the value ('onChange' above) triggers the reset selection of the input
             * if the user selects a value from suggestion panel for some sub-string of
             * the input value we need to preserve the selection at the end of the sub-string
             * and avoid resetting and putting the selection at the end of input string.
             */
            setTimeout(() => {
                if (!inputReference.current) {
                    return
                }

                const isLastItemEdited = searchTermIndex === repositories.length - 1
                const endOfSelectedItem = [...repositories.slice(0, searchTermIndex), selectValue].join(separatorString)
                    .length

                const nextCaretPosition = isLastItemEdited
                    ? // Put cursor at the end the input value
                      newRepositoriesSerializedValue.length
                    : endOfSelectedItem

                inputReference.current.setSelectionRange(nextCaretPosition, nextCaretPosition)
            }, 0)
        }
    }

    const trackInputCursorChange = (event: MouseEvent | KeyboardEvent | FocusEvent): void => {
        const target = event.target as HTMLInputElement

        if (caretPosition !== target.selectionStart) {
            /**
             * After the moment when user selected the value from the suggestion panel we closed
             * this panel by setPanel(false) but if the user is changing the cursor position we
             * need to re-open suggestion panel for the new suggestions.
             */
            setPanel(true)
            setCaretPosition(target.selectionStart)
        }
    }

    const handleInputFocus = (event: FocusEvent): void => {
        setPanel(true)
        trackInputCursorChange(event)
    }

    const handleInputBlur = (event: FocusEvent<HTMLInputElement>): void => {
        onBlur?.(event)
    }

    return (
        <Combobox openOnFocus={true} onSelect={handleSelect} className={styles.combobox}>
            <ComboboxInput
                {...otherProps}
                as={FlexTextArea}
                ref={inputReference}
                autocomplete={false}
                value={value}
                onChange={handleInputChange}
                onFocus={handleInputFocus}
                onBlur={handleInputBlur}
                onClick={trackInputCursorChange}
            />

            {panel && (
                <ComboboxPopover hidden={false} className={styles.comboboxReachPopover}>
                    <SuggestionsPanel value={searchValue} suggestions={suggestions} className={styles.popover} />
                </ComboboxPopover>
            )}
        </Combobox>
    )
}) as ForwardReferenceComponent<'input', RepositoryFieldProps>

RepositoriesField.displayName = 'RepositoriesField'

export { RepositoriesField }
