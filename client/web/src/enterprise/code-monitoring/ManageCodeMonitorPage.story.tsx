import { storiesOf } from '@storybook/react'
import { NEVER, of } from 'rxjs'
import { fake } from 'sinon'

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
                updateCodeMonitor={fake()}
                fetchCodeMonitor={fake(() => of(mockCodeMonitor))}
                deleteCodeMonitor={fake(() => NEVER)}
                isSourcegraphDotCom={false}
            />
        )}
    </WebStory>
))
