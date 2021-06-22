import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'
import sinon from 'sinon'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'

import { AuthenticatedUser } from '../../auth'
import { ListUserCodeMonitorsVariables } from '../../graphql-operations'
import { EnterpriseWebStory } from '../components/EnterpriseWebStory'

import { CodeMonitoringPage } from './CodeMonitoringPage'
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
    toggleCodeMonitorEnabled: sinon.fake(),
    settingsCascade: EMPTY_SETTINGS_CASCADE,
}

add(
    'Code monitoring list page',
    () => (
        <EnterpriseWebStory initialEntries={['/code-monitoring']}>
            {props => <CodeMonitoringPage {...props} {...additionalProps} />}
        </EnterpriseWebStory>
    ),
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
        },
    }
)

add(
    'Code monitoring getting started page',
    () => (
        <EnterpriseWebStory initialEntries={['/code-monitoring/getting-started']}>
            {props => <CodeMonitoringPage {...props} {...additionalProps} showGettingStarted={true} />}
        </EnterpriseWebStory>
    ),
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
        },
    }
)
