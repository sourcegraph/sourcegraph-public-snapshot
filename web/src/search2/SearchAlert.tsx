import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { buildSearchURLQuery } from './index'

interface Props {
    title: string
    description?: string
    proposedQueries?: GQL.ISearchQuery2Description[]

    location: H.Location
}

interface State {
    alert: GQL.ISearchAlert | null
}

export class SearchAlert extends React.Component<Props, State> {

    public state: State = {
        alert: null,
    }

    public render(): JSX.Element | null {
        return (
            <div className='search-alert'>
                <div className='search-alert__message'>
                    {<h2 className='search-alert__title'>{this.props.title}</h2>}
                    {this.props.description && <p className='search-alert__description'>{this.props.description}</p>}
                </div>
                {
                    this.props.proposedQueries &&
                    <ul className='search-alert__proposed-queries'>{
                        this.props.proposedQueries.map((proposedQuery, i) =>
                            <li key={i} className='search-alert__proposed-query'>
                                <span className='search-alert__proposed-query-did-you-mean'>
                                    Did you mean:{' '}
                                </span>
                                <Link
                                    className='search-alert__proposed-query-link'
                                    to={'?' + buildSearchURLQuery(proposedQuery.query)}
                                    title={`${proposedQuery.query.scopeQuery} ${proposedQuery.query.query}`}
                                >
                                    {proposedQuery.query.scopeQuery} {proposedQuery.query.query}
                                </Link>
                                <span className='search-alert__proposed-query-description'>
                                    {proposedQuery.description && ` — ${proposedQuery.description}`}
                                </span>
                            </li>
                        )
                    }</ul>
                }
            </div>
        )
    }
}
