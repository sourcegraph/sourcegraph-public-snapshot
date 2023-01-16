import { FC, useId, useState } from 'react'

import classNames from 'classnames'

import {
    H3,
    H4,
    Label,
    MultiCombobox,
    MultiComboboxInput,
    MultiComboboxPopover,
    MultiComboboxList,
    MultiComboboxOption,
} from '@sourcegraph/wildcard'

import { BatchChangeState } from '../../../graphql-operations'

import styles from './BatchChangeListFilter.module.scss'

/** Returns string with capitalized first letter */
const format = (filter: BatchChangeState): string => {
    const str = filter.toString()

    return str.charAt(0).toUpperCase() + str.slice(1).toLowerCase()
}

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

    // Render only non-selected filters and filters that match with search term value
    const suggestions = filters.filter(filter => !selectedFilters.includes(filter) && filter.toLowerCase().includes(searchTerm.toLowerCase()))

    return (
        <Label htmlFor={id} className={classNames(className, styles.root)}>
            <H4 as={H3} className="mb-0 mr-2">
                Status
            </H4>

            <MultiCombobox
                selectedItems={selectedFilters}
                getItemName={format}
                getItemKey={format}
                onSelectedItemsChange={onFiltersChange}
                aria-label="Select batch change status to filter."
            >
                <MultiComboboxInput
                    id={id}
                    value={searchTerm}
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
                        <span className={styles.noFilters}>
                            {!searchTerm ? (
                                <>All filters are selected</>
                            ) : (
                                <>No options</>
                            )}
                        </span>
                    )}
                </MultiComboboxPopover>
            </MultiCombobox>
        </Label>
    )
}
