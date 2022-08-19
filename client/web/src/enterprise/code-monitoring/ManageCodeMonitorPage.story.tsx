import { Meta, Story } from '@storybook/react'
import { NEVER, of } from 'rxjs'
import { fake } from 'sinon'

import { WebStory } from '../../components/WebStory'

import { ManageCodeMonitorPage } from './ManageCodeMonitorPage'
import { mockCodeMonitor, mockUser } from './testing/util'

const config: Meta = {
    title: 'web/enterprise/code-monitoring/ManageCodeMonitorPage',
}

export default config

window.context.emailEnabled = true

export const ManageCodeMonitorPageStory: Story = () => (
    <WebStory>
        {props => (
            <ManageCodeMonitorPage
                {...props}
                authenticatedUser={{ ...mockUser, id: 'foobar', username: 'alice', email: 'alice@alice.com' }}
                updateCodeMonitor={fake()}
                fetchCodeMonitor={fake(() => of(mockCodeMonitor))}
                deleteCodeMonitor={fake(() => NEVER)}
                isSourcegraphDotCom={false}
            />
        )}
    </WebStory>
)

ManageCodeMonitorPageStory.storyName = 'ManageCodeMonitorPage'
ManageCodeMonitorPageStory.parameters = {
    design: {
        type: 'figma',
        url:
            'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
    },
    chromatic: { disableSnapshot: false },
}
