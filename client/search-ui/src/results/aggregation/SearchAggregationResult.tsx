import { FC, HTMLAttributes } from 'react'

import { mdiArrowCollapse } from '@mdi/js'

import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { Button, H2, Icon } from '@sourcegraph/wildcard'

import { AggregationCardMode, AggregationChartCard } from './AggregationChartCard'
import { AggregationModeControls } from './AggregationModeControls'
import { getAggregationData, useAggregationSearchMode, useAggregationUIMode, useSearchAggregationData } from './hooks'
import { AggregationUIMode } from './types'

import styles from './SearchAggregationResult.module.scss'

interface SearchAggregationResultProps extends HTMLAttributes<HTMLElement> {
    query: string
    patternType: SearchPatternType
}

export const SearchAggregationResult: FC<SearchAggregationResultProps> = props => {
    const { query, patternType, ...attributes } = props

    const [, setAggregationUIMode] = useAggregationUIMode()
    const [aggregationMode, setAggregationMode] = useAggregationSearchMode()

    const { data, error, loading } = useSearchAggregationData({ query, patternType, aggregationMode })

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

            <hr className="mt-2 mb-3" />

            <div className={styles.controls}>
                <AggregationModeControls
                    mode={aggregationMode}
                    availability={data?.searchQueryAggregate?.modeAvailability}
                    onModeChange={setAggregationMode}
                />
            </div>

            {loading ? (
                <AggregationChartCard type={AggregationCardMode.Loading} className={styles.chartContainer} />
            ) : error ? (
                <AggregationChartCard
                    type={AggregationCardMode.Error}
                    errorMessage={error.message}
                    className={styles.chartContainer}
                />
            ) : (
                <AggregationChartCard
                    mode={aggregationMode}
                    type={AggregationCardMode.Data}
                    data={getAggregationData(data)}
                    className={styles.chartContainer}
                />
            )}

            {data && (
                <ul className={styles.listResult}>
                    {getAggregationData(data).map(datum => (
                        <li key={datum.label} className={styles.listResultItem}>
                            <span>{datum.label}</span>
                            <span>{datum.count}</span>
                        </li>
                    ))}
                </ul>
            )}
        </section>
    )
}
