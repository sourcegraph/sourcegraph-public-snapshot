import { CreateCodeMonitorPage } from './CreateCodeMonitorPage'
import * as React from 'react'
import * as H from 'history'
import { AuthenticatedUser } from '../../auth'
import sinon from 'sinon'
import { mount } from 'enzyme'
import { NEVER, of } from 'rxjs'
import { mockCodeMonitor } from './testing/util'
import { CreateCodeMonitorVariables } from '../../graphql-operations'
import { act } from 'react-dom/test-utils'

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
        history,
        deleteCodeMonitor: sinon.spy((id: string) => NEVER),
        createCodeMonitor: sinon.spy((monitor: CreateCodeMonitorVariables) =>
            of({ description: mockCodeMonitor.node.description })
        ),
    }
    let clock: sinon.SinonFakeTimers

    beforeAll(() => {
        clock = sinon.useFakeTimers()
    })

    afterAll(() => {
        clock.restore()
    })

    test('createCodeMonitor is called on submit', () => {
        let component = mount(<CreateCodeMonitorPage {...props} />)
        const nameInput = component.find('.test-name-input')
        nameInput.simulate('change', { target: { value: 'Test updated' } })

        const triggerButton = component.find('.test-trigger-button')
        triggerButton.simulate('click')
        const triggerInput = component.find('.test-trigger-input')
        expect(triggerInput.length).toBe(1)
        act(() => {
            triggerInput.simulate('change', { target: { value: 'test type:diff patterntype:literal' } })
            clock.tick(600)
        })
        component = component.update()
        expect(component.find('.is-valid').length).toBe(1)
        const submitTrigger = component.find('.test-submit-trigger')
        submitTrigger.simulate('click')
        const actionButton = component.find('.test-action-button')
        actionButton.simulate('click')
        const submitAction = component.find('.test-submit-action')
        submitAction.simulate('click')
        const submitMonitor = component.find('.test-submit-monitor')
        submitMonitor.simulate('submit')
        sinon.assert.called(props.createCodeMonitor)
        props.createCodeMonitor.resetHistory()
        component.unmount()
    })

    test('createCodeMonitor is not called on submit when trigger or action is incomplete', () => {
        let component = mount(<CreateCodeMonitorPage {...props} />)
        const monitorForm = component.find('.test-monitor-form').first()
        const nameInput = component.find('.test-name-input')
        nameInput.simulate('change', { target: { value: 'Test updated' } })
        monitorForm.simulate('submit')
        // Pressing enter does not call createCodeMonitor because other fields not complete
        sinon.assert.notCalled(props.createCodeMonitor)

        const triggerButton = component.find('.test-trigger-button')
        triggerButton.simulate('click')
        const triggerInput = component.find('.test-trigger-input')
        expect(triggerInput.length).toBe(1)
        act(() => {
            triggerInput.simulate('change', { target: { value: 'test type:diff patterntype:literal' } })
            clock.tick(600)
        })
        component = component.update()
        expect(component.find('.is-valid').length).toBe(1)
        const submitTrigger = component.find('.test-submit-trigger')
        submitTrigger.simulate('click')

        monitorForm.simulate('submit')
        // Pressing enter still does not call createCodeMonitor
        sinon.assert.notCalled(props.createCodeMonitor)

        const actionButton = component.find('.test-action-button')
        actionButton.simulate('click')
        const submitAction = component.find('.test-submit-action')
        submitAction.simulate('click')

        // Pressing enter calls createCodeMonitor when all sections are complete
        monitorForm.simulate('submit')
        sinon.assert.calledOnce(props.createCodeMonitor)
        props.createCodeMonitor.resetHistory()
        component.unmount()
    })

    test('Actions area button is disabled while trigger is incomplete', () => {
        const component = mount(<CreateCodeMonitorPage {...props} />)
        const actionButton = component.find('.test-action-button')
        expect(actionButton.prop('disabled')).toBe(true)
    })
})
