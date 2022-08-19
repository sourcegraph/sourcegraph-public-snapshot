import { FC } from 'react'

import { mdiArrowExpand, mdiPlus } from '@mdi/js'
import { ParentSize } from '@visx/responsive'

import { Button, Icon } from '@sourcegraph/wildcard'

import {
    AggregationChart,
    AggregationModeControls,
    AggregationUIMode,
    useAggregationSearchMode,
    useAggregationUIMode,
} from '../aggregation'
import { LANGUAGE_USAGE_DATA, LanguageUsageDatum } from '../aggregation/search-aggregation-mock-data'

import styles from './SearchAggregations.module.scss'

const getValue = (datum: LanguageUsageDatum): number => datum.value
const getColor = (datum: LanguageUsageDatum): string => datum.fill
const getLink = (datum: LanguageUsageDatum): string => datum.linkURL
const getName = (datum: LanguageUsageDatum): string => datum.name

interface SearchAggregationsProps {}

export const SearchAggregations: FC<SearchAggregationsProps> = props => {
    const [aggregationMode, setAggregationMode] = useAggregationSearchMode()
    const [aggregationUIMode, setAggregationUIMode] = useAggregationUIMode()

    // Hide search aggregation side panel when we're showing the full UI mode
    if (aggregationUIMode !== AggregationUIMode.Sidebar) {
        return null
    }

    return (
        <article className="pt-2">
            <AggregationModeControls
                size="sm"
                className="mb-3"
                mode={aggregationMode}
                onModeChange={setAggregationMode}
            />

            <ParentSize className={styles.chartContainer}>
                {parent => (
                    <AggregationChart
                        mode={aggregationMode}
                        width={parent.width}
                        height={parent.height}
                        data={LANGUAGE_USAGE_DATA}
                        getDatumName={getName}
                        getDatumValue={getValue}
                        getDatumColor={getColor}
                        getDatumLink={getLink}
                    />
                )}
            </ParentSize>

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

                <Button variant="secondary" outline={true} size="sm">
                    <Icon aria-hidden={true} svgPath={mdiPlus} /> Save insight
                </Button>
            </footer>
        </article>
    )
}
