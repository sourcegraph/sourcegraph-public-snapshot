import { ChangeEvent, FunctionComponent, useState } from 'react'

import { gql, useQuery } from '@apollo/client'
import { Combobox, ComboboxInput, ComboboxList, ComboboxOption, ComboboxOptionText } from '@reach/combobox'
import classNames from 'classnames'
import { noop } from 'lodash'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isDefined } from '@sourcegraph/common'
import { InputProps, Link, LoadingSpinner, useDebounce } from '@sourcegraph/wildcard'

import { GetSearchContextsResult } from '../../../../../../../../../graphql-operations'
import { DrillDownInput } from '../drill-down-input/DrillDownInput'

import styles from './DrillDownSearchContextFilter.module.scss'

export const SEARCH_CONTEXT_GQL = gql`
    query GetSearchContexts($query: String!) {
        searchContexts(first: 10, query: $query) {
            nodes {
                id
                name
                query
                description
            }
            pageInfo {
                hasNextPage
            }
        }
    }
`

interface DrillDownSearchContextFilter extends InputProps {}

export const DrillDownSearchContextFilter: FunctionComponent<DrillDownSearchContextFilter> = props => {
    const { value = '', className, onChange = noop, ...attributes } = props
    const [showSuggest, setShowSuggest] = useState<boolean>(true)

    const handleSelect = (value: string): void => {
        setShowSuggest(false)
        onChange(value)
    }

    const handleChange = (event: ChangeEvent<HTMLInputElement>): void => {
        setShowSuggest(true)
        onChange(event)
    }

    return (
        <Combobox onSelect={handleSelect}>
            <ComboboxInput
                {...attributes}
                as={DrillDownInput}
                placeholder="global (default)"
                prefix="context:"
                value={value.toString()}
                className={classNames(className, styles.input)}
                onChange={handleChange}
            />

            {showSuggest && <SuggestPanel query={value.toString()} />}
        </Combobox>
    )
}

interface SuggestPanelProps {
    query: string
}

const SuggestPanel: FunctionComponent<SuggestPanelProps> = props => {
    const { query } = props

    const debouncedQuery = useDebounce(query, 700)
    const { data, loading, error } = useQuery<GetSearchContextsResult>(SEARCH_CONTEXT_GQL, {
        variables: { query: debouncedQuery },
    })

    const queryBasedContexts =
        data?.searchContexts.nodes.filter(node => isDefined(node.query) && node.query !== '') ?? []

    return (
        <ComboboxList className={styles.suggestionList}>
            {loading ? (
                <LoadingSpinner />
            ) : error ? (
                <ErrorAlert error={error} />
            ) : data ? (
                queryBasedContexts.length === 0 ? (
                    <span className={styles.suggestNoDataFound}>
                        No query-based search contexts found. <Link to="/contexts/new">Create search context</Link>
                    </span>
                ) : (
                    queryBasedContexts.map(context => (
                        <ComboboxOption key={context.id} value={context.name} className={styles.suggestItem}>
                            <small className={styles.suggestItemName}>
                                <ComboboxOptionText />
                            </small>
                            <small className={styles.suggestItemDescription}>{context.description}</small>
                        </ComboboxOption>
                    ))
                )
            ) : null}
        </ComboboxList>
    )
}
