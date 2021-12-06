import classNames from 'classnames'
import React, { useCallback, useEffect, useState } from 'react'
import { Form } from 'reactstrap'

import { useDebounce } from '@sourcegraph/wildcard'

import { CatalogEntityFiltersProps } from '../../../../core/entity-filters'

import styles from './CatalogExplorerViewOptionsRow.module.scss'

interface Props extends CatalogEntityFiltersProps {
    before?: React.ReactFragment
    toggle: React.ReactFragment
    className?: string
}

export const CatalogExplorerViewOptionsRow: React.FunctionComponent<Props> = ({
    before,
    toggle,
    filters,
    onFiltersChange,
    className,
}) => {
    const [query, setQuery] = useState(filters.query)

    const debouncedQuery = useDebounce(query, 200)
    useEffect(() => {
        if (filters.query !== debouncedQuery) {
            onFiltersChange({ ...filters, query: debouncedQuery })
        }
    }, [filters, onFiltersChange, debouncedQuery])

    const onQueryChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => setQuery(event.currentTarget.value),
        []
    )

    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        event => {
            event.preventDefault()
            onFiltersChange({ ...filters, query })
        },
        [filters, onFiltersChange, query]
    )

    return (
        <Form className={classNames(styles.form, className)} onSubmit={onSubmit}>
            {before && <div>{before}</div>}
            {toggle}
            <div className={classNames('form-group mb-0')}>
                <label htmlFor="entity-list-filters__query" className="sr-only">
                    Query
                </label>
                <input
                    id="entity-list-filters__query"
                    className={classNames('form-control')}
                    type="search"
                    onChange={onQueryChange}
                    placeholder="Search..."
                    value={query}
                />
            </div>
            <button type="submit" className="sr-only">
                Filter
            </button>
        </Form>
    )
}
