import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'
import { SearchPatternType } from '../../graphql-operations'
import { AggregateStreamingSearchResults } from '../stream'

interface SearchAlertProps {
    alert: Required<AggregateStreamingSearchResults>['alert']
    patternType: SearchPatternType | undefined
    caseSensitive: boolean
    versionContext?: string
    children?: ReactNode[]
}

export const SearchAlert: React.FunctionComponent<SearchAlertProps> = ({
    alert,
    patternType,
    caseSensitive,
    versionContext,
    children,
}) => (
    <div className="alert alert-info m-2" data-testid="alert-container">
        <h3>
            <AlertCircleIcon className="icon-inline" /> {alert.title}
        </h3>
        <p>{alert.description}</p>

        {alert.proposedQueries && (
            <>
                <h4>Did you mean:</h4>
                <ul className="list-unstyled">
                    {alert.proposedQueries.map(proposedQuery => (
                        <li key={proposedQuery.query}>
                            <Link
                                className="btn btn-secondary btn-sm"
                                data-testid="proposed-query-link"
                                to={
                                    '/search?' +
                                    buildSearchURLQuery(
                                        proposedQuery.query,
                                        patternType || SearchPatternType.literal,
                                        caseSensitive,
                                        versionContext,
                                        {}
                                    )
                                }
                            >
                                {proposedQuery.query || proposedQuery.description}
                            </Link>
                            {proposedQuery.query && proposedQuery.description && ` — ${proposedQuery.description}`}
                        </li>
                    ))}
                </ul>
            </>
        )}

        {children}
    </div>
)
