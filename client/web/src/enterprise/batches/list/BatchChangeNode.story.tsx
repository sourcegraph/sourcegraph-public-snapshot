import { boolean } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import isChromatic from 'chromatic/isChromatic'
import { subDays } from 'date-fns'
import React from 'react'

import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { BatchChangeNode } from './BatchChangeNode'
import { nodes, now } from './testData'
import styles from './BatchChangeListPage.module.scss'
import classNames from 'classnames'

const { add } = storiesOf('web/batches/BatchChangeNode', module).addDecorator(story => (
    <div className={classNames(styles.batchChangeListPageGrid, 'p-3 container web-content')}>{story()}</div>
))

for (const key of Object.keys(nodes)) {
    add(key, () => (
        <EnterpriseWebStory>
            {props => (
                <BatchChangeNode
                    {...props}
                    node={nodes[key]}
                    displayNamespace={boolean('Display namespace', true)}
                    now={isChromatic() ? () => subDays(now, 5) : undefined}
                />
            )}
        </EnterpriseWebStory>
    ))
}
