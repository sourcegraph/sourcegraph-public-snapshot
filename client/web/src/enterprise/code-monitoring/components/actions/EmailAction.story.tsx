import type { Meta, StoryFn } from '@storybook/react'
import sinon from 'sinon'

import { H2 } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../components/WebStory'
import { mockAuthenticatedUser } from '../../testing/util'
import type { ActionProps } from '../FormActionArea'

import { EmailAction } from './EmailAction'

const config: Meta = {
    title: 'web/enterprise/code-monitoring/actions/EmailAction',
    parameters: {},
}

export default config

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
window.context.emailEnabled = true

export const EmailActionStory: StoryFn = () => (
    <WebStory>
        {() => (
            <>
                <H2>Action card disabled</H2>
                <EmailAction {...defaultProps} disabled={true} />

                <H2>Closed, not populated</H2>
                <EmailAction {...defaultProps} />

                <H2>Open, not populated</H2>
                <EmailAction {...defaultProps} _testStartOpen={true} />

                <H2>Closed, populated, enabled</H2>
                <EmailAction {...defaultProps} action={action} />

                <H2>Open, populated, enabled</H2>
                <EmailAction {...defaultProps} _testStartOpen={true} action={action} />

                <H2>Closed, populated, disabled</H2>
                <EmailAction {...defaultProps} action={{ ...action, enabled: false }} />

                <H2>Open, populated, disabled</H2>
                <EmailAction {...defaultProps} _testStartOpen={true} action={{ ...action, enabled: false }} />
            </>
        )}
    </WebStory>
)

EmailActionStory.storyName = 'EmailAction'
