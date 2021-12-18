import classNames from 'classnames'
import React, { useCallback, useState } from 'react'
import { Form } from 'reactstrap'

import { ComponentKind } from '../../../../../../graphql-operations'
import { ComponentFiltersProps } from '../../../../core/entity-filters'

import { CatalogExplorerViewOptionInput } from './CatalogExplorerViewOptionInput'
import styles from './CatalogExplorerViewOptionsRow.module.scss'

interface Props extends ComponentFiltersProps {
    before?: React.ReactFragment
    toggle: React.ReactFragment
    className?: string
}

export const CatalogExplorerViewOptionsRow: React.FunctionComponent<Props> = ({
    before,
    toggle,
    filters,
    filtersQueryParsed,
    onFiltersChange,
    onFiltersQueryFieldChange,
    className,
}) => {
    const [queryInput, setQueryInput] = useState<string>()
    const query = queryInput ?? filters.query

    const onQueryChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => setQueryInput(event.currentTarget.value),
        []
    )

    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        event => {
            event.preventDefault()
            onFiltersChange({ ...filters, query })
        },
        [filters, onFiltersChange, query]
    )

    const onQueryKindChange = useCallback(
        (value: ComponentKind | undefined): void => {
            setQueryInput(undefined)
            onFiltersQueryFieldChange('is', value)
        },
        [onFiltersQueryFieldChange]
    )

    return (
        <Form className={classNames('form-inline', styles.form, className)} onSubmit={onSubmit}>
            {before && <div>{before}</div>}
            {toggle}
            <CatalogExplorerViewOptionInput<ComponentKind>
                label="Type"
                values={[
                    ComponentKind.SERVICE,
                    ComponentKind.APPLICATION,
                    ComponentKind.LIBRARY,
                    ComponentKind.TOOL,
                    ComponentKind.WEBSITE,
                    ComponentKind.OTHER,
                ]}
                selected={filtersQueryParsed.is}
                onChange={onQueryKindChange}
                className={classNames('mb-0', styles.inputSelect)}
            />
            <div className={classNames('form-group mb-0 flex-grow-1')}>
                <label htmlFor="entity-list-filters__query" className="sr-only">
                    Query
                </label>
                <input
                    id="entity-list-filters__query"
                    className={classNames('form-control flex-1')}
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
