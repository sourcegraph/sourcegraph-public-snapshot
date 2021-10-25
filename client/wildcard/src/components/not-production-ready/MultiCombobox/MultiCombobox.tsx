import { Combobox, ComboboxInput, ComboboxList, ComboboxOption, ComboboxPopover } from '@reach/combobox'
import classnames from 'classnames'
import { noop } from 'lodash'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { ChangeEvent, ReactElement, useMemo, useState } from 'react'

import { Button } from '@sourcegraph/wildcard'

import styles from './MultiCombobox.module.scss'

interface ComboboxProps<T> {
    values: T[]
    getTokenKey: (token: T) => string
    getTokenTitle: (token: T) => string
    getSuggestions: (searchInput: string) => T[]
    onChange: (newValues: T[]) => void
    onSearchChange?: (searchInput: string) => void
}

/**
 * THIS COMPONENT IS NOT READY FOR USE.
 */
export function MultiCombobox<T>(props: ComboboxProps<T>): ReactElement {
    const { values, getTokenKey, getTokenTitle, getSuggestions, onSearchChange = noop, onChange } = props

    const [searchValue, setSearchValue] = useState('')
    const suggestions = useMemo(() => getSuggestions(searchValue), [getSuggestions, searchValue])

    const handleSearchInputChange = (event: ChangeEvent<HTMLInputElement>): void => {
        setSearchValue(event.target.value)
        onSearchChange(event.target.value)
    }

    const handleOnSelect = (value: string): void => {
        const suggestion = suggestions.find(suggestion => getTokenTitle(suggestion) === value)

        if (suggestion) {
            onChange([...values, suggestion])
            setSearchValue('')
        }
    }

    const handleTokenRemoveClick = (token: T): void => {
        onChange(values.filter(valueToken => getTokenKey(token) !== getTokenKey(valueToken)))
    }

    return (
        <div className={styles.multicomboboxContainer}>
            <div className={styles.badges}>
                {values.map(token => (
                    <span key={getTokenKey(token)} className={classnames('badge badge-secondary', styles.badge)}>
                        {getTokenTitle(token)}

                        <Button
                            className={classnames('btn-icon', styles.badgeButton)}
                            outline={true}
                            onClick={() => handleTokenRemoveClick(token)}
                        >
                            <CloseIcon />
                        </Button>
                    </span>
                ))}
            </div>

            <Combobox className={styles.multicombobox} onSelect={handleOnSelect}>
                <ComboboxInput
                    placeholder="Type value for searching"
                    className={classnames(styles.input)}
                    value={searchValue}
                    onChange={handleSearchInputChange}
                />

                <ComboboxPopover className={styles.popover} portal={true}>
                    <ComboboxList className={styles.suggestionsList}>
                        {suggestions.map(suggestion => (
                            <ComboboxOption
                                key={getTokenKey(suggestion)}
                                value={getTokenTitle(suggestion)}
                                className={styles.suggestionsListItem}
                            />
                        ))}
                    </ComboboxList>
                </ComboboxPopover>
            </Combobox>
        </div>
    )
}
