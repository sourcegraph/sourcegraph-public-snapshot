import React, { useCallback, useRef } from 'react'

import classNames from 'classnames'
import { useMergeRefs } from 'use-callback-ref'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { useAutoFocus, Input } from '@sourcegraph/wildcard'

import { FilterControl, FilteredConnectionFilter, FilteredConnectionFilterValue } from '../FilterControl'
import { OrderControl, OrderedConnectionOrderingOption, OrderedConnectionOrderValue } from '../OrderControl'

import styles from './ConnectionForm.module.scss'

export interface ConnectionFormProps {
    /** Hides the filter input field. */
    hideSearch?: boolean

    /** CSS class name for the <input> element */
    inputClassName?: string

    /** CSS class name for the <form> element */
    formClassName?: string

    /** Placeholder text for the <input> element */
    inputPlaceholder?: string

    /** Value of the <input> element */
    inputValue?: string

    /** aria-label for the <input> element */
    inputAriaLabel?: string

    /** Called when the <input> element value changes */
    onInputChange?: React.ChangeEventHandler<HTMLInputElement>

    /** Autofocuses the filter input field. */
    autoFocus?: boolean

    /**
     * Ordering options to display next to the order input field.
     *
     * Ordering options are mutually exclusive.
     */
    orderingOptions?: OrderedConnectionOrderingOption[]

    onOrderingOptionSelect?: (
        orderingOption: OrderedConnectionOrderingOption,
        value: OrderedConnectionOrderValue
    ) => void

    orderValues?: Map<string, OrderedConnectionOrderValue>

    /**
     * Filters to display next to the filter input field.
     *
     * Filters are mutually exclusive.
     */
    filters?: FilteredConnectionFilter[]

    onFilterSelect?: (filter: FilteredConnectionFilter, value: FilteredConnectionFilterValue) => void

    /** An element rendered as a sibling of the filters. */
    additionalFilterElement?: React.ReactElement

    filterValues?: Map<string, FilteredConnectionFilterValue>

    compact?: boolean
}

/**
 * FilteredConnection form input.
 * Supports <input> for querying and <select>/<radio> controls for filtering
 */
export const ConnectionForm = React.forwardRef<HTMLInputElement, ConnectionFormProps>(
    (
        {
            hideSearch,
            formClassName,
            inputClassName,
            inputPlaceholder,
            inputAriaLabel,
            inputValue,
            onInputChange,
            autoFocus,
            orderingOptions,
            onOrderingOptionSelect,
            orderValues,
            filters,
            onFilterSelect,
            filterValues,
            additionalFilterElement,
            compact,
        },
        reference
    ) => {
        const localReference = useRef<HTMLInputElement>(null)
        const mergedReference = useMergeRefs([localReference, reference])
        const handleSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(event => {
            // Do nothing. The <input onChange> handler will pick up any changes shortly.
            event.preventDefault()
        }, [])

        useAutoFocus({ autoFocus, reference: localReference })

        return (
            <Form
                className={classNames(styles.form, !compact && styles.noncompact, formClassName)}
                onSubmit={handleSubmit}
            >
                {orderingOptions && onOrderingOptionSelect && orderValues && (
                    <OrderControl
                        orderingOptions={orderingOptions}
                        onValueSelect={onOrderingOptionSelect}
                        values={orderValues}
                    >
                        {additionalFilterElement}
                    </OrderControl>
                )}
                {filters && onFilterSelect && filterValues && (
                    <FilterControl filters={filters} onValueSelect={onFilterSelect} values={filterValues}>
                        {additionalFilterElement}
                    </FilterControl>
                )}
                {!hideSearch && (
                    <Input
                        className={classNames(styles.input, inputClassName)}
                        type="search"
                        placeholder={inputPlaceholder}
                        name="query"
                        value={inputValue}
                        onChange={onInputChange}
                        autoFocus={autoFocus}
                        autoComplete="off"
                        autoCorrect="off"
                        autoCapitalize="off"
                        ref={mergedReference}
                        spellCheck={false}
                        aria-label={inputAriaLabel}
                        variant={compact ? 'small' : 'regular'}
                    />
                )}
            </Form>
        )
    }
)
ConnectionForm.displayName = 'ConnectionForm'
