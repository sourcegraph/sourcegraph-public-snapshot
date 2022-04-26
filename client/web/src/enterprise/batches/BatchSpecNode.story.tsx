import { storiesOf } from '@storybook/react'
import classNames from 'classnames'
import { addDays } from 'date-fns'

import { WebStory } from '../../components/WebStory'

import { BatchSpecNode } from './BatchSpecNode'
import { NODES } from './testData'

import styles from './BatchSpecsPage.module.scss'

const { add } = storiesOf('web/batches/settings/specs', module).addDecorator(story => (
    <div className={classNames(styles.specsGrid, 'p-3 container')}>{story()}</div>
))

const NOW = () => addDays(new Date(), 1)

add('BatchSpecNode', () => (
    <WebStory>
        {props => (
            <>
                {NODES.map(node => (
                    <BatchSpecNode {...props} key={node.id} node={node} now={NOW} />
                ))}
            </>
        )}
    </WebStory>
))
