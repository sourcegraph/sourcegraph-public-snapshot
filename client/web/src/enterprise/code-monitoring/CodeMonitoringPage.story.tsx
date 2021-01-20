import React from 'react'
import { CodeMonitoringPage } from './CodeMonitoringPage'
import { storiesOf } from '@storybook/react'
import { WebStory } from '../../components/WebStory'
import { AuthenticatedUser } from '../../auth'

import { ListUserCodeMonitorsVariables } from '../../graphql-operations'
import { of } from 'rxjs'
import { mockCodeMonitorNodes } from './testing/util'

const { add } = storiesOf('web/enterprise/code-monitoring/CodeMonitoringPage', module)

const additionalProps = {
    authenticatedUser: { id: 'foobar', username: 'alice', email: 'alice@alice.com' } as AuthenticatedUser,
    fetchUserCodeMonitors: ({ id, first, after }: ListUserCodeMonitorsVariables) =>
        of({
            nodes: mockCodeMonitorNodes,
            pageInfo: {
                endCursor: 'foo10',
                hasNextPage: true,
            },
            totalCount: 12,
        }),
}

add(
    'Code monitoring list page',
    () => <WebStory>{props => <CodeMonitoringPage {...props} {...additionalProps} />}</WebStory>,
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
        },
    }
)
