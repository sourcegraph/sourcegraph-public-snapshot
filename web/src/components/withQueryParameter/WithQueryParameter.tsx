import H from 'history'
import React, { useCallback } from 'react'

/**
 * React component props for children of {@link WithQueryParameter}.
 */
export interface QueryParameterProps {
    /** The query. */
    query: string

    /** Called when the query changes. */
    onQueryChange: (query: string) => void
}

interface Props {
    defaultQuery?: string
    history: H.History
    location: H.Location
    children: (props: QueryParameterProps) => JSX.Element | null
}

/**
 * Wraps a component and provides `query` and `onQueryChange` properties that read/write from the
 * URL query string's 'q' parameter.
 */
export const WithQueryParameter: React.FunctionComponent<Props> = ({
    defaultQuery = '',
    history,
    location,
    children,
}) => {
    const q = new URLSearchParams(location.search).get('q')
    const query = q === null ? defaultQuery : q
    const onQueryChange = useCallback(
        (query: string) => {
            const params = new URLSearchParams(location.search)
            params.set('q', query)
            history.push({ search: `${params}` })
        },
        [history, location.search]
    )

    return children({ query, onQueryChange })
}

export const useQueryParameter = (
    { location, history }: { location: H.Location; history: H.History },
    defaultQuery = ''
): [string, (query: string) => void] => {
    const q = new URLSearchParams(location.search).get('q')
    const query = q === null ? defaultQuery : q
    const onQueryChange = useCallback(
        (query: string) => {
            const params = new URLSearchParams(location.search)
            params.set('q', query)
            history.push({ search: `${params}` })
        },
        [history, location.search]
    )
    return [query, onQueryChange]
}

export const withQueryParameter = <P extends object>(
    component: React.FunctionComponent<P & QueryParameterProps>
): React.FunctionComponent<P & Pick<Props, Exclude<keyof Props, 'children'>>> => {
    const Component = component
    return props => (
        <WithQueryParameter {...props}>
            {queryParameterProps => <Component {...props} {...queryParameterProps} />}
        </WithQueryParameter>
    )
}
