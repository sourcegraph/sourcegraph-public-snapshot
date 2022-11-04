import { Meta, Story } from '@storybook/react'
import sinon from 'sinon'

import { AuthenticatedUser } from '../../auth'
import { WebStory } from '../../components/WebStory'

import { CreateCodeMonitorPage } from './CreateCodeMonitorPage'

const config: Meta = {
    title: 'web/enterprise/code-monitoring/CreateCodeMonitorPage',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}

export default config

window.context.emailEnabled = true

export const CreateCodeMonitorPageStory: Story = () => (
    <WebStory>
        {props => (
            <CreateCodeMonitorPage
                {...props}
                authenticatedUser={{ id: 'foobar', username: 'alice', email: 'alice@alice.com' } as AuthenticatedUser}
                createCodeMonitor={sinon.fake()}
                isSourcegraphDotCom={false}
            />
        )}
    </WebStory>
)

CreateCodeMonitorPageStory.storyName = 'CreateCodeMonitorPage'
CreateCodeMonitorPageStory.parameters = {
    design: {
        type: 'figma',
        url:
            'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
    },
}
