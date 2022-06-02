import React, { useCallback, useRef } from 'react'

import classNames from 'classnames'
import { useMergeRefs } from 'use-callback-ref'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { useAutoFocus, Input } from '@sourcegraph/wildcard'

import { FilterControl, FilteredConnectionFilter, FilteredConnectionFilterValue } from '../FilterControl'

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
     * Filters to display next to the filter input field.
     *
     * Filters are mutually exclusive.
     */
    filters?: FilteredConnectionFilter[]

    onValueSelect?: (filter: FilteredConnectionFilter, value: FilteredConnectionFilterValue) => void

    /** An element rendered as a sibling of the filters. */
    additionalFilterElement?: React.ReactElement

    values?: Map<string, FilteredConnectionFilterValue>

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
            filters,
            onValueSelect,
            additionalFilterElement,
            values,
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
                className={classNames(
                    'w-100 d-inline-flex justify-content-between flex-row',
                    !compact && styles.noncompact,
                    formClassName
                )}
                onSubmit={handleSubmit}
            >
                {filters && onValueSelect && values && (
                    <FilterControl filters={filters} onValueSelect={onValueSelect} values={values}>
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
                    />
                )}
            </Form>
        )
    }
)
