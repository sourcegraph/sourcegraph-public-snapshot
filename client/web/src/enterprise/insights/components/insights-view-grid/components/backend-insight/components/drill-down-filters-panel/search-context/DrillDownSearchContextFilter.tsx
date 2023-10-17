import { type ChangeEvent, type FocusEvent, forwardRef, type InputHTMLAttributes, memo, useState } from 'react'

import { gql, useQuery } from '@apollo/client'
import { Combobox, ComboboxInput, ComboboxList, ComboboxOption, ComboboxOptionText } from '@reach/combobox'
import classNames from 'classnames'
import { noop } from 'lodash'

import { isDefined } from '@sourcegraph/common'
import { Link, LoadingSpinner, useDebounce, ErrorAlert, type InputProps } from '@sourcegraph/wildcard'

import type { GetSearchContextsResult } from '../../../../../../../../../graphql-operations'
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

interface DrillDownSearchContextFilter extends InputProps, InputHTMLAttributes<HTMLInputElement> {}

export const DrillDownSearchContextFilter = forwardRef<HTMLInputElement, DrillDownSearchContextFilter>(
    (props, reference) => {
        const { value = '', className, onChange = noop, onFocus = noop, ...attributes } = props
        const [showSuggest, setShowSuggest] = useState<boolean>(false)
        const debouncedQuery = useDebounce(value, 700)

        const handleSelect = (value: string): void => {
            setShowSuggest(false)
            onChange(value)
        }

        const handleChange = (event: ChangeEvent<HTMLInputElement>): void => {
            setShowSuggest(true)
            onChange(event)
        }

        const handleFocus = (event: FocusEvent<HTMLInputElement>): void => {
            setShowSuggest(true)
            onFocus(event)
        }

        return (
            <Combobox onSelect={handleSelect}>
                <ComboboxInput
                    {...attributes}
                    ref={reference}
                    as={DrillDownInput}
                    placeholder="global (default)"
                    prefix="context:"
                    value={value.toString()}
                    className={classNames(className, styles.input)}
                    onChange={handleChange}
                    onFocus={handleFocus}
                />

                {showSuggest && (
                    <ComboboxList className={styles.suggestionList}>
                        <SuggestPanel query={debouncedQuery.toString()} />
                    </ComboboxList>
                )}
            </Combobox>
        )
    }
)

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
