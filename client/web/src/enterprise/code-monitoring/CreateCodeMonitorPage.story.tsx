import type { Meta, StoryFn } from '@storybook/react'
import sinon from 'sinon'

import type { AuthenticatedUser } from '../../auth'
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

export const CreateCodeMonitorPageStory: StoryFn = () => (
    <WebStory>
        {props => (
            <CreateCodeMonitorPage
                {...props}
                authenticatedUser={
                    {
                        id: 'foobar',
                        username: 'alice',
                        emails: [{ email: 'alice@alice.com', isPrimary: true, verified: true }],
                    } as AuthenticatedUser
                }
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
        url: 'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
    },
}
