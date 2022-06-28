import * as util from 'util'

import * as React from 'react'

import { render } from '@testing-library/react'

import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'

import { AuthenticatedUser } from '../auth'

import { withActivation } from './withActivation'

const Component: React.FunctionComponent<
    React.PropsWithChildren<ActivationProps & { authenticatedUser: AuthenticatedUser | null }>
> = (props: ActivationProps & { authenticatedUser: AuthenticatedUser | null }) => (
    <div>activation steps: {props.activation ? util.inspect(props.activation) : 'undefined'}</div>
)

describe.skip('withActivation', () => {
    const ComponentWithActivation = withActivation(Component)

    test('no user', () => {
        const component = render(<ComponentWithActivation authenticatedUser={null} />)
        expect(component.asFragment()).toMatchSnapshot()
    })
    test('user, admin', () => {
        const mockUser = {
            id: 'userID',
            username: 'username',
            email: 'user@me.com',
            siteAdmin: true,
        }
        const component = render(<ComponentWithActivation authenticatedUser={mockUser as AuthenticatedUser} />)
        expect(component.asFragment()).toMatchSnapshot()
    })
    test('user, non-admin', () => {
        const mockUser = {
            id: 'userID',
            username: 'username',
            email: 'user@me.com',
            siteAdmin: false,
        }
        const component = render(<ComponentWithActivation authenticatedUser={mockUser as AuthenticatedUser} />)
        expect(component.asFragment()).toMatchSnapshot()
    })
})
