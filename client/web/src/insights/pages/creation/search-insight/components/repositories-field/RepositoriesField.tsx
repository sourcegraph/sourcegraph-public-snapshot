import {
    Combobox,
    ComboboxInput,
    ComboboxPopover,
} from '@reach/combobox'
import React, {
    MouseEvent,
    ChangeEvent,
    FocusEvent,
    useRef,
    useState,
    forwardRef,
    useImperativeHandle, Ref, InputHTMLAttributes
} from 'react'

import { FlexTextarea } from './components/flex-textarea/FlexTextArea'
import { SuggestionsPanel } from './components/suggestion-panel/SuggestionPanel';
import { useRepoSuggestions } from './hooks/use-repo-suggestions';
import styles from './RepositoriesField.module.scss'
import { getSuggestionsSearchTerm } from './utils/get-suggestions-search-term';

interface RepositoriesFieldProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'value' | 'onChange'> {
    value: string
    onChange: (value: string) => void
}

/**
 * Renders repository input with suggestion panel.
 */
export const RepositoriesField = forwardRef((props: RepositoriesFieldProps, reference:  Ref<HTMLInputElement | null>) => {
    const { value, onChange, onBlur, ...otherProps } = props

    const inputReference = useRef<HTMLInputElement>(null)

    const [caretPosition, setCaretPosition] = useState<number|null>(null)
    const [panel, setPanel] = useState(false)

    const { value: search } = getSuggestionsSearchTerm({ value, caretPosition })
    const { suggestions } = useRepoSuggestions({ search, disable: !panel })

    // Support top level reference prop
    useImperativeHandle(reference, () => inputReference.current)

    const handleInputChange = (event: ChangeEvent<HTMLInputElement>): void => {
        onChange(event.target.value)
        setCaretPosition(event.target.selectionStart)
        setPanel(true)
    }

    const handleSelect = (selectValue: string): void => {
        const separatorString = ', '
        const { repositories, index } = getSuggestionsSearchTerm({ value, caretPosition })

        if (index !== null) {
            const newRepositoriesSerializedValue = [
                ...repositories.slice(0, index),
                selectValue,
                ...repositories.slice(index+1),
            ].join(separatorString)

            onChange(newRepositoriesSerializedValue)
            setPanel(false)

            /**
             * Setting value (setValue above) triggers reset selection of input
             * if user select value from suggestion panel for some sub-string of
             * input value we need preserve selection at the end of sub-string and
             * avoid reset and putting selection at the end of input string.
             */
            setTimeout(() => {
                if (!inputReference.current) {
                    return
                }

                const endOfSelectedItem = [...repositories.slice(0, index), selectValue]
                    .join(separatorString)
                    .length

                inputReference.current.setSelectionRange(endOfSelectedItem,endOfSelectedItem)
            }, 0)
        }
    }

    const trackInputCursorChange = (event: MouseEvent | KeyboardEvent | FocusEvent): void => {
        const target = event.target as HTMLInputElement

        if (caretPosition !== target.selectionStart) {
            /**
             * After the moment when user selected value from suggestion panel we closed
             * this panel by setPanel(false) but if user is changing cursor position we
             * need to re-open panel for new suggestions.
             * */
            setPanel(true)
            setCaretPosition(target.selectionStart)
        }
    }

    const handleInputFocus = (event: FocusEvent): void => {
        setPanel(true)
        trackInputCursorChange(event)
    }

    const handleInputBlur = (event: FocusEvent<HTMLInputElement>): void => {
        // setPanel(false)
        onBlur?.(event)
    }

    return (
        <Combobox
            openOnFocus={true}
            onSelect={handleSelect}
            aria-label="choose a fruit"
            className={styles.combobox}>

            <ComboboxInput
                {...otherProps}
                as={FlexTextarea}
                ref={inputReference}
                autocomplete={false}
                value={value}
                onChange={handleInputChange}
                onFocus={handleInputFocus}
                onBlur={handleInputBlur}
                onClick={trackInputCursorChange}
            />

            {
                panel &&
                <ComboboxPopover>
                    <SuggestionsPanel suggestions={suggestions}/>
                </ComboboxPopover>
            }
        </Combobox>
    )
})
