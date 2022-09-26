import { MouseEvent, useCallback } from 'react'

import { mdiArrowRight, mdiChevronDown, mdiChevronUp } from '@mdi/js'

import { formatSearchParameters, renderMarkdown } from '@sourcegraph/common'
import { SyntaxHighlightedSearchQuery, smartSearchIconSvgPath } from '@sourcegraph/search-ui'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import {
    Link,
    createLinkUrl,
    Icon,
    Collapse,
    CollapseHeader,
    CollapsePanel,
    H2,
    Text,
    Button,
} from '@sourcegraph/wildcard'

import { SearchPatternType } from '../../graphql-operations'

import styles from './QuerySuggestion.module.scss'

interface SmartSearchProps {
    alert: Required<AggregateStreamingSearchResults>['alert'] | undefined
    onDisableSmartSearch: () => void
}

const processDescription = (description: string): string => {
    const split = description.split(' ⚬ ')

    split[0] = split[0][0].toUpperCase() + split[0].slice(1)
    return split.join(', ')
}

export const SmartSearch: React.FunctionComponent<React.PropsWithChildren<SmartSearchProps>> = ({
    alert,
    onDisableSmartSearch,
}) => {
    const [isCollapsed, setIsCollapsed] = useTemporarySetting('search.results.collapseSmartSearch')

    const disableSmartSearch = useCallback(
        (event: MouseEvent) => {
            event.stopPropagation() // Don't trigger the collapse toggle
            onDisableSmartSearch()
        },
        [onDisableSmartSearch]
    )

    return alert?.kind && alert.kind !== 'lucky-search-queries' ? null : (
        <div className={styles.root}>
            <Collapse isOpen={!isCollapsed} onOpenChange={opened => setIsCollapsed(!opened)}>
                <CollapseHeader className={styles.collapseButton}>
                    <div className={styles.header}>
                        <span className="d-flex align-items-center">
                            <Icon aria-hidden={true} svgPath={smartSearchIconSvgPath} className={styles.smartIcon} />
                            <span>
                                <H2 className={styles.title}>
                                    <Markdown
                                        wrapper="span"
                                        dangerousInnerHTML={renderMarkdown(
                                            alert?.title || '**Smart Search** is also showing additional results.'
                                        )}
                                    />
                                </H2>
                                <span className="text-muted">
                                    Don't want these?{' '}
                                    <Button
                                        variant="link"
                                        size="sm"
                                        className={styles.disableButton}
                                        onClick={disableSmartSearch}
                                    >
                                        Disable <b>Smart Search</b>
                                    </Button>
                                </span>
                            </span>
                        </span>
                        {isCollapsed ? (
                            <Icon aria-label="Expand" svgPath={mdiChevronDown} />
                        ) : (
                            <Icon aria-label="Collapse" svgPath={mdiChevronUp} />
                        )}
                    </div>
                </CollapseHeader>
                <CollapsePanel>
                    <Text className={styles.description}>{alert?.description}</Text>
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
                                        <span className={styles.listItemDescription}>{`${processDescription(
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

export const smartSearchEvent = (alertTitle: string, descriptions: string[]): string[] => {
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

    const prefix = alertTitle.match(/your query found \*\*no results\*\*/)
        ? 'SearchResultsAutoPure'
        : 'SearchResultsAutoAdded'

    const events = []
    for (const rule of rules) {
        events.push(`${prefix}${rule}`)
    }
    return events
}
