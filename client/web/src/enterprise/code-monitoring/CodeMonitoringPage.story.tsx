import { storiesOf } from '@storybook/react'
import { NEVER, of } from 'rxjs'
import sinon from 'sinon'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'

import { AuthenticatedUser } from '../../auth'
import { WebStory } from '../../components/WebStory'
import { ListCodeMonitors, ListUserCodeMonitorsVariables } from '../../graphql-operations'

import { CodeMonitoringPage } from './CodeMonitoringPage'
import { mockCodeMonitorNodes } from './testing/util'

const { add } = storiesOf('web/enterprise/code-monitoring/CodeMonitoringPage', module).addParameters({
    chromatic: {
        // Delay screenshot taking, so <CodeMonitoringPage /> is ready to show content.
        delay: 600,
        disableSnapshot: false,
    },
})

const generateMockFetchMonitors = (count: number) => ({ id, first, after }: ListUserCodeMonitorsVariables) => {
    const result: ListCodeMonitors = {
        nodes: mockCodeMonitorNodes.slice(0, count),
        pageInfo: {
            endCursor: `foo${count}`,
            hasNextPage: count > 10,
        },
        totalCount: count,
    }
    return of(result)
}

const additionalProps = {
    authenticatedUser: { id: 'foobar', username: 'alice', email: 'alice@alice.com' } as AuthenticatedUser,
    toggleCodeMonitorEnabled: sinon.fake(),
    settingsCascade: EMPTY_SETTINGS_CASCADE,
}

const additionalPropsShortList = { ...additionalProps, fetchUserCodeMonitors: generateMockFetchMonitors(3) }
const additionalPropsLongList = { ...additionalProps, fetchUserCodeMonitors: generateMockFetchMonitors(12) }
const additionalPropsEmptyList = { ...additionalProps, fetchUserCodeMonitors: generateMockFetchMonitors(0) }
const additionalPropsAlwaysLoading = { ...additionalProps, fetchUserCodeMonitors: () => NEVER }

add(
    'Code monitoring list page - less than 10 results',
    () => <WebStory>{props => <CodeMonitoringPage {...props} {...additionalPropsShortList} />}</WebStory>,
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
        },
    }
)

add(
    'Code monitoring list page - more than 10 results',
    () => <WebStory>{props => <CodeMonitoringPage {...props} {...additionalPropsLongList} />}</WebStory>,
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
        },
    }
)

add(
    'Code monitoring list page - loading',
    () => <WebStory>{props => <CodeMonitoringPage {...props} {...additionalPropsAlwaysLoading} />}</WebStory>,
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
        },
    }
)

add(
    'Code monitoring list page - empty, show getting started',
    () => <WebStory>{props => <CodeMonitoringPage {...props} {...additionalPropsEmptyList} />}</WebStory>,
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
        },
    }
)

add('Code monitoring list page - unauthenticated, show getting started', () => (
    <WebStory initialEntries={['/code-monitoring']}>
        {props => <CodeMonitoringPage {...props} {...additionalProps} authenticatedUser={null} />}
    </WebStory>
))

add(
    'Code monitoring empty list page',
    () => (
        <WebStory initialEntries={['/code-monitoring/getting-started']}>
            {props => <CodeMonitoringPage {...props} {...additionalPropsEmptyList} testForceTab="list" />}
        </WebStory>
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
    'Code monitoring empty list page - unauthenticated',
    () => (
        <WebStory initialEntries={['/code-monitoring/getting-started']}>
            {props => (
                <CodeMonitoringPage
                    {...props}
                    {...additionalPropsEmptyList}
                    authenticatedUser={null}
                    testForceTab="list"
                />
            )}
        </WebStory>
    ),
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/6WMfHdPt2ovTE1P527brwc/Code-monitor-getting-started-21161?node-id=1%3A1650',
        },
    }
)
