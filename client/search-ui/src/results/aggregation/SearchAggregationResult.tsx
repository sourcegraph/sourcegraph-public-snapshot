import { FC, HTMLAttributes } from 'react'

import { mdiArrowCollapse, mdiPlus } from '@mdi/js'
import { ParentSize } from '@visx/responsive'

import { Button, H2, Icon } from '@sourcegraph/wildcard'

import { AggregationChart } from './AggregationChart'
import { AggregationModeControls } from './AggregationModeControls'
import { useAggregationSearchMode, useAggregationUIMode } from './hooks'
import { LANGUAGE_USAGE_DATA, LanguageUsageDatum } from './search-aggregation-mock-data'
import { AggregationUIMode } from './types'

import styles from './SearchAggregationResult.module.scss'

const getValue = (datum: LanguageUsageDatum): number => datum.value
const getColor = (datum: LanguageUsageDatum): string => datum.fill
const getLink = (datum: LanguageUsageDatum): string => datum.linkURL
const getName = (datum: LanguageUsageDatum): string => datum.name

interface SearchAggregationResultProps extends HTMLAttributes<HTMLElement> {}

export const SearchAggregationResult: FC<SearchAggregationResultProps> = attributes => {
    const [aggregationMode, setAggregationMode] = useAggregationSearchMode()
    const [, setAggregationUIMode] = useAggregationUIMode()

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
                <AggregationModeControls mode={aggregationMode} onModeChange={setAggregationMode} />

                <Button variant="secondary" outline={true}>
                    <Icon aria-hidden={true} className="mr-1" svgPath={mdiPlus} />
                    Save insight
                </Button>
            </div>

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

            <ul className={styles.listResult}>
                {LANGUAGE_USAGE_DATA.map(datum => (
                    <li key={getName(datum)} className={styles.listResultItem}>
                        <span>{getName(datum)}</span>
                        <span>{getValue(datum)}</span>
                    </li>
                ))}
            </ul>
        </section>
    )
}
