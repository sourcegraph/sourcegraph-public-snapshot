import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import React from 'react'
import sinon from 'sinon'

import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { CodeMonitorFields } from '../../../graphql-operations'
import { mockAuthenticatedUser } from '../testing/util'

import { FormActionArea } from './FormActionArea'

describe('FormActionArea', () => {
    const mockActions: CodeMonitorFields['actions'] = {
        nodes: [
            {
                __typename: 'MonitorEmail',
                id: 'id1',
                recipients: { nodes: [{ id: mockAuthenticatedUser.id }] },
                enabled: true,
            },
        ],
    }

    test('Error is shown if code monitor has empty description', () => {
        const { asFragment } = renderWithBrandedContext(
            <FormActionArea
                actions={mockActions}
                actionsCompleted={true}
                setActionsCompleted={sinon.spy()}
                disabled={false}
                authenticatedUser={mockAuthenticatedUser}
                onActionsChange={sinon.spy()}
                monitorName=""
            />
        )

        userEvent.click(screen.getByTestId('form-action-toggle-email'))

        expect(asFragment()).toMatchSnapshot()
    })
})
