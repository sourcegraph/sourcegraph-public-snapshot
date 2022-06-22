import { Story } from '@storybook/react'
import { NEVER, of } from 'rxjs'
import sinon from 'sinon'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'

import { AuthenticatedUser } from '../../auth'
import { WebStory } from '../../components/WebStory'
import { ListCodeMonitors, ListUserCodeMonitorsVariables } from '../../graphql-operations'

import { CodeMonitoringPage } from './CodeMonitoringPage'
import { mockCodeMonitorNodes } from './testing/util'

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

const config = {
    title: 'web/enterprise/code-monitoring/CodeMonitoringPage',
    parameters: {
        chromatic: {
            // Delay screenshot taking, so <CodeMonitoringPage /> is ready to show content.
            delay: 600,
            disableSnapshot: false,
        },
    },
}

export default config

export const CodeMonitoringListPageLessThan10Results: Story = () => (
    <WebStory>{props => <CodeMonitoringPage {...props} {...additionalPropsShortList} />}</WebStory>
)

CodeMonitoringListPageLessThan10Results.storyName = 'Code monitoring list page - less than 10 results'

CodeMonitoringListPageLessThan10Results.parameters = {
    design: {
        type: 'figma',
        url:
            'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
    },
}

export const CodeMonitoringListPageMoreThan10Results: Story = () => (
    <WebStory>{props => <CodeMonitoringPage {...props} {...additionalPropsLongList} />}</WebStory>
)

CodeMonitoringListPageMoreThan10Results.storyName = 'Code monitoring list page - more than 10 results'

CodeMonitoringListPageMoreThan10Results.parameters = {
    design: {
        type: 'figma',
        url:
            'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
    },
}

export const CodeMonitoringListPageLoading: Story = () => (
    <WebStory>{props => <CodeMonitoringPage {...props} {...additionalPropsAlwaysLoading} />}</WebStory>
)

CodeMonitoringListPageLoading.storyName = 'Code monitoring list page - loading'

CodeMonitoringListPageLoading.parameters = {
    design: {
        type: 'figma',
        url:
            'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
    },
}

export const CodeMonitoringListPageEmptyShowGettingStarted: Story = () => (
    <WebStory>{props => <CodeMonitoringPage {...props} {...additionalPropsEmptyList} />}</WebStory>
)

CodeMonitoringListPageEmptyShowGettingStarted.storyName = 'Code monitoring list page - empty, show getting started'

CodeMonitoringListPageEmptyShowGettingStarted.parameters = {
    design: {
        type: 'figma',
        url:
            'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
    },
}

export const CodeMonitoringListPageUnauthenticatedShowGettingStarted: Story = () => (
    <WebStory initialEntries={['/code-monitoring']}>
        {props => <CodeMonitoringPage {...props} {...additionalProps} authenticatedUser={null} />}
    </WebStory>
)

CodeMonitoringListPageUnauthenticatedShowGettingStarted.storyName =
    'Code monitoring list page - unauthenticated, show getting started'

export const CodeMonitoringEmptyListPage: Story = () => (
    <WebStory initialEntries={['/code-monitoring/getting-started']}>
        {props => <CodeMonitoringPage {...props} {...additionalPropsEmptyList} testForceTab="list" />}
    </WebStory>
)

CodeMonitoringEmptyListPage.storyName = 'Code monitoring empty list page'

CodeMonitoringEmptyListPage.parameters = {
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/6WMfHdPt2ovTE1P527brwc/Code-monitor-getting-started-21161?node-id=87%3A277',
    },
}

export const CodeMonitoringEmptyListPageUnauthenticated: Story = () => (
    <WebStory initialEntries={['/code-monitoring/getting-started']}>
        {props => (
            <CodeMonitoringPage {...props} {...additionalPropsEmptyList} authenticatedUser={null} testForceTab="list" />
        )}
    </WebStory>
)

CodeMonitoringEmptyListPageUnauthenticated.storyName = 'Code monitoring empty list page - unauthenticated'

CodeMonitoringEmptyListPageUnauthenticated.parameters = {
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/6WMfHdPt2ovTE1P527brwc/Code-monitor-getting-started-21161?node-id=1%3A1650',
    },
}
