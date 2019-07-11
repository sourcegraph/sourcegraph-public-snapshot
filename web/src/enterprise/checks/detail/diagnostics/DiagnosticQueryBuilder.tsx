import React, { useCallback, useEffect, useState } from 'react'
import * as sourcegraph from 'sourcegraph'
import { Form } from '../../../../components/Form'
import { QueryParameterProps } from '../../../../components/withQueryParameter/WithQueryParameter'
import { DiagnosticQueryBuilderFilterDropdownButton } from './DiagnosticQueryBuilderFilterDropdownButton'

interface Props extends QueryParameterProps {
    parsedQuery: sourcegraph.DiagnosticQuery

    className?: string
}

/**
 * A query builder for a diagnostic query.
 */
export const DiagnosticQueryBuilder: React.FunctionComponent<Props> = ({
    parsedQuery,
    query,
    onQueryChange,
    className = '',
}) => {
    const [uncommittedValue, setUncommittedValue] = useState(query)
    useEffect(() => setUncommittedValue(query), [query])

    const onSubmit = useCallback<React.FormEventHandler>(
        e => {
            e.preventDefault()
            onQueryChange(uncommittedValue)
        },
        [onQueryChange, uncommittedValue]
    )
    const onChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => setUncommittedValue(e.currentTarget.value),
        []
    )

    return (
        <Form className={`form ${className}`} onSubmit={onSubmit}>
            <div className="input-group">
                <div className="input-group-prepend">
                    <DiagnosticQueryBuilderFilterDropdownButton />
                </div>
                <input
                    type="text"
                    className="form-control"
                    aria-label="Filter diagnostics"
                    autoCapitalize="off"
                    value={uncommittedValue}
                    onChange={onChange}
                />
            </div>
        </Form>
    )
}
