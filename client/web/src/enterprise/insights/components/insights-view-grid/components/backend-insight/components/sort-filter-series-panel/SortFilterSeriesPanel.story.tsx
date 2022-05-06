import { useState } from 'react'

import { Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../../../../../components/WebStory'
import { SeriesDisplayOptions } from '../../../../../../core/types/insight/common'

import { SortFilterSeriesPanel } from './SortFilterSeriesPanel'

import styles from './SortFilterSeriesPanel.module.scss'

const defaultStory: Meta = {
    title: 'web/insights/SortFilterSeriesPanel',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
}

export default defaultStory

export const Primary: Story = () => {
    const [value, setValue] = useState<SeriesDisplayOptions>({
        sortOptions: {
            mode: 'RESULT_COUNT',
            direction: 'DESC',
        },
        limit: 20,
    })

    return (
        <div className="d-flex">
            <div className={styles.container}>
                <SortFilterSeriesPanel value={value} onChange={setValue} />
            </div>
            <pre className="p-4">{JSON.stringify(value, null, 2)}</pre>
        </div>
    )
}
