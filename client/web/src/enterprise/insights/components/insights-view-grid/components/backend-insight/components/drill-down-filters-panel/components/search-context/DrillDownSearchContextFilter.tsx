import React, { ChangeEvent, useState } from 'react'

import { gql, useQuery } from '@apollo/client'
import { Combobox, ComboboxInput, ComboboxList, ComboboxOption, ComboboxOptionText } from '@reach/combobox'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Link, LoadingSpinner } from '@sourcegraph/wildcard'

import { GetSearchContextsResult } from '../../../../../../../../../../graphql-operations'
import { DrillDownRegExpInput } from '../drill-down-reg-exp-input/DrillDownRegExpInput'

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

interface DrillDownSearchContextFilter {}

interface SearchContextState {
    query: string
    showSuggest: boolean
}

const INITIAL_STATE: SearchContextState = {
    query: '',
    showSuggest: true,
}

export const DrillDownSearchContextFilter: React.FunctionComponent<DrillDownSearchContextFilter> = props => {
    const [searchState, setSearchState] = useState<SearchContextState>(INITIAL_STATE)

    const handleSelect = (value: string): void => {
        console.log('select')

        setSearchState({
            query: value,
            showSuggest: false,
        })
    }

    const handleChange = (event: ChangeEvent<HTMLInputElement>): void => {
        console.log('input change')

        setSearchState({
            query: event.target.value,
            showSuggest: true,
        })
    }

    return (
        <Combobox onSelect={handleSelect}>
            <ComboboxInput
                as={DrillDownRegExpInput}
                placeholder="global (default)"
                prefix="context:"
                value={searchState.query}
                onChange={handleChange}
                className={styles.input}
            />

            {searchState.showSuggest && <SuggestPanel query={searchState.query} />}
        </Combobox>
    )
}

interface SuggestPanelProps {
    query: string
}

const SuggestPanel: React.FunctionComponent<SuggestPanelProps> = props => {
    const { query } = props

    const { data, loading, error } = useQuery<GetSearchContextsResult>(SEARCH_CONTEXT_GQL, {
        variables: { query },
    })

    return (
        <ComboboxList className={styles.suggestionList}>
            {loading ? (
                <LoadingSpinner />
            ) : error ? (
                <ErrorAlert error={error} />
            ) : data ? (
                data.searchContexts.nodes.length === 0 ? (
                    <span className={styles.suggestNoDataFound}>
                        No query-based search contexts found. <Link to="/contexts/new">Create search context</Link>
                    </span>
                ) : (
                    data.searchContexts.nodes.map(context => (
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
