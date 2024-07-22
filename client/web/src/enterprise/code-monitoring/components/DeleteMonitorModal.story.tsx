import type { Meta, StoryFn } from '@storybook/react'
import { NEVER } from 'rxjs'
import sinon from 'sinon'

import { WebStory } from '../../../components/WebStory'
import type { CodeMonitorFields } from '../../../graphql-operations'
import { mockCodeMonitor } from '../testing/util'

import { DeleteMonitorModal } from './DeleteMonitorModal'

const config: Meta = {
    title: 'web/enterprise/code-monitoring/DeleteMonitorModal',
    parameters: {},
}

export default config

export const DeleteMonitorModalStory: StoryFn = () => (
    <WebStory>
        {props => (
            <DeleteMonitorModal
                {...props}
                isOpen={true}
                codeMonitor={mockCodeMonitor.node as CodeMonitorFields}
                toggleDeleteModal={sinon.fake()}
                deleteCodeMonitor={sinon.fake(() => NEVER)}
            />
        )}
    </WebStory>
)

DeleteMonitorModalStory.storyName = 'DeleteMonitorModal'
DeleteMonitorModalStory.parameters = {
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
    },
}
