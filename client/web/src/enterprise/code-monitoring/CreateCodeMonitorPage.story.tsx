import React from 'react'
import { CreateCodeMonitorPage } from './CreateCodeMonitorPage'
import { storiesOf } from '@storybook/react'
import { WebStory } from '../../components/WebStory'
import { AuthenticatedUser } from '../../auth'

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
