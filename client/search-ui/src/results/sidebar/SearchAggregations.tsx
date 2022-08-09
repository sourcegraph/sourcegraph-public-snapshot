import { FC } from 'react'

import { mdiArrowExpand, mdiPlus } from '@mdi/js'
import { ParentSize } from '@visx/responsive'

// TODO: import Chart UI from the wildcard, see https://github.com/sourcegraph/sourcegraph/issues/40136
// eslint-disable-next-line no-restricted-imports
import { BarChart } from '@sourcegraph/web/src/charts'
import { Button, Icon } from '@sourcegraph/wildcard'

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
        name: 'JavaScript',
        value: 422,
        fill: '#f1e05a',
        linkURL: 'https://en.wikipedia.org/wiki/JavaScript',
    },
    {
        name: 'CSS',
        value: 273,
        fill: '#563d7c',
        linkURL: 'https://en.wikipedia.org/wiki/CSS',
    },
    {
        name: 'HTML',
        value: 129,
        fill: '#e34c26',
        linkURL: 'https://en.wikipedia.org/wiki/HTML',
    },
    {
        name: 'Markdown',
        value: 35,
        fill: '#083fa1',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
]

interface SearchAggregationsProps {}

export const SearchAggregations: FC<SearchAggregationsProps> = props => (
    <div>
        <header className={styles.buttonGroup}>
            <Button variant="secondary" size="sm">
                Repository
            </Button>

            <Button variant="secondary" size="sm">
                Capture group
            </Button>

            <Button variant="secondary" size="sm">
                Author
            </Button>
            <Button variant="secondary" size="sm">
                File path
            </Button>
        </header>

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
                />
            )}
        </ParentSize>

        <footer className={styles.actions}>
            <Button variant="secondary" size="sm" className={styles.detailsAction}>
                <Icon aria-hidden={true} svgPath={mdiArrowExpand} /> Details
            </Button>

            <Button variant="primary" size="sm">
                <Icon aria-hidden={true} svgPath={mdiPlus} /> Create code insight
            </Button>
        </footer>
    </div>
)
