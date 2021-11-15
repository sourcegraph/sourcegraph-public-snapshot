import { storiesOf } from '@storybook/react'
import { cloneDeep } from 'lodash'
import React from 'react'
import { NEVER, of } from 'rxjs'
import sinon from 'sinon'

import { WebStory } from '../../components/WebStory'

import { ManageCodeMonitorPage } from './ManageCodeMonitorPage'
import { mockCodeMonitor, mockUser } from './testing/util'

const { add } = storiesOf('web/enterprise/code-monitoring/ManageCodeMonitorPage', module).addParameters({
    design: {
        type: 'figma',
        url:
            'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
    },
})

add('Example', () => (
    <WebStory>
        {props => (
            <ManageCodeMonitorPage
                {...props}
                authenticatedUser={{ ...mockUser, id: 'foobar', username: 'alice', email: 'alice@alice.com' }}
                updateCodeMonitor={sinon.fake()}
                fetchCodeMonitor={sinon.fake((id: string) => of(mockCodeMonitor))}
                deleteCodeMonitor={sinon.fake((id: string) => NEVER)}
            />
        )}
    </WebStory>
))

add('Disabled toggles', () => {
    const monitor = cloneDeep(mockCodeMonitor) // Deep clone so we can manipulate this object
    monitor.node.enabled = false
    monitor.node.actions.enabled = false
    monitor.node.actions.nodes[0].enabled = false

    return (
        <WebStory>
            {props => (
                <ManageCodeMonitorPage
                    {...props}
                    authenticatedUser={{ ...mockUser, id: 'foobar', username: 'alice', email: 'alice@alice.com' }}
                    updateCodeMonitor={sinon.fake()}
                    fetchCodeMonitor={sinon.fake((id: string) => of(monitor))}
                    deleteCodeMonitor={sinon.fake((id: string) => NEVER)}
                />
            )}
        </WebStory>
    )
})
