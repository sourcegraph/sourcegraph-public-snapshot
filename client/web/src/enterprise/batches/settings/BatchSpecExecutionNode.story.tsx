import { storiesOf } from '@storybook/react'
import classNames from 'classnames'
import { addDays } from 'date-fns'
import React from 'react'

import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { BatchSpecExecutionNode } from './BatchSpecExecutionNode'
import styles from './BatchSpecExecutionsPage.module.scss'
import { NODES } from './testData'

const { add } = storiesOf('web/batches/settings/executions', module).addDecorator(story => (
    <div className={classNames(styles.executionsGrid, 'p-3 container')}>{story()}</div>
))

const NOW = () => addDays(new Date(), 1)

add('BatchSpecExecutionNode', () => (
    <EnterpriseWebStory>
        {props => (
            <>
                {NODES.map(node => (
                    <BatchSpecExecutionNode {...props} key={node.id} node={node} now={NOW} />
                ))}
            </>
        )}
    </EnterpriseWebStory>
))
