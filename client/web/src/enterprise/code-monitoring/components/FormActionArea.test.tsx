import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import React from 'react'
import sinon from 'sinon'

import { AuthenticatedUser } from '../../../auth'
import { CodeMonitorFields } from '../../../graphql-operations'

import { FormActionArea } from './FormActionArea'

describe('FormActionArea', () => {
    const authenticatedUser = {
        id: 'foobar',
        username: 'alice',
        email: 'alice@alice.com',
    } as AuthenticatedUser
    const mockActions: CodeMonitorFields['actions'] = {
        nodes: [
            {
                __typename: 'MonitorEmail',
                id: 'id1',
                recipients: { nodes: [{ id: authenticatedUser.id }] },
                enabled: true,
            },
        ],
    }

    test('Error is shown if code monitor has empty description', () => {
        const { asFragment } = render(
            <FormActionArea
                actions={mockActions}
                actionsCompleted={true}
                setActionsCompleted={sinon.spy()}
                disabled={false}
                authenticatedUser={authenticatedUser}
                onActionsChange={sinon.spy()}
                monitorName=""
            />
        )

        userEvent.click(screen.getByTestId('form-action-toggle-email-notification'))

        expect(asFragment()).toMatchSnapshot()
    })
})
