import React from 'react'
import { CodeMonitoringPage } from './CodeMonitoringPage'
import { storiesOf } from '@storybook/react'
import { AuthenticatedUser } from '../../auth'

import { ListUserCodeMonitorsVariables } from '../../graphql-operations'
import { of } from 'rxjs'
import { mockCodeMonitorNodes } from './testing/util'
import sinon from 'sinon'
import { EnterpriseWebStory } from '../components/EnterpriseWebStory'
import { EMPTY_SETTINGS_CASCADE } from '../../../../shared/src/settings/settings'

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
    () => <EnterpriseWebStory>{props => <CodeMonitoringPage {...props} {...additionalProps} />}</EnterpriseWebStory>,
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
        },
    }
)

add(
    'Code monitoring list page empty state',
    () => (
        <EnterpriseWebStory>
            {props => (
                <CodeMonitoringPage
                    {...props}
                    {...additionalProps}
                    fetchUserCodeMonitors={({ id, first, after }: ListUserCodeMonitorsVariables) =>
                        of({
                            nodes: [],
                            pageInfo: {
                                endCursor: '',
                                hasNextPage: false,
                            },
                            totalCount: 0,
                        })
                    }
                />
            )}
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
