import { FC, useMemo, useState } from 'react'

import { gql, useQuery } from '@apollo/client';
import { mdiArrowExpand, mdiPlus } from '@mdi/js'
import { ParentSize } from '@visx/responsive'

import { ButtonGroup, Button, Icon, BarChart } from '@sourcegraph/wildcard'

import { LANGUAGE_USAGE_DATA, LanguageUsageDatum } from './search-aggregation-mock-data';

import styles from './SearchAggregations.module.scss'

const IS_CODE_INSIGHTS_ENABLED_QUERY = gql`
    query IsCodeInsightsEnabled {
        enterpriseLicenseHasFeature(feature: "code-insights")
    }
`

const getValue = (datum: LanguageUsageDatum): number => datum.value
const getColor = (datum: LanguageUsageDatum): string => datum.fill
const getLink = (datum: LanguageUsageDatum): string => datum.linkURL
const getName = (datum: LanguageUsageDatum): string => datum.name

const MAX_TRUNCATED_LABEL_LENGTH = 10
const getTruncatedTick = (tick: string): string => (tick.length >= MAX_TRUNCATED_LABEL_LENGTH ? `${tick.slice(0, MAX_TRUNCATED_LABEL_LENGTH)}...` : tick)
const getTruncatedTickFromTheEnd = (tick: string): string => (tick.length >= MAX_TRUNCATED_LABEL_LENGTH ? `...${tick.slice(-MAX_TRUNCATED_LABEL_LENGTH)}` : tick)

const getTruncationFormatter = (aggregationMode: AggregationModes): (tick: string) => string => {
    switch (aggregationMode) {
        // These types possible have long labels with the same pattern at the start of the string,
        // so we truncate their labels from the end
        case AggregationModes.Repository:
        case AggregationModes.FilePath:
            return getTruncatedTickFromTheEnd

        default:
            return getTruncatedTick
    }
}

enum AggregationModes {
    Repository,
    FilePath,
    Author,
    CaptureGroups,
}

interface SearchAggregationsProps {
}

export const SearchAggregations: FC<SearchAggregationsProps> = props => {
    const [aggregationMode, setAggregationMode] = useState(AggregationModes.Repository)
    const { data } = useQuery<{ enterpriseLicenseHasFeature: boolean }>(IS_CODE_INSIGHTS_ENABLED_QUERY, { fetchPolicy: 'cache-first' })

    const getTruncatedXLabel = useMemo(() => getTruncationFormatter(aggregationMode), [aggregationMode])

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
                    <BarChart
                        width={parent.width}
                        height={parent.height}
                        data={LANGUAGE_USAGE_DATA}
                        getDatumName={getName}
                        getDatumValue={getValue}
                        getDatumColor={getColor}
                        getDatumLink={getLink}
                        pixelsPerYTick={20}
                        pixelsPerXTick={20}
                        maxAngleXTick={45}
                        getTruncatedXTick={getTruncatedXLabel}
                    />
                )}
            </ParentSize>

            <footer className={styles.actions}>
                <Button variant="secondary" size="sm" outline={true} className={styles.detailsAction}>
                    <Icon aria-hidden={true} svgPath={mdiArrowExpand}/> Expand
                </Button>

                {
                    data?.enterpriseLicenseHasFeature &&
                    <Button variant="secondary" outline={true} size="sm">
                        <Icon aria-hidden={true} svgPath={mdiPlus}/> Save insight
                    </Button>
                }
            </footer>
        </article>
    )
}
