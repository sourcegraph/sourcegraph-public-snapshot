import { storiesOf } from '@storybook/react'
import { NEVER, of } from 'rxjs'
import sinon from 'sinon'

import { WebStory } from '../../components/WebStory'

import { ManageCodeMonitorPage } from './ManageCodeMonitorPage'
import { mockCodeMonitor, mockUser } from './testing/util'

const { add } = storiesOf('web/enterprise/code-monitoring/ManageCodeMonitorPage', module).addParameters({
    design: {
        type: 'figma',
        url:
            'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
    },
    chromatic: { disableSnapshot: false },
})

add('ManageCodeMonitorPage', () => (
    <WebStory>
        {props => (
            <ManageCodeMonitorPage
                {...props}
                authenticatedUser={{ ...mockUser, id: 'foobar', username: 'alice', email: 'alice@alice.com' }}
                updateCodeMonitor={sinon.fake()}
                fetchCodeMonitor={sinon.fake((id: string) => of(mockCodeMonitor))}
                deleteCodeMonitor={sinon.fake((id: string) => NEVER)}
                isSourcegraphDotCom={false}
            />
        )}
    </WebStory>
))
