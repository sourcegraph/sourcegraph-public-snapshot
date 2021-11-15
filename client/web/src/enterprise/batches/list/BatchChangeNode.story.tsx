import { boolean } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import isChromatic from 'chromatic/isChromatic'
import classNames from 'classnames'
import { subDays } from 'date-fns'
import React from 'react'

import { WebStory } from '../../../components/WebStory'

import styles from './BatchChangeListPage.module.scss'
import { BatchChangeNode } from './BatchChangeNode'
import { nodes, now } from './testData'

const { add } = storiesOf('web/batches/BatchChangeNode', module).addDecorator(story => (
    <div className={classNames(styles.batchChangeListPageGrid, 'p-3 container')}>{story()}</div>
))

for (const key of Object.keys(nodes)) {
    add(key, () => (
        <WebStory>
            {props => (
                <BatchChangeNode
                    {...props}
                    node={nodes[key]}
                    displayNamespace={boolean('Display namespace', true)}
                    now={isChromatic() ? () => subDays(now, 5) : undefined}
                />
            )}
        </WebStory>
    ))
}
