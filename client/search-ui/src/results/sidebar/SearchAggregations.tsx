import { FC } from 'react'

import { mdiArrowExpand, mdiPlus } from '@mdi/js'
import { ParentSize } from '@visx/responsive'

import { ButtonGroup, Button, Icon } from '@sourcegraph/wildcard'

import { AggregationChart, AggregationModes, useAggregationSearchMode } from '../aggregation'

import { LANGUAGE_USAGE_DATA, LanguageUsageDatum } from './search-aggregation-mock-data'

import styles from './SearchAggregations.module.scss'

const getValue = (datum: LanguageUsageDatum): number => datum.value
const getColor = (datum: LanguageUsageDatum): string => datum.fill
const getLink = (datum: LanguageUsageDatum): string => datum.linkURL
const getName = (datum: LanguageUsageDatum): string => datum.name

interface SearchAggregationsProps {}

export const SearchAggregations: FC<SearchAggregationsProps> = props => {
    const [aggregationMode, setAggregationMode] = useAggregationSearchMode()

    return (
        <article className="pt-2">
            <ButtonGroup className="mb-3">
                <Button
                    className={styles.aggregationTypeControl}
                    variant="secondary"
                    size="sm"
                    outline={aggregationMode !== AggregationModes.Repository}
                    onClick={() => setAggregationMode(AggregationModes.Repository)}
                >
                    Repo
                </Button>

                <Button
                    className={styles.aggregationTypeControl}
                    variant="secondary"
                    size="sm"
                    outline={aggregationMode !== AggregationModes.FilePath}
                    onClick={() => setAggregationMode(AggregationModes.FilePath)}
                >
                    File
                </Button>

                <Button
                    className={styles.aggregationTypeControl}
                    variant="secondary"
                    size="sm"
                    outline={aggregationMode !== AggregationModes.Author}
                    onClick={() => setAggregationMode(AggregationModes.Author)}
                >
                    Author
                </Button>
                <Button
                    className={styles.aggregationTypeControl}
                    variant="secondary"
                    size="sm"
                    outline={aggregationMode !== AggregationModes.CaptureGroups}
                    onClick={() => setAggregationMode(AggregationModes.CaptureGroups)}
                >
                    Capture group
                </Button>
            </ButtonGroup>

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
                <Button variant="secondary" size="sm" outline={true} className={styles.detailsAction}>
                    <Icon aria-hidden={true} svgPath={mdiArrowExpand} /> Expand
                </Button>

                <Button variant="secondary" outline={true} size="sm">
                    <Icon aria-hidden={true} svgPath={mdiPlus} /> Save insight
                </Button>
            </footer>
        </article>
    )
}
