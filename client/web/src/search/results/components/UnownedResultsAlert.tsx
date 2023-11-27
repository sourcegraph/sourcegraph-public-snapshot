import React, { useMemo } from 'react'

import { mdiChevronDown, mdiChevronUp } from '@mdi/js'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded'
import { renderMarkdown } from '@sourcegraph/common'
import type {
    SearchPatternTypeProps,
    CaseSensitivityProps,
    SearchContextProps,
    QueryState,
} from '@sourcegraph/shared/src/search'
import { findFilter, FilterKind } from '@sourcegraph/shared/src/search/query/query'
import { omitFilter, appendFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import {
    Alert,
    Collapse,
    H3,
    CollapseHeader,
    Button,
    Icon,
    CollapsePanel,
    Markdown,
    Text,
    Link,
} from '@sourcegraph/wildcard'

import { SearchPatternType } from '../../../graphql-operations'
import { buildSearchURLQueryFromQueryState } from '../../../stores'

export interface UnownedResultsAlertProps
    extends SearchPatternTypeProps,
        Pick<CaseSensitivityProps, 'caseSensitive'>,
        Pick<SearchContextProps, 'selectedSearchContextSpec'> {
    alertTitle: string
    alertDescription?: string | null
    queryState: QueryState
}

export const UnownedResultsAlert: React.FunctionComponent<React.PropsWithChildren<UnownedResultsAlertProps>> = ({
    alertTitle,
    alertDescription,
    patternType,
    caseSensitive,
    selectedSearchContextSpec,
    queryState,
}) => {
    const [isCollapsed, setIsCollapsed] = useTemporarySetting('search.results.collapseUnownedResultsAlert')

    const unownedFilesSearchLink = useMemo(() => {
        if (!queryState) {
            return ''
        }

        let query = queryState.query
        const selectFilter = findFilter(queryState.query, 'select', FilterKind.Global)
        if (selectFilter && selectFilter.value?.value === 'file.owners') {
            query = omitFilter(query, selectFilter)
        }
        query = appendFilter(query, '-file', 'has.owner()')

        const searchParams = buildSearchURLQueryFromQueryState({
            query,
            caseSensitive,
            patternType,
            searchContextSpec: selectedSearchContextSpec,
        })
        return `/search?${searchParams}`
    }, [queryState, patternType, caseSensitive, selectedSearchContextSpec])

    return (
        <Alert className="my-2" variant="info">
            <Collapse isOpen={!isCollapsed} onOpenChange={opened => setIsCollapsed(!opened)}>
                <CollapseHeader
                    as={Button}
                    outline={false}
                    className="w-100 align-items-center justify-content-between"
                    aria-label={isCollapsed ? 'Show alert details' : 'Hide alert details'}
                    variant="icon"
                >
                    <H3 className="mb-0">{alertTitle}</H3>
                    <Icon aria-hidden={true} svgPath={isCollapsed ? mdiChevronDown : mdiChevronUp} />
                </CollapseHeader>
                <CollapsePanel className="mt-2">
                    {alertDescription && (
                        <Markdown
                            className="mb-2"
                            dangerousInnerHTML={renderMarkdown(alertDescription, {
                                // Disable autolinks so revision specifications are not rendered as email links
                                // (for example, "sourcegraph@4.0.1")
                                disableAutolinks: true,
                            })}
                        />
                    )}
                    <Text className="d-flex align-items-baseline mb-0">
                        <span>See unowned files:</span>
                        <small>
                            <Button
                                variant="secondary"
                                outline={true}
                                // Use a link so a new tab with the search can be opened easily.
                                as={Link}
                                to={unownedFilesSearchLink}
                                size="sm"
                                className="ml-2"
                            >
                                <SyntaxHighlightedSearchQuery
                                    query="-file:has.owner()"
                                    searchPatternType={SearchPatternType.standard}
                                />
                            </Button>
                        </small>
                    </Text>
                </CollapsePanel>
            </Collapse>
        </Alert>
    )
}
