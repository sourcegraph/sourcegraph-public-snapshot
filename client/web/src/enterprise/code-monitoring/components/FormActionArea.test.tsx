import { mount } from 'enzyme'
import React from 'react'
import { act } from 'react-dom/test-utils'
import sinon from 'sinon'

import { AuthenticatedUser } from '../../../auth'

import { FormActionArea } from './FormActionArea'

describe('FormActionArea', () => {
    const authenticatedUser = {
        id: 'foobar',
        username: 'alice',
        email: 'alice@alice.com',
    } as AuthenticatedUser
    const mockActions = {
        nodes: [{ id: 'id1', recipients: { nodes: [{ id: authenticatedUser.id }] }, enabled: true }],
    }

    test('Error is shown if code monitor has empty description', () => {
        let component = mount(
            <FormActionArea
                actions={mockActions}
                actionsCompleted={true}
                setActionsCompleted={sinon.spy()}
                disabled={false}
                authenticatedUser={authenticatedUser}
                onActionsChange={sinon.spy()}
                description=""
            />
        )
        act(() => {
            const triggerButton = component.find('.test-action-button')
            triggerButton.simulate('click')
        })
        component = component.update()
        expect(component).toMatchSnapshot()
    })
})
