import { ChangeEvent, FunctionComponent, memo, useState } from 'react'

import { gql, useQuery } from '@apollo/client'
import { Combobox, ComboboxInput, ComboboxList, ComboboxOption, ComboboxOptionText } from '@reach/combobox'
import classNames from 'classnames'
import { noop } from 'lodash'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isDefined } from '@sourcegraph/common'
import { FormInputProps, Link, LoadingSpinner, useDebounce } from '@sourcegraph/wildcard'

import { GetSearchContextsResult } from '../../../../../../../../../graphql-operations'
import { TruncatedText } from '../../../../../../trancated-text/TruncatedText'
import { DrillDownInput } from '../drill-down-input/DrillDownInput'

import styles from './DrillDownSearchContextFilter.module.scss'

export const SEARCH_CONTEXT_GQL = gql`
    query GetSearchContexts($query: String!) {
        searchContexts(query: $query) {
            nodes {
                id
                spec
                query
                description
            }
            pageInfo {
                hasNextPage
            }
        }
    }
`

interface DrillDownSearchContextFilter extends FormInputProps {}

export const DrillDownSearchContextFilter: FunctionComponent<DrillDownSearchContextFilter> = props => {
    const { value = '', className, onChange = noop, ...attributes } = props
    const [showSuggest, setShowSuggest] = useState<boolean>(true)
    const debouncedQuery = useDebounce(value, 700)

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

            {showSuggest && (
                <ComboboxList className={styles.suggestionList}>
                    <SuggestPanel query={debouncedQuery.toString()} />
                </ComboboxList>
            )}
        </Combobox>
    )
}

interface SuggestPanelProps {
    query: string
}

const SuggestPanel = memo<SuggestPanelProps>(props => {
    const { query } = props

    const { data, loading, error } = useQuery<GetSearchContextsResult>(SEARCH_CONTEXT_GQL, {
        fetchPolicy: 'network-only',
        variables: { query },
    })

    const queryBasedContexts =
        data?.searchContexts.nodes.filter(node => isDefined(node.query) && node.query !== '') ?? []

    if (loading) {
        return <LoadingSpinner />
    }

    if (error) {
        return <ErrorAlert error={error} />
    }

    if (!data) {
        return null
    }

    if (queryBasedContexts.length === 0) {
        return (
            <span className={styles.suggestNoDataFound}>
                No query-based search contexts found.{' '}
                <Link to="/contexts/new" rel="noreferrer noopener" target="_blank">
                    Create search context
                </Link>
            </span>
        )
    }

    return (
        <>
            {queryBasedContexts.map(context => (
                <ComboboxOption key={context.id} value={context.spec} className={styles.suggestItem}>
                    <TruncatedText as="small" className={styles.suggestItemName}>
                        <ComboboxOptionText />
                    </TruncatedText>

                    <TruncatedText as="small" className={styles.suggestItemDescription}>
                        {context.description}
                    </TruncatedText>
                </ComboboxOption>
            ))}
        </>
    )
})
