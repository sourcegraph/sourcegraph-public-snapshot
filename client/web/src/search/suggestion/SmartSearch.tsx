import { MouseEvent, useCallback } from 'react'

import { mdiChevronDown, mdiChevronUp } from '@mdi/js'

import { smartSearchIconSvgPath } from '@sourcegraph/branded'
import { SmartSearchListItem } from '@sourcegraph/branded/src/search-ui/components/SmartSearchListItem'
import { pluralize } from '@sourcegraph/common'
import { AggregateStreamingSearchResults, AlertKind, SmartSearchAlertKind } from '@sourcegraph/shared/src/search/stream'
import { Icon, Collapse, CollapseHeader, CollapsePanel, H2, Text, Button } from '@sourcegraph/wildcard'

import styles from './QuerySuggestion.module.scss'

interface SmartSearchProps {
    alert: Required<AggregateStreamingSearchResults>['alert'] | undefined
    onDisableSmartSearch: () => void
}

const alertContent: {
    [key in SmartSearchAlertKind]: (queryCount: number) => { title: JSX.Element; subtitle: JSX.Element }
} = {
    'smart-search-additional-results': (queryCount: number) => ({
        title: (
            <>
                <b>Smart Search</b> is also showing <b>additional results</b>.
            </>
        ),
        subtitle: (
            <>
                Smart Search added results for the following similar {pluralize('query', queryCount, 'queries')} that
                might interest you:
            </>
        ),
    }),
    'smart-search-pure-results': (queryCount: number) => ({
        title: (
            <>
                <b>Smart Search</b> is showing <b>related results</b> as your query found <b>no results</b>.
            </>
        ),
        subtitle: (
            <>
                To get additional results, Smart Search also ran {pluralize('this', queryCount, 'these')}{' '}
                {pluralize('query', queryCount, 'queries')}:
            </>
        ),
    }),
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

    if (
        !alert?.kind ||
        (alert.kind !== 'smart-search-additional-results' && alert.kind !== 'smart-search-pure-results')
    ) {
        return null
    }

    const content = alertContent[alert.kind](alert.proposedQueries?.length || 0)

    return (
        <div className={styles.root}>
            <Collapse isOpen={!isCollapsed} onOpenChange={opened => setIsCollapsed(!opened)}>
                <CollapseHeader className={styles.collapseButton}>
                    <div className={styles.header}>
                        <span className="d-flex align-items-baseline">
                            <Icon aria-hidden={true} svgPath={smartSearchIconSvgPath} className={styles.smartIcon} />
                            <span>
                                <H2 className={styles.title}>{content.title} </H2>
                                <span className="text-muted d-inline-block">
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
                        <span className="d-flex align-items-center flex-shrink-0 ml-2">
                            {isCollapsed ? (
                                <>
                                    <span className="text-muted mr-2 flex-shrink-0">Show queries</span>
                                    <Icon aria-label="Expand" svgPath={mdiChevronDown} />
                                </>
                            ) : (
                                <>
                                    <span className="text-muted mr-2 flex-shrink-0">Hide queries</span>
                                    <Icon aria-label="Collapse" svgPath={mdiChevronUp} />
                                </>
                            )}
                        </span>
                    </div>
                </CollapseHeader>
                <CollapsePanel>
                    <Text className={styles.description}>{content.subtitle}</Text>
                    <ul className={styles.container}>
                        {alert?.proposedQueries?.map(entry => (
                            <SmartSearchListItem proposedQuery={entry} key={entry.query} />
                        ))}
                    </ul>
                </CollapsePanel>
            </Collapse>
        </div>
    )
}

export const smartSearchEvent = (alertKind: AlertKind, alertTitle: string, descriptions: string[]): string[] => {
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

    const prefix = alertKind === 'smart-search-pure-results' ? 'SearchResultsAutoPure' : 'SearchResultsAutoAdded'

    const events = []
    for (const rule of rules) {
        events.push(`${prefix}${rule}`)
    }
    return events
}
