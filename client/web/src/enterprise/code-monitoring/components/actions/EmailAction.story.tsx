import { storiesOf } from '@storybook/react'
import React from 'react'
import sinon from 'sinon'

import { WebStory } from '../../../../components/WebStory'
import { mockAuthenticatedUser } from '../../testing/util'

import { EmailAction, EmailActionProps } from './EmailAction'

const { add } = storiesOf('web/enterprise/code-monitoring/actions/EmailAction', module).addParameters({
    chromatic: { disableSnapshot: false },
})

const defaultProps: EmailActionProps = {
    action: undefined,
    setAction: sinon.fake(),
    disabled: false,
    authenticatedUser: mockAuthenticatedUser,
    monitorName: 'Example code monitor',
    triggerTestEmailAction: sinon.fake(),
}

const action: EmailActionProps['action'] = {
    __typename: 'MonitorEmail',
    id: 'id1',
    recipients: { nodes: [{ id: 'userID' }] },
    enabled: true,
}

add('EmailAction', () => (
    <WebStory>
        {() => (
            <>
                <h2>Action card disabled</h2>
                <EmailAction {...defaultProps} disabled={true} />

                <h2>Closed, not populated</h2>
                <EmailAction {...defaultProps} />

                <h2>Open, not populated</h2>
                <EmailAction {...defaultProps} _testStartOpen={true} />

                <h2>Closed, populated, enabled</h2>
                <EmailAction {...defaultProps} action={action} />

                <h2>Open, populated, enabled</h2>
                <EmailAction {...defaultProps} _testStartOpen={true} action={action} />

                <h2>Closed, populated, disabled</h2>
                <EmailAction {...defaultProps} action={{ ...action, enabled: false }} />

                <h2>Open, populated, disabled</h2>
                <EmailAction {...defaultProps} _testStartOpen={true} action={{ ...action, enabled: false }} />
            </>
        )}
    </WebStory>
))
