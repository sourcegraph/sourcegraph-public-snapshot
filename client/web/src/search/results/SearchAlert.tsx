import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { FilterType } from '../../../../shared/src/search/query/filters'
import { Filter } from '../../../../shared/src/search/query/token'
import { FilterKind, findFilter } from '../../../../shared/src/search/query/validate'
import { replaceRange } from '../../../../shared/src/util/strings'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'
import { SearchPatternType } from '../../graphql-operations'
import { AggregateStreamingSearchResults } from '../stream'

interface SearchAlertProps {
    alert: Required<AggregateStreamingSearchResults>['alert']
    patternType: SearchPatternType | undefined
    caseSensitive: boolean
    versionContext?: string
    searchContextSpec?: string
    children?: ReactNode[]
}

const getGlobalContextFilterToken = (query: string): Filter | undefined =>
    findFilter(query, FilterType.context, FilterKind.Global)

const getContextFilterValueFromQuery = (query: string): string | undefined => {
    const token = getGlobalContextFilterToken(query)
    if (token?.value?.type === 'literal') {
        return token.value.value
    }
    if (token?.value?.type === 'quoted') {
        return token.value.quotedValue
    }
    return undefined
}

const omitContextFilterFromQuery = (query: string): string => {
    const token = getGlobalContextFilterToken(query)
    return token ? replaceRange(query, token.range) : query
}

export const SearchAlert: React.FunctionComponent<SearchAlertProps> = ({
    alert,
    patternType,
    caseSensitive,
    versionContext,
    searchContextSpec,
    children,
}) => (
    <div className="alert alert-info m-2" data-testid="alert-container">
        <h3>
            <AlertCircleIcon className="icon-inline" /> {alert.title}
        </h3>
        <p>{alert.description}</p>

        {alert.proposedQueries && (
            <>
                <h4>Try instead:</h4>
                <ul className="list-unstyled">
                    {alert.proposedQueries.map(proposedQuery => (
                        <li key={proposedQuery.query}>
                            <Link
                                className="btn btn-secondary btn-sm"
                                data-testid="proposed-query-link"
                                to={
                                    '/search?' +
                                    buildSearchURLQuery(
                                        omitContextFilterFromQuery(proposedQuery.query),
                                        patternType || SearchPatternType.literal,
                                        caseSensitive,
                                        versionContext,
                                        getContextFilterValueFromQuery(proposedQuery.query) || searchContextSpec
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
