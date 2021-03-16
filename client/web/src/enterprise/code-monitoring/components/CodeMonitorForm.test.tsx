import React from 'react'
import { mount } from 'enzyme'
import { createMemoryHistory, createLocation } from 'history'
import { CodeMonitorForm, CodeMonitorFormProps } from './CodeMonitorForm'
import { NEVER } from 'rxjs'
import { AuthenticatedUser } from '../../../auth'

const PROPS: CodeMonitorFormProps = {
    history: createMemoryHistory(),
    location: createLocation('/code-monitoring/new'),
    onSubmit: () => NEVER,
    submitButtonLabel: '',
    authenticatedUser: {
        id: 'userID',
        username: 'username',
        email: 'user@me.com',
        siteAdmin: true,
    } as AuthenticatedUser,
}

describe('CodeMonitorForm', () => {
    test('Uses trigger-query when present in URL search params', () => {
        const component = mount(
            <CodeMonitorForm {...PROPS} location={createLocation('/code-monitoring/new?trigger-query=foo')} />
        )
        const triggerQuery = component.find('[data-testid="trigger-query-edit"]')
        expect(triggerQuery.length).toStrictEqual(1)
        expect(triggerQuery.at(0).prop('value')).toStrictEqual('foo')
    })
})
