import React, { useCallback } from 'react'

import { Select, Text } from '@sourcegraph/wildcard'

import { RadioButtons } from '../RadioButtons'

import styles from './OrderControl.module.scss'

export interface OrderedConnectionOrderValue {
    value: string
    label: string
    tooltip?: string
    args: { [name: string]: string | number | boolean }
}

/**
 * An order selector to display next to the search input field.
 */
export interface OrderedConnectionOrderingOption {
    /** The UI label for the ordering option. */
    label: string

    /** "radio" or "select" */
    type: string

    /**
     * The URL string for this ordering option (conventionally the label, lowercased and without spaces and punctuation).
     */
    id: string

    /** An optional tooltip to display for this order option. */
    tooltip?: string

    values: OrderedConnectionOrderValue[]
}

interface OrderControlProps {
    /** All ordering options. */
    orderingOptions: OrderedConnectionOrderingOption[]

    /** Called when an order is selected. */
    onValueSelect: (orderingOption: OrderedConnectionOrderingOption, value: OrderedConnectionOrderValue) => void

    values: Map<string, OrderedConnectionOrderValue>
}

export const OrderControl: React.FunctionComponent<React.PropsWithChildren<OrderControlProps>> = ({
    orderingOptions,
    values,
    onValueSelect,
    children,
}) => {
    const onChange = useCallback(
        (order: OrderedConnectionOrderingOption, id: string) => {
            const value = order.values.find(value => value.value === id)
            if (value === undefined) {
                return
            }
            onValueSelect(order, value)
        },
        [onValueSelect]
    )

    return (
        <div className={styles.orderControl}>
            {orderingOptions.map(orderingOption => {
                if (orderingOption.type === 'radio') {
                    return (
                        <RadioButtons
                            key={orderingOption.id}
                            name={orderingOption.id}
                            className="d-inline-flex flex-row"
                            selected={values.get(orderingOption.id)?.value}
                            nodes={orderingOption.values.map(({ value, label, tooltip }) => ({
                                tooltip,
                                label,
                                id: value,
                            }))}
                            onChange={event => onChange(orderingOption, event.currentTarget.value)}
                        />
                    )
                }

                if (orderingOption.type === 'select') {
                    const orderLabelId = `ordered-select-label-${orderingOption.id}`
                    return (
                        <div key={orderingOption.id} className="d-inline-flex flex-row align-center flex-wrap">
                            <div className="d-inline-flex flex-row mr-3 align-items-baseline">
                                <Text className="text-xl-center text-nowrap mr-2" id={orderLabelId}>
                                    {orderingOption.label}:
                                </Text>
                                <Select
                                    aria-labelledby={orderLabelId}
                                    id=""
                                    name={orderingOption.id}
                                    onChange={event => onChange(orderingOption, event.currentTarget.value)}
                                    value={values.get(orderingOption.id)?.value}
                                    className="mb-0"
                                >
                                    {orderingOption.values.map(value => (
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
