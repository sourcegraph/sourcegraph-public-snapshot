import React, { ReactNode } from 'react'

import { renderMarkdown } from '@sourcegraph/common'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Button, Link, Alert, Typography } from '@sourcegraph/wildcard'

import { SearchPatternType } from '../../graphql-operations'

interface SearchAlertProps {
    alert: Required<AggregateStreamingSearchResults>['alert']
    patternType: SearchPatternType | undefined
    caseSensitive: boolean
    searchContextSpec?: string
    children?: ReactNode[]
}

export const SearchAlert: React.FunctionComponent<React.PropsWithChildren<SearchAlertProps>> = ({
    alert,
    patternType,
    caseSensitive,
    searchContextSpec,
    children,
}) => (
    <Alert className="my-2 mr-3" data-testid="alert-container" variant="info">
        <Typography.H3>{alert.title}</Typography.H3>

        {alert.description && <Markdown className="mb-3" dangerousInnerHTML={renderMarkdown(alert.description)} />}

        {alert.proposedQueries && (
            <>
                <Typography.H4>Did you mean:</Typography.H4>
                <ul className="list-unstyled">
                    {alert.proposedQueries.map(proposedQuery => (
                        <li key={proposedQuery.query}>
                            <Button
                                data-testid="proposed-query-link"
                                to={
                                    '/search?' +
                                    buildSearchURLQuery(
                                        proposedQuery.query,
                                        patternType || SearchPatternType.literal,
                                        caseSensitive,
                                        searchContextSpec
                                    )
                                }
                                variant="secondary"
                                size="sm"
                                as={Link}
                            >
                                {proposedQuery.query || proposedQuery.description}
                            </Button>
                            {proposedQuery.query && proposedQuery.description && ` — ${proposedQuery.description}`}
                        </li>
                    ))}
                </ul>
            </>
        )}

        {children}
    </Alert>
)
