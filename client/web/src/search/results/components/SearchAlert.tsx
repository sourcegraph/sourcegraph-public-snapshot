import React, { type ReactNode } from 'react'

import { renderMarkdown } from '@sourcegraph/common'
import type { AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Button, Link, Alert, H3, H4, Markdown } from '@sourcegraph/wildcard'

import { SearchPatternType } from '../../../graphql-operations'

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
    <Alert className="my-2" data-testid="alert-container" variant="info">
        <H3>{alert.title}</H3>

        {alert.description && (
            <Markdown
                className="mb-3"
                dangerousInnerHTML={renderMarkdown(alert.description, {
                    // Disable autolinks so revision specifications are not rendered as email links
                    // (for example, "sourcegraph@4.0.1")
                    disableAutolinks: true,
                })}
            />
        )}

        {alert.proposedQueries && (
            <>
                <H4>Did you mean:</H4>
                <ul className="list-unstyled">
                    {alert.proposedQueries.map(proposedQuery => (
                        <li key={proposedQuery.query}>
                            <Button
                                data-testid="proposed-query-link"
                                to={
                                    '/search?' +
                                    buildSearchURLQuery(
                                        proposedQuery.query,
                                        patternType || SearchPatternType.standard,
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
