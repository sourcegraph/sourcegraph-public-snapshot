import { storiesOf } from '@storybook/react'
import classNames from 'classnames'
import { addDays } from 'date-fns'
import React from 'react'

import { WebStory } from '../../../components/WebStory'

import { BatchSpecNode } from './BatchSpecNode'
import styles from './BatchSpecsPage.module.scss'
import { NODES } from './testData'

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
