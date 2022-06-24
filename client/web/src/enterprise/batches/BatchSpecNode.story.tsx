import { DecoratorFn, Meta, Story } from '@storybook/react'
import classNames from 'classnames'
import { addDays } from 'date-fns'

import { WebStory } from '../../components/WebStory'

import { BatchSpecNode } from './BatchSpecNode'
import { NODES } from './testData'

import styles from './BatchSpecsPage.module.scss'

const NOW = () => addDays(new Date(), 1)

const decorator: DecoratorFn = story => <div className={classNames(styles.specsGrid, 'p-3 container')}>{story()}</div>

const config: Meta = {
    title: 'web/batches/settings/specs',
    decorators: [decorator],
}

export default config

export const BatchSpecNodeStory: Story = () => (
    <WebStory>
        {props => (
            <>
                {NODES.map(node => (
                    <BatchSpecNode {...props} key={node.id} node={node} now={NOW} />
                ))}
            </>
        )}
    </WebStory>
)

BatchSpecNodeStory.storyName = 'BatchSpecNode'
