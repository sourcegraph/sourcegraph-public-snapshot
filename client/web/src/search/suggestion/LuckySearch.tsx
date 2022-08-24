import { mdiChevronDown, mdiChevronUp, mdiInformationOutline } from '@mdi/js'

import { formatSearchParameters } from '@sourcegraph/common'
import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { Link, H3, createLinkUrl, Tooltip, Icon, Collapse, CollapseHeader, CollapsePanel } from '@sourcegraph/wildcard'

import { SearchPatternType } from '../../graphql-operations'

import styles from './QuerySuggestion.module.scss'

interface LuckySearchProps {
    alert: Required<AggregateStreamingSearchResults>['alert'] | undefined
}

export const LuckySearch: React.FunctionComponent<React.PropsWithChildren<LuckySearchProps>> = ({ alert }) => {
    const [isCollapsed, setIsCollapsed] = useTemporarySetting('search.results.collapseLuckySearch')

    return alert?.kind && alert.kind !== 'lucky-search-queries' ? null : (
        <div className={styles.root}>
            <Collapse isOpen={!isCollapsed} onOpenChange={opened => setIsCollapsed(!opened)}>
                <CollapseHeader className={styles.collapseButton}>
                    <H3 className={styles.header}>
                        <span>
                            {alert?.title || 'Also showing additional results'}
                            <Tooltip
                                content={
                                    alert?.description ||
                                    'We returned all the results for your query. We also added results for similar queries that might interest you.'
                                }
                            >
                                <Icon
                                    className="ml-1"
                                    tabIndex={0}
                                    aria-label="More information"
                                    svgPath={mdiInformationOutline}
                                />
                            </Tooltip>
                        </span>
                        {isCollapsed ? (
                            <Icon aria-label="Expand" svgPath={mdiChevronDown} />
                        ) : (
                            <Icon aria-label="Collapse" svgPath={mdiChevronUp} />
                        )}
                    </H3>
                </CollapseHeader>
                <CollapsePanel>
                    <ul className={styles.container}>
                        {alert?.proposedQueries?.map(entry => (
                            <li key={entry.query}>
                                <Link
                                    to={createLinkUrl({
                                        pathname: '/search',
                                        search: formatSearchParameters(new URLSearchParams({ q: entry.query })),
                                    })}
                                >
                                    <span className={styles.suggestion}>
                                        <SyntaxHighlightedSearchQuery
                                            query={entry.query}
                                            searchPatternType={SearchPatternType.standard}
                                        />
                                    </span>
                                    <i>{`— ${entry.description}`}</i>
                                </Link>
                            </li>
                        ))}
                    </ul>
                </CollapsePanel>
            </Collapse>
        </div>
    )
}

export const luckySearchEvent = (alertTitle: string, descriptions: string[]): string[] => {
    const rules = descriptions.map(entry => {
        if (entry.match(/patterns as regular expressions/)) {
            return 'Regexp'
        }
        if (entry.match(/unquote patterns/)) {
            return 'Unquote'
        }
        if (entry.match(/AND patterns together/)) {
            return 'And'
        }
        if (entry.match(/language filter for pattern/)) {
            return 'Lang'
        }
        if (entry.match(/search type for pattern/)) {
            return 'Type'
        }
        return 'Other'
    })

    const prefix = alertTitle.match(/No results for original query/)
        ? 'SearchResultsAutoPure'
        : 'SearchResultsAutoAdded'

    const events = []
    for (const rule of rules) {
        events.push(`${prefix}${rule}`)
    }
    return events
}
