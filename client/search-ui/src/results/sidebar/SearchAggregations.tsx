import { FC } from 'react'

import { mdiArrowExpand } from '@mdi/js'

import { SearchAggregationMode, SearchPatternType } from '@sourcegraph/shared/src/schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Icon } from '@sourcegraph/wildcard'

import {
    AggregationModeControls,
    AggregationUIMode,
    useAggregationSearchMode,
    useAggregationUIMode,
    AggregationChartCard,
    useSearchAggregationData,
} from '../aggregation'

import styles from './SearchAggregations.module.scss'

interface SearchAggregationsProps extends TelemetryProps {
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
    const { query, patternType, telemetryService, onQuerySubmit } = props

    const [, setAggregationUIMode] = useAggregationUIMode()
    const [aggregationMode, setAggregationMode] = useAggregationSearchMode()
    const { data, error, loading } = useSearchAggregationData({ query, patternType, aggregationMode, limit: 10 })

    const handleBarLinkClick = (query: string): void => {
        onQuerySubmit(query)
        telemetryService.log(
            'GroupResultsChartBarClick',
            { aggregationMode, uiMode: 'sidebar' },
            { aggregationMode, uiMode: 'sidebar' }
        )
    }

    const handleBarHover = (): void => {
        telemetryService.log(
            'GroupResultsChartBarHover',
            { aggregationMode, uiMode: 'sidebar' },
            { aggregationMode, uiMode: 'sidebar' }
        )
    }

    const handleExpandClick = (): void => {
        setAggregationUIMode(AggregationUIMode.SearchPage)
        telemetryService.log('GroupResultsExpandViewOpen', { aggregationMode }, { aggregationMode })
    }

    const handleAggregationModeChange = (mode: SearchAggregationMode): void => {
        setAggregationMode(mode)
        telemetryService.log(`GroupResults${aggregationMode}`, { uiMode: 'sidebar' }, { uiMode: 'sidebar' })
    }

    return (
        <article className="pt-2">
            <AggregationModeControls
                size="sm"
                className="mb-3"
                mode={aggregationMode}
                availability={data?.searchQueryAggregate?.modeAvailability}
                telemetryService={telemetryService}
                onModeChange={handleAggregationModeChange}
            />

            <AggregationChartCard
                aria-label="Sidebar search aggregation chart"
                data={data?.searchQueryAggregate?.aggregations}
                loading={loading}
                error={error}
                mode={aggregationMode}
                className={styles.chartContainer}
                onBarLinkClick={handleBarLinkClick}
                onBarHover={handleBarHover}
            />

            <footer className={styles.actions}>
                <Button
                    variant="secondary"
                    size="sm"
                    outline={true}
                    className={styles.detailsAction}
                    data-testid="expand-aggregation-ui"
                    onClick={handleExpandClick}
                >
                    <Icon aria-hidden={true} svgPath={mdiArrowExpand} /> Expand
                </Button>
            </footer>
        </article>
    )
}
