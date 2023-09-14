import { type FC, useCallback, useId, useState } from 'react'

import classNames from 'classnames'
import { upperFirst } from 'lodash'

import {
    H3,
    H4,
    Label,
    MultiCombobox,
    MultiComboboxInput,
    MultiComboboxPopover,
    MultiComboboxList,
    MultiComboboxEmptyList,
    MultiComboboxOption,
} from '@sourcegraph/wildcard'

import type { BatchChangeState } from '../../../graphql-operations'

import styles from './BatchChangeListFilter.module.scss'

/** Returns string with capitalized first letter */
const format = (filter: BatchChangeState): string => upperFirst(filter.toLowerCase())

interface BatchChangeListFiltersProps {
    filters: BatchChangeState[]
    selectedFilters: BatchChangeState[]
    onFiltersChange: (filters: BatchChangeState[]) => void
    className?: string
}

export const BatchChangeListFilters: FC<BatchChangeListFiltersProps> = props => {
    const { filters, selectedFilters, onFiltersChange, className } = props

    const id = useId()
    const [searchTerm, setSearchTerm] = useState('')

    const handleFilterChange = useCallback(
        (newFilters: BatchChangeState[]) => {
            if (newFilters.length > selectedFilters.length) {
                // Reset value when we add new filter
                // see https://github.com/sourcegraph/sourcegraph/pull/46450#discussion_r1070840089
                setSearchTerm('')
            }

            onFiltersChange(newFilters)
        },
        [selectedFilters, onFiltersChange]
    )

    // Render only non-selected filters and filters that match with search term value
    const suggestions = filters.filter(
        filter => !selectedFilters.includes(filter) && filter.toLowerCase().includes(searchTerm.toLowerCase())
    )

    return (
        <Label htmlFor={id} className={classNames(className, styles.root)}>
            <H4 as={H3} className="mb-0 mr-2">
                Status
            </H4>

            <MultiCombobox
                selectedItems={selectedFilters}
                getItemName={format}
                getItemKey={format}
                onSelectedItemsChange={handleFilterChange}
                aria-label="Select batch change status to filter."
            >
                <MultiComboboxInput
                    id={id}
                    value={searchTerm}
                    autoCorrect="false"
                    autoComplete="off"
                    placeholder="Select filter..."
                    onChange={event => setSearchTerm(event.target.value)}
                />

                <MultiComboboxPopover>
                    <MultiComboboxList items={suggestions}>
                        {filters =>
                            filters.map((filter, index) => (
                                <MultiComboboxOption key={filter.toString()} value={format(filter)} index={index} />
                            ))
                        }
                    </MultiComboboxList>

                    {suggestions.length === 0 && (
                        <MultiComboboxEmptyList>
                            {!searchTerm ? <>All filters are selected</> : <>No options</>}
                        </MultiComboboxEmptyList>
                    )}
                </MultiComboboxPopover>
            </MultiCombobox>
        </Label>
    )
}
