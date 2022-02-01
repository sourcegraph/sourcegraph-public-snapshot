import { storiesOf } from '@storybook/react'
import React from 'react'
import sinon from 'sinon'

import { AuthenticatedUser } from '../../auth'
import { WebStory } from '../../components/WebStory'

import { CreateCodeMonitorPage } from './CreateCodeMonitorPage'

const { add } = storiesOf('web/enterprise/code-monitoring/CreateCodeMonitorPage', module)

add(
    'Example',
    () => (
        <WebStory>
            {props => (
                <CreateCodeMonitorPage
                    {...props}
                    authenticatedUser={
                        { id: 'foobar', username: 'alice', email: 'alice@alice.com' } as AuthenticatedUser
                    }
                    createCodeMonitor={sinon.fake()}
                />
            )}
        </WebStory>
    ),
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
        },
    }
)
