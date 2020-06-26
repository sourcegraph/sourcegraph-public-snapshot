import * as util from 'util'
import * as React from 'react'
import { ActivationProps } from '../../../shared/src/components/activation/Activation'
import * as GQL from '../../../shared/src/graphql/schema'
import { withActivation } from './withActivation'
import { mount } from 'enzyme'

const Component: React.FunctionComponent<ActivationProps & { authenticatedUser: GQL.IUser | null }> = (
    props: ActivationProps & { authenticatedUser: GQL.IUser | null }
) => <div>activation steps: {props.activation ? util.inspect(props.activation) : 'undefined'}</div>

describe.skip('withActivation', () => {
    const ComponentWithActivation = withActivation(Component)

    test('no user', () => {
        expect(mount(<ComponentWithActivation authenticatedUser={null} />).children()).toMatchSnapshot()
    })
    test('user, admin', () => {
        const mockUser = {
            id: 'userID',
            username: 'username',
            email: 'user@me.com',
            siteAdmin: true,
        }
        expect(
            mount(<ComponentWithActivation authenticatedUser={mockUser as GQL.IUser} />).children()
        ).toMatchSnapshot()
    })
    test('user, non-admin', () => {
        const mockUser = {
            id: 'userID',
            username: 'username',
            email: 'user@me.com',
            siteAdmin: false,
        }
        expect(
            mount(<ComponentWithActivation authenticatedUser={mockUser as GQL.IUser} />).children()
        ).toMatchSnapshot()
    })
})
