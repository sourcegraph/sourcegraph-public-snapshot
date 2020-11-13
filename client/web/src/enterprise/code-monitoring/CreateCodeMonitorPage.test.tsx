import { CreateCodeMonitorPage } from './CreateCodeMonitorPage'
import * as React from 'react'
import * as H from 'history'
import { AuthenticatedUser } from '../../auth'
import sinon from 'sinon'
import { mount } from 'enzyme'

describe('CreateCodeMonitorPage', () => {
    const mockUser = {
        id: 'userID',
        username: 'username',
        email: 'user@me.com',
        siteAdmin: true,
    } as AuthenticatedUser

    const history = H.createMemoryHistory()
    const props = {
        location: history.location,
        authenticatedUser: mockUser,
        breadcrumbs: [{ depth: 0, breadcrumb: null }],
        setBreadcrumb: sinon.spy(),
        useBreadcrumb: sinon.spy(),
    }
    test('Actions area button is disabled while trigger is incomplete', () => {
        const component = mount(<CreateCodeMonitorPage {...props} />)
        const actionButton = component.find('.test-action-button')
        expect(actionButton.prop('disabled')).toBe(true)
    })

    test('Actions area button is active when trigger is complete', () => {
        const component = mount(<CreateCodeMonitorPage {...props} />)
        const triggerButton = component.find('.test-trigger-button')
        triggerButton.simulate('click')
        const triggerInput = component.find('.test-trigger-input')
        triggerInput.simulate('change', { target: { value: 'foobar type:diff patterntype:regexp' } })
        const triggerSubmit = component.find('.test-submit-trigger')
        triggerSubmit.simulate('click')

        const actionButton = component.find('.test-action-button')
        expect(actionButton.prop('disabled')).toBe(false)
    })

    test('Can create code monitor only if trigger and action is completed', () => {
        const component = mount(<CreateCodeMonitorPage {...props} />)
        let submitMonitor = component.find('.test-submit-monitor')
        expect(submitMonitor.prop('disabled')).toBe(true)

        const triggerButton = component.find('.test-trigger-button')
        triggerButton.simulate('click')
        const triggerInput = component.find('.test-trigger-input')
        triggerInput.simulate('change', { target: { value: 'foobar type:diff patterntype:regexp' } })
        const triggerSubmit = component.find('.test-submit-trigger')
        triggerSubmit.simulate('click')

        const actionButton = component.find('.test-action-button')
        actionButton.simulate('click')
        const submitAction = component.find('.test-submit-action')
        submitAction.simulate('click')

        submitMonitor = component.find('.test-submit-monitor')
        expect(submitMonitor.prop('disabled')).toBe(false)
    })
})
