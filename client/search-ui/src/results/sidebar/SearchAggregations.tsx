import { FC } from 'react'

import { mdiArrowExpand } from '@mdi/js'

import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { Button, Icon } from '@sourcegraph/wildcard'

import {
    AggregationModeControls,
    AggregationUIMode,
    useAggregationSearchMode,
    useAggregationUIMode,
    AggregationCardMode,
    AggregationChartCard,
    useSearchAggregationData,
    getAggregationData,
    getOtherGroupCount,
} from '../aggregation'

import styles from './SearchAggregations.module.scss'

interface SearchAggregationsProps {
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

export const SearchAggregations: FC<SearchAggregationsProps> = props => {
    const { query, patternType, onQuerySubmit } = props

    const [, setAggregationUIMode] = useAggregationUIMode()
    const [aggregationMode, setAggregationMode] = useAggregationSearchMode()
    const { data, error, loading } = useSearchAggregationData({ query, patternType, aggregationMode, limit: 10 })

    return (
        <article className="pt-2">
            <AggregationModeControls
                size="sm"
                className="mb-3"
                mode={aggregationMode}
                availability={data?.searchQueryAggregate?.modeAvailability}
                onModeChange={setAggregationMode}
            />

            {loading ? (
                <AggregationChartCard
                    aria-label="Sidebar search aggregation chart"
                    type={AggregationCardMode.Loading}
                    className={styles.chartContainer}
                />
            ) : error ? (
                <AggregationChartCard
                    aria-label="Sidebar search aggregation chart"
                    type={AggregationCardMode.Error}
                    errorMessage={error.message}
                    className={styles.chartContainer}
                />
            ) : (
                <AggregationChartCard
                    aria-label="Sidebar search aggregation chart"
                    mode={aggregationMode}
                    type={AggregationCardMode.Data}
                    data={getAggregationData(data)}
                    missingCount={getOtherGroupCount(data)}
                    className={styles.chartContainer}
                    onBarLinkClick={onQuerySubmit}
                />
            )}

            <footer className={styles.actions}>
                <Button
                    variant="secondary"
                    size="sm"
                    outline={true}
                    className={styles.detailsAction}
                    data-testid="expand-aggregation-ui"
                    onClick={() => setAggregationUIMode(AggregationUIMode.SearchPage)}
                >
                    <Icon aria-hidden={true} svgPath={mdiArrowExpand} /> Expand
                </Button>
            </footer>
        </article>
    )
}
