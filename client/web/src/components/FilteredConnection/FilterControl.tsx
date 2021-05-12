import React, { useCallback } from 'react'

export interface FilterValue {
    value: string
    label: string
    tooltip?: string
    args: { [name: string]: string | number | boolean }
}

/**
 * A filter to display next to the filter input field.
 */
export interface Filter {
    /** The UI label for the filter. */
    label: string

    type: string

    /**
     * The URL string for this filter (conventionally the label, lowercased and without spaces and punctuation).
     */
    id: string

    /** An optional tooltip to display for this filter. */
    tooltip?: string

    values: FilterValue[]
}

interface FilterControlProps {
    /** All filters. */
    filters: Filter[]

    /** Called when a filter is selected. */
    onDidSelectValue: (filter: Filter, value: FilterValue) => void

    values: Map<string, FilterValue>
}

export const FilterControl: React.FunctionComponent<FilterControlProps> = ({
    filters,
    values,
    onDidSelectValue,
    children,
}) => {
    const onChange = useCallback(
        (filter: Filter, id: string) => {
            const value = filter.values.find(value => value.value === id)
            if (value === undefined) {
                return
            }
            onDidSelectValue(filter, value)
        },
        [onDidSelectValue]
    )

    return (
        <div className="filtered-connection-filter-control">
            {filters.map(filter => (
                <div className="d-inline-flex flex-row radio-buttons" key={filter.id}>
                    {filter.type === 'radio' &&
                        filter.values.map(value => (
                            <label key={value.value} className="radio-buttons__item" title={value.tooltip}>
                                <input
                                    className="radio-buttons__input"
                                    name={value.value}
                                    type="radio"
                                    onChange={event => onChange(filter, event.currentTarget.value)}
                                    value={value.value}
                                    checked={values.get(filter.id) && values.get(filter.id)!.value === value.value}
                                />{' '}
                                <small>
                                    <div className="radio-buttons__label">{value.label}</div>
                                </small>
                            </label>
                        ))}
                    {filter.type === 'select' && (
                        <div className="d-inline-flex flex-row mr-3 align-items-baseline">
                            <p className="text-xl-center text-nowrap mr-2">{filter.label}:</p>
                            <select
                                className="form-control"
                                name={filter.id}
                                onChange={event => onChange(filter, event.currentTarget.value)}
                            >
                                {filter.values.map(value => (
                                    <option key={value.value} value={value.value} label={value.label} />
                                ))}
                            </select>
                        </div>
                    )}
                </div>
            ))}
            {children}
        </div>
    )
}
