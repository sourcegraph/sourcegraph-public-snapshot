import * as React from 'react'
import { Link } from 'react-router-dom'
import { buildSearchURLQuery } from './index'

interface Props {
    query: GQL.ISearchQuery

    /**
     * Called when the user mouseups on the component. This captures normal
     * clicks and middle-clicks.
     */
    onMouseUp?: () => void
}

export const QueryButton: React.StatelessComponent<Props> = (props: Props) => (
    <Link
        className="query-button"
        to={'/search?' + buildSearchURLQuery(props.query)}
        title={`${props.query.scopeQuery} ${props.query.query}`}
        onMouseUp={props.onMouseUp}
    >
        {props.query.scopeQuery && <span className="query-button__scope">{props.query.scopeQuery}</span>}{' '}
        {props.query.query}
    </Link>
)
