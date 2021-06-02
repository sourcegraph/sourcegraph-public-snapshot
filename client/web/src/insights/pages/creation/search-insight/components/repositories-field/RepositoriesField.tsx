import {
    Combobox,
    ComboboxInput,
    ComboboxPopover,
    ComboboxList,
    ComboboxOption,
} from '@reach/combobox';
import classnames from 'classnames';
import React, { ChangeEvent, ReactElement, useEffect, useState } from 'react';

import { useDebounce } from '@sourcegraph/wildcard/src';

import { fetchRepositorySuggestions } from '../../../../../core/backend/requests/fetch-repository-suggestions';

import './RepositoriesField.module.scss'

interface RepositoriesFieldProps {

}

interface RepositorySuggestion {
    name: string
}

export function RepositoriesField(props: RepositoriesFieldProps): ReactElement | null {
    const {} = props
    const [value, setValue] = useState<string>('')
    const [suggestions, setSuggestions] = useState<RepositorySuggestion[]>([])

    const debouncedValue = useDebounce(value, 500);

    useEffect(() => {
        if (debouncedValue.trim() !== '') {

            let isOutdatedRequest = false;

            fetchRepositorySuggestions(debouncedValue).toPromise()
                .then(suggestions => !isOutdatedRequest && setSuggestions(suggestions))
                .catch(error => {
                    console.error(error)
                })

            return () => {
                isOutdatedRequest = true
            }
        }

        return
    }, [debouncedValue])

    const handleInputChange = (event: ChangeEvent<HTMLInputElement>): void => {
        console.log('INPUT change', event.target.value)

        setValue(event.target.value)
    }

    const handleSelect = (selectValue: string): void => {
        console.log('Select', selectValue)

        setValue(selectValue)
    }

    return (
        <Combobox onSelect={handleSelect} aria-label="choose a fruit">
            <ComboboxInput
                as="input"
                autocomplete={false}
                value={value}
                onChange={handleInputChange}
                className={classnames('form-control')}
            />
            <ComboboxPopover>
                {
                    suggestions.length > 0
                        ? <ComboboxList>
                            {
                                suggestions.map(suggestion =>
                                    <ComboboxOption key={suggestion.name} value={suggestion.name} />
                                )
                            }
                        </ComboboxList>
                        : (
                            // eslint-disable-next-line react/forbid-dom-props
                            <span style={{ display: 'block', margin: 8 }}>
                                No results found
                            </span>
                        )
                }
            </ComboboxPopover>
        </Combobox>
    )
}
