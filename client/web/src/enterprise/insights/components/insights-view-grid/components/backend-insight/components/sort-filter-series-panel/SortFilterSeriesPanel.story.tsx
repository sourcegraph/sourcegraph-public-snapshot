import { useState } from 'react'

import { Meta, Story } from '@storybook/react'

import { SeriesSortMode, SeriesSortDirection } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../../../../../../../components/WebStory'
import { SeriesDisplayOptionsInputRequired } from '../../../../../../core/types/insight/common'

import { SortFilterSeriesPanel } from './SortFilterSeriesPanel'

import styles from './SortFilterSeriesPanel.module.scss'

const defaultStory: Meta = {
    title: 'web/insights/SortFilterSeriesPanel',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
}

export default defaultStory

export const Primary: Story = () => {
    const [value, setValue] = useState<SeriesDisplayOptionsInputRequired>({
        limit: 20,
        sortOptions: {
            mode: SeriesSortMode.RESULT_COUNT,
            direction: SeriesSortDirection.DESC,
        },
    })

    return (
        <div className="d-flex">
            <div className={styles.container}>
                <SortFilterSeriesPanel
                    seriesCount={20}
                    selectedOption={value.sortOptions}
                    limit={value.limit}
                    onChange={setValue}
                />
            </div>
            <pre className="p-4">{JSON.stringify(value, null, 2)}</pre>
        </div>
    )
}
