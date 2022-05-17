import React, { useCallback } from 'react'

import classNames from 'classnames'

import { Select } from '@sourcegraph/wildcard'

import { RadioButtons } from '../RadioButtons'

import styles from './FilterControl.module.scss'

export interface FilteredConnectionFilterValue {
    value: string
    label: string
    tooltip?: string
    args: { [name: string]: string | number | boolean }
}

/**
 * A filter to display next to the filter input field.
 */
export interface FilteredConnectionFilter {
    /** The UI label for the filter. */
    label: string

    type: string

    /**
     * The URL string for this filter (conventionally the label, lowercased and without spaces and punctuation).
     */
    id: string

    /** An optional tooltip to display for this filter. */
    tooltip?: string

    values: FilteredConnectionFilterValue[]
}

interface FilterControlProps {
    /** All filters. */
    filters: FilteredConnectionFilter[]

    /** Called when a filter is selected. */
    onValueSelect: (filter: FilteredConnectionFilter, value: FilteredConnectionFilterValue) => void

    values: Map<string, FilteredConnectionFilterValue>
}

export const FilterControl: React.FunctionComponent<React.PropsWithChildren<FilterControlProps>> = ({
    filters,
    values,
    onValueSelect,
    children,
}) => {
    const onChange = useCallback(
        (filter: FilteredConnectionFilter, id: string) => {
            const value = filter.values.find(value => value.value === id)
            if (value === undefined) {
                return
            }
            onValueSelect(filter, value)
        },
        [onValueSelect]
    )

    return (
        <div className={styles.filterControl}>
            {filters.map(filter => {
                if (filter.type === 'radio') {
                    return (
                        <RadioButtons
                            key={filter.id}
                            name={filter.id}
                            className="d-inline-flex flex-row"
                            selected={values.get(filter.id)?.value}
                            nodes={filter.values.map(({ value, label, tooltip }) => ({
                                tooltip,
                                label,
                                id: value,
                            }))}
                            onChange={event => onChange(filter, event.currentTarget.value)}
                        />
                    )
                }

                if (filter.type === 'select') {
                    const filterLabelId = `filtered-select-label-${filter.id}`
                    return (
                        <div
                            key={filter.id}
                            className={classNames('d-inline-flex flex-row align-center flex-wrap', styles.select)}
                        >
                            <div className="d-inline-flex flex-row mr-3 align-items-baseline">
                                <p className="text-xl-center text-nowrap mr-2" id={filterLabelId}>
                                    {filter.label}:
                                </p>
                                <Select
                                    aria-labelledby={filterLabelId}
                                    id=""
                                    name={filter.id}
                                    onChange={event => onChange(filter, event.currentTarget.value)}
                                    className="mb-0"
                                >
                                    {filter.values.map(value => (
                                        <option key={value.value} value={value.value} label={value.label} />
                                    ))}
                                </Select>
                            </div>
                        </div>
                    )
                }

                return null
            })}
            {children}
        </div>
    )
}
