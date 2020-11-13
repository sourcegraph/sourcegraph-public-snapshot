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
})
