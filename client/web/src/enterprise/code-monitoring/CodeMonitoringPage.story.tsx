import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'
import sinon from 'sinon'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'

import { AuthenticatedUser } from '../../auth'
import { EMPTY_FEATURE_FLAGS } from '../../featureFlags/featureFlags'
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
    featureFlags: EMPTY_FEATURE_FLAGS,
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
                'https://www.figma.com/file/6WMfHdPt2ovTE1P527brwc/Code-monitor-getting-started-21161?node-id=87%3A277',
        },
    }
)

add(
    'Code monitoring getting started page - unauthenticated',
    () => (
        <EnterpriseWebStory initialEntries={['/code-monitoring/getting-started']}>
            {props => (
                <CodeMonitoringPage
                    {...props}
                    {...additionalProps}
                    showGettingStarted={true}
                    authenticatedUser={null}
                    featureFlags={new Map([['w1-signup-optimisation', true]])}
                />
            )}
        </EnterpriseWebStory>
    ),
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/6WMfHdPt2ovTE1P527brwc/Code-monitor-getting-started-21161?node-id=1%3A1650',
        },
    }
)
