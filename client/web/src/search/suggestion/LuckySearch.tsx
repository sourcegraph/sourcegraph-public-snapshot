import { mdiArrowRight, mdiChevronDown, mdiChevronUp, mdiHelpCircleOutline } from '@mdi/js'

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

const processDescription = (description: string): string => {
    const split = description.split(' ⚬ ')

    split[0] = split[0][0].toUpperCase() + split[0].slice(1)
    return split.join(', ')
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
                                    svgPath={mdiHelpCircleOutline}
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
                            <li key={entry.query} className={styles.listItem}>
                                <Link
                                    to={createLinkUrl({
                                        pathname: '/search',
                                        search: formatSearchParameters(new URLSearchParams({ q: entry.query })),
                                    })}
                                    className={styles.link}
                                >
                                    <span>
                                        <span className={styles.description}>{`${processDescription(
                                            entry.description || ''
                                        )}`}</span>
                                        {entry.annotations
                                            ?.filter(({ name }) => name === 'ResultCount')
                                            ?.map(({ name, value }) => (
                                                <span key={name} className="text-muted">
                                                    {' '}
                                                    ({value})
                                                </span>
                                            ))}
                                    </span>
                                    <Icon svgPath={mdiArrowRight} aria-hidden={true} className="mx-2 text-body" />
                                    <span className={styles.suggestion}>
                                        <SyntaxHighlightedSearchQuery
                                            query={entry.query}
                                            searchPatternType={SearchPatternType.standard}
                                        />
                                    </span>
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
