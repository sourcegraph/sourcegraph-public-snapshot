import { storiesOf } from '@storybook/react'
import sinon from 'sinon'

import { Typography } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../components/WebStory'
import { mockAuthenticatedUser } from '../../testing/util'
import { ActionProps } from '../FormActionArea'

import { EmailAction } from './EmailAction'

const { add } = storiesOf('web/enterprise/code-monitoring/actions/EmailAction', module).addParameters({
    chromatic: { disableSnapshot: false },
})

const defaultProps: ActionProps = {
    action: undefined,
    setAction: sinon.fake(),
    disabled: false,
    authenticatedUser: mockAuthenticatedUser,
    monitorName: 'Example code monitor',
}

const action: ActionProps['action'] = {
    __typename: 'MonitorEmail',
    id: 'id1',
    recipients: { nodes: [{ id: 'userID' }] },
    enabled: true,
    includeResults: false,
}

add('EmailAction', () => (
    <WebStory>
        {() => (
            <>
                <Typography.H2>Action card disabled</Typography.H2>
                <EmailAction {...defaultProps} disabled={true} />

                <Typography.H2>Closed, not populated</Typography.H2>
                <EmailAction {...defaultProps} />

                <Typography.H2>Open, not populated</Typography.H2>
                <EmailAction {...defaultProps} _testStartOpen={true} />

                <Typography.H2>Closed, populated, enabled</Typography.H2>
                <EmailAction {...defaultProps} action={action} />

                <Typography.H2>Open, populated, enabled</Typography.H2>
                <EmailAction {...defaultProps} _testStartOpen={true} action={action} />

                <Typography.H2>Closed, populated, disabled</Typography.H2>
                <EmailAction {...defaultProps} action={{ ...action, enabled: false }} />

                <Typography.H2>Open, populated, disabled</Typography.H2>
                <EmailAction {...defaultProps} _testStartOpen={true} action={{ ...action, enabled: false }} />
            </>
        )}
    </WebStory>
))
