import * as React from 'react'
import * as H from 'history'
import { AuthenticatedUser } from '../../auth'
import sinon from 'sinon'
import { mount } from 'enzyme'
import { ManageCodeMonitorPage } from './ManageCodeMonitorPage'
import { mockCodeMonitor } from './testing/util'
import { of } from 'rxjs'

describe('ManageCodeMonitorPage', () => {
    const mockUser = {
        id: 'userID',
        username: 'username',
        email: 'user@me.com',
        siteAdmin: true,
    } as AuthenticatedUser

    const history = H.createMemoryHistory()
    const props = {
        history,
        location: history.location,
        authenticatedUser: mockUser,
        breadcrumbs: [{ depth: 0, breadcrumb: null }],
        setBreadcrumb: sinon.spy(),
        useBreadcrumb: sinon.spy(),
        fetchUserCodeMonitors: sinon.spy(),
        updateCodeMonitor: sinon.spy(),
        fetchCodeMonitor: sinon.spy((id: string) => of(mockCodeMonitor)),
        match: {
            params: { id: 'test-id' },
            isExact: true,
            path: history.location.pathname,
            url: 'https://sourcegraph.com',
        },
        toggleCodeMonitorEnabled: sinon.spy((id: string, enabled: boolean) => of({ id: 'test', enabled: true })),
    }
    test('Form is pre-loaded with code monitor data', () => {
        const component = mount(<ManageCodeMonitorPage {...props} />)
        expect(props.fetchCodeMonitor.calledOnce)

        const nameInput = component.find('.test-name-input')
        expect(nameInput.length).toBe(1)
        const nameValue = nameInput.getDOMNode().getAttribute('value')
        expect(nameValue).toBe('Test code monitor')
        const currentQueryValue = component.find('.test-existing-query')
        const currentActionEmailValue = component.find('.test-existing-action-email')
        expect(currentQueryValue.getDOMNode().innerHTML).toBe('test')
        expect(currentActionEmailValue.getDOMNode().innerHTML).toBe('user@me.com')
        component.unmount()
    })

    test('Updating the form executes the update request', () => {
        const component = mount(<ManageCodeMonitorPage {...props} />)
        const nameInput = component.find('.test-name-input')
        const nameValue = nameInput.getDOMNode().getAttribute('value')
        expect(nameValue).toBe('Test code monitor')
        nameInput.simulate('change', { target: { value: 'Test updated' } })
        const submitButton = component.find('.test-submit-monitor')
        submitButton.simulate('click')
        expect(props.updateCodeMonitor.calledOnce)
        expect(props.updateCodeMonitor.calledWith({ ...mockCodeMonitor, name: 'Test code monitor' }))
        component.unmount()
    })

    test('Clicking Edit in the trigger area opens the query form', () => {
        const component = mount(<ManageCodeMonitorPage {...props} />)
        let triggerInput = component.find('.test-trigger-input')
        expect(triggerInput.length).toBe(0)
        const editTrigger = component.find('.test-edit-trigger')
        editTrigger.simulate('click')
        triggerInput = component.find('.test-trigger-input')
        expect(triggerInput.length).toBe(1)
    })

    test('Clicking Edit in the action area opens the action form', () => {
        const component = mount(<ManageCodeMonitorPage {...props} />)
        let triggerInput = component.find('.test-action-form')
        expect(triggerInput.length).toBe(0)
        const editTrigger = component.find('.test-edit-action')
        editTrigger.simulate('click')
        triggerInput = component.find('.test-action-form')
        expect(triggerInput.length).toBe(1)
    })

    test('Save button is disabled when no changes have been made, enabled when changes have been made', () => {
        const component = mount(<ManageCodeMonitorPage {...props} />)
        let submitButton = component.find('.test-submit-monitor')
        expect(submitButton.prop('disabled')).toBe(true)
        const nameInput = component.find('.test-name-input')
        nameInput.simulate('change', { target: { value: 'Test code monitor updated' } })
        submitButton = component.find('.test-submit-monitor')
        expect(submitButton.prop('disabled')).toBe(false)
    })

    test('Cancelling after changes have been made shows confirmation prompt', () => {
        const component = mount(<ManageCodeMonitorPage {...props} />)
        const confirmStub = sinon.stub(window, 'confirm')
        const nameInput = component.find('.test-name-input')
        nameInput.simulate('change', { target: { value: 'Test code monitor updated' } })
        const cancelButton = component.find('.test-cancel-monitor')
        cancelButton.simulate('click')
        expect(confirmStub.calledOnce)
        confirmStub.restore()
    })

    test('Cancelling without any changes made does not show confirmation prompt', () => {
        const component = mount(<ManageCodeMonitorPage {...props} />)
        const confirmStub = sinon.stub(window, 'confirm')
        const cancelButton = component.find('.test-cancel-monitor')
        cancelButton.simulate('click')
        expect(confirmStub.notCalled)
        confirmStub.restore()
    })
})
