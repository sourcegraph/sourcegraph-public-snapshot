import { FC, useState } from 'react'

import { mdiArrowExpand, mdiPlus } from '@mdi/js'
import { ParentSize } from '@visx/responsive'

import { ButtonGroup, Button, Icon, BarChart } from '@sourcegraph/wildcard'

import styles from './SearchAggregations.module.scss'

interface LanguageUsageDatum {
    name: string
    value: number
    fill: string
    linkURL: string
    group?: string
}

const getValue = (datum: LanguageUsageDatum): number => datum.value
const getColor = (datum: LanguageUsageDatum): string => datum.fill
const getLink = (datum: LanguageUsageDatum): string => datum.linkURL
const getName = (datum: LanguageUsageDatum): string => datum.name

// Mock data for bar chart, will be removed and replace with
// actual data in https://github.com/sourcegraph/sourcegraph/issues/39956
const LANGUAGE_USAGE_DATA: LanguageUsageDatum[] = [
    {
        name: 'Julia',
        value: 1000,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/JavaScript',
    },
    {
        name: 'Erlang',
        value: 700,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/JavaScript',
    },
    {
        name: 'SQL',
        value: 550,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/JavaScript',
    },
    {
        name: 'Cobol',
        value: 500,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/JavaScript',
    },
    {
        name: 'JavaScript',
        value: 422,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/JavaScript',
    },
    {
        name: 'CSS',
        value: 273,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/CSS',
    },
    {
        name: 'HTML',
        value: 129,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/HTML',
    },
    {
        name: 'С++',
        value: 110,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
    {
        name: 'TypeScript',
        value: 95,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
    {
        name: 'Elm',
        value: 84,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
    {
        name: 'Rust',
        value: 60,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
    {
        name: 'Go',
        value: 45,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
    {
        name: 'Markdown',
        value: 35,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
    {
        name: 'Zig',
        value: 20,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
    {
        name: 'XML',
        value: 5,
        fill: 'var(--primary)',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
]

enum AggregationModes {
    Repository,
    FilePath,
    Author,
    CaptureGroups,
}

interface SearchAggregationsProps {}

export const SearchAggregations: FC<SearchAggregationsProps> = props => {
    const [aggregationMode, setAggregationMode] = useState(AggregationModes.Repository)

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
