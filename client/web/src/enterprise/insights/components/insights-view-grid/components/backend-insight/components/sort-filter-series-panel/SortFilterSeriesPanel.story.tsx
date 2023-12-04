import { useState } from 'react'

import type { Meta, StoryFn } from '@storybook/react'

import { SeriesSortMode, SeriesSortDirection } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../../../../../../../components/WebStory'
import type { DrillDownFiltersFormValues } from '../drill-down-filters-panel'

import { SortFilterSeriesPanel } from './SortFilterSeriesPanel'

import styles from './SortFilterSeriesPanel.module.scss'

const defaultStory: Meta = {
    title: 'web/insights/SortFilterSeriesPanel',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
}

export default defaultStory

export const Primary: StoryFn = () => {
    const [value, setValue] = useState<DrillDownFiltersFormValues['seriesDisplayOptions']>({
        limit: 20,
        numSamples: null,
        sortOptions: {
            mode: SeriesSortMode.RESULT_COUNT,
            direction: SeriesSortDirection.DESC,
        },
    })

    return (
        <div className="d-flex">
            <div className={styles.container}>
                <SortFilterSeriesPanel value={value} isNumSamplesFilterAvailable={true} onChange={setValue} />
            </div>
            <pre className="p-4">{JSON.stringify(value, null, 2)}</pre>
        </div>
    )
}
