import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import CloseBoxIcon from 'mdi-react/CloseBoxIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import * as sourcegraph from 'sourcegraph'
import { Form } from '../../../../components/Form'
import { QueryParameterProps } from '../../../../components/withQueryParameter/WithQueryParameter'
import {
    ListHeaderQueryLinksButtonGroup,
    ListHeaderQueryLinksNav,
} from '../../../threads/components/ListHeaderQueryLinks'
import { DiagnosticQueryBuilderQuickFilterDropdownButton } from './DiagnosticQueryBuilderQuickFilterDropdownButton'
import { DiagnosticQueryBuilderRepositoryFilterDropdownButton } from './DiagnosticQueryBuilderRepositoryFilterDropdownButton'
import { DiagnosticQueryBuilderFilterDropdownButton } from './DiagnosticQueryBuilderTagFilterDropdownButton'

interface Props extends QueryParameterProps {
    parsedQuery: sourcegraph.DiagnosticQuery

    className?: string
    location: H.Location
}

const QUERY_FIELDS_IN_USE = ['is']

/**
 * A query builder for a diagnostic query.
 */
export const DiagnosticQueryBuilder: React.FunctionComponent<Props> = ({
    parsedQuery,
    query,
    onQueryChange,
    className = '',
    location,
}) => {
    const [uncommittedValue, setUncommittedValue] = useState(query)
    useEffect(() => setUncommittedValue(query), [query])

    const [isFocused, setIsFocused] = useState(false)
    const onFocus = useCallback(() => setIsFocused(true), [])
    const onBlur = useCallback(() => setIsFocused(false), [])

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
        <div className={`diagnostic-query-builder ${className}`}>
            <Form className="form d-flex align-items-stretch" onSubmit={onSubmit}>
                <DiagnosticQueryBuilderQuickFilterDropdownButton buttonClassName="px-3 border-top-0" />
                <div
                    className={`input-group bg-form-control border border-top-0 ${
                        isFocused ? 'form-control-focus' : ''
                    }`}
                >
                    <div className="input-group-prepend">
                        <span className="input-group-text border-0 pl-2 pr-1 bg-form-control">
                            <SearchIcon className="icon-inline" />
                        </span>
                    </div>
                    <input
                        type="text"
                        className="form-control border-0 pl-1"
                        aria-label="Filter diagnostics"
                        autoCapitalize="off"
                        value={uncommittedValue}
                        onChange={onChange}
                        onFocus={onFocus}
                        onBlur={onBlur}
                    />
                    <div className="input-group-append">
                        <Link className="btn btn-link">
                            <CloseBoxIcon className="icon-inline mr-2" />
                            Clear filters
                        </Link>
                    </div>
                </div>
            </Form>
            <div className="d-flex align-items-center">
                <ListHeaderQueryLinksNav
                    query={query}
                    links={[
                        {
                            icon: AlertCircleOutlineIcon,
                            label: 'open',
                            count: 12,
                            queryField: 'is',
                            queryValues: ['open'], // TODO!(sqs): un-hardcode
                            removeQueryFields: QUERY_FIELDS_IN_USE,
                        },
                        {
                            icon: CheckIcon,
                            label: 'pending fix',
                            count: 7,
                            queryField: 'is',
                            queryValues: ['pending'], // TODO!(sqs): un-hardcode
                            removeQueryFields: QUERY_FIELDS_IN_USE,
                        },
                    ]}
                    location={location}
                    className="mr-4"
                    itemClassName="p-3"
                    itemActiveClassName="font-weight-bold"
                    itemInactiveClassName="btn-link"
                />
                <DiagnosticQueryBuilderFilterDropdownButton
                    items={[
                        { text: 'no-undef', count: 41 },
                        { text: 'sourcegraph/import-submodule', count: 21 },
                        { text: 'semicolon', count: 17 },
                        { text: 'react-hooks/deps', count: 11 },
                        { text: 'module', count: 3 },
                    ]}
                    pluralNoun="tags"
                    buttonText="Tags"
                    headerText="Filter by tag"
                    queryPlaceholderText="Filter tags"
                    buttonClassName="btn-link"
                />
                <DiagnosticQueryBuilderFilterDropdownButton
                    items={[
                        { text: 'github.com/sourcegraph/sourcegraph', count: 41 },
                        { text: 'github.com/sourcegraph/codeintellify', count: 21 },
                        { text: 'github.com/sourcegraph/javascript-typescript-langserver', count: 17 },
                        { text: 'github.com/sourcegraph/event-positions', count: 11 },
                        { text: 'github.com/sourcegraph/icons', count: 3 },
                    ]}
                    pluralNoun="repositories"
                    buttonText="Repositories"
                    headerText="Filter by repository"
                    queryPlaceholderText="Filter repositories"
                    buttonClassName="btn-link"
                />
            </div>
        </div>
    )
}
