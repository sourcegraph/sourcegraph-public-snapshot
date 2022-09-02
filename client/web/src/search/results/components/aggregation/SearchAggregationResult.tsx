import { FC, HTMLAttributes } from 'react'

import { mdiArrowCollapse } from '@mdi/js'

import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { Button, H2, Icon, Code, Card, CardBody } from '@sourcegraph/wildcard'

import { AggregationChartCard, getAggregationData } from './AggregationChartCard'
import { AggregationLimitLabel } from './AggregationLimitLabel'
import { AggregationModeControls } from './AggregationModeControls'
import {
    isNonExhaustiveAggregationResults,
    useAggregationSearchMode,
    useAggregationUIMode,
    useSearchAggregationData,
} from './hooks'
import { AggregationUIMode } from './types'

import styles from './SearchAggregationResult.module.scss'

interface SearchAggregationResultProps extends HTMLAttributes<HTMLElement> {
    /**
     * Current submitted query, note that this query isn't a live query
     * that is synced with typed query in the search box, this query is submitted
     * see `searchQueryFromURL` state in the global query Zustand store.
     */
    query: string

    /** Current search query pattern type. */
    patternType: SearchPatternType

    /**
     * Emits whenever a user clicks one of aggregation chart segments (bars).
     * That should update the query and re-trigger search (but this should be connected
     * to this UI through its consumer)
     */
    onQuerySubmit: (newQuery: string) => void
}

export const SearchAggregationResult: FC<SearchAggregationResultProps> = props => {
    const { query, patternType, onQuerySubmit, ...attributes } = props

    const [, setAggregationUIMode] = useAggregationUIMode()
    const [aggregationMode, setAggregationMode] = useAggregationSearchMode()
    const { data, error, loading } = useSearchAggregationData({
        query,
        patternType,
        aggregationMode,
        limit: 30,
        proactive: true,
    })

    const handleCollapseClick = (): void => {
        setAggregationUIMode(AggregationUIMode.Sidebar)
    }

    return (
        <section {...attributes}>
            <header className={styles.header}>
                <H2 className="m-0">Group results by</H2>
                <Button
                    variant="secondary"
                    outline={true}
                    aria-label="Close aggregation full UI mode"
                    onClick={handleCollapseClick}
                >
                    <Icon aria-hidden={true} className="mr-1" svgPath={mdiArrowCollapse} />
                    Collapse
                </Button>
            </header>

            <span className="text-muted">
                Aggregation is based on results with no count limitation (<Code>count:all</Code>).
            </span>

            <Card as={CardBody} className={styles.card}>
                <div className={styles.controls}>
                    <AggregationModeControls
                        loading={loading}
                        mode={aggregationMode}
                        availability={data?.searchQueryAggregate?.modeAvailability}
                        onModeChange={setAggregationMode}
                    />

                    {isNonExhaustiveAggregationResults(data) && <AggregationLimitLabel size="md" />}
                </div>

                <AggregationChartCard
                    aria-label="Expanded search aggregation chart"
                    mode={aggregationMode}
                    data={data?.searchQueryAggregate?.aggregations}
                    loading={loading}
                    error={error}
                    size="md"
                    className={styles.chartContainer}
                    onBarLinkClick={onQuerySubmit}
                />

                {data && (
                    <ul className={styles.listResult}>
                        {getAggregationData(data.searchQueryAggregate.aggregations).map(datum => (
                            <li key={datum.label} className={styles.listResultItem}>
                                <span>{datum.label}</span>
                                <span>{datum.count}</span>
                            </li>
                        ))}
                    </ul>
                )}
            </Card>
        </section>
    )
}
