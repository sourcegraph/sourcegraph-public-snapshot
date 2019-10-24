import * as React from 'react'
import renderer from 'react-test-renderer'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import * as GQL from '../../../shared/src/graphql/schema'
import { withActivation } from './withActivation'

const Component: React.FunctionComponent<ActivationProps & { authenticatedUser: GQL.IUser | null }> = (
    props: ActivationProps & { authenticatedUser: GQL.IUser | null }
) => <div>activation JSON: {props.activation ? JSON.stringify(props.activation, null, 2) : 'undefined'}</div>

describe('withActivation', () => {
    const ComponentWithActivation = withActivation(Component)

    test('no user', () => {
        const component = renderer.create(<ComponentWithActivation authenticatedUser={null} />)
        expect(component.toJSON()).toMatchSnapshot()
    })
    test('user, admin', () => {
        const mockUser = {
            id: 'userID',
            username: 'username',
            email: 'user@me.com',
            siteAdmin: true,
        }
        const component = renderer.create(<ComponentWithActivation authenticatedUser={mockUser as GQL.IUser} />)
        expect(component.toJSON()).toMatchSnapshot()
    })
    test('user, non-admin', () => {
        const mockUser = {
            id: 'userID',
            username: 'username',
            email: 'user@me.com',
            siteAdmin: false,
        }
        const component = renderer.create(<ComponentWithActivation authenticatedUser={mockUser as GQL.IUser} />)
        expect(component.toJSON()).toMatchSnapshot()
    })
})
