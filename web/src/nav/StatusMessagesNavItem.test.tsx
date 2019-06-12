import React from 'react'
import renderer from 'react-test-renderer'
import { setLinkComponent } from '../../../shared/src/components/Link'
import * as GQL from '../../../shared/src/graphql/schema'
import { StatusMessagesNavItem } from './StatusMessagesNavItem'

describe('StatusMessagesNavItem', () => {
    setLinkComponent(props => <a {...props} />)
    afterAll(() => setLinkComponent(() => null)) // reset global env for other tests

    test('no messages', () => {
        expect(renderer.create(<StatusMessagesNavItem messages={[]} />).toJSON()).toMatchSnapshot()
    })

    describe('one CLONING message', () => {
        const message: GQL.IStatusMessage = {
            __typename: 'StatusMessage',
            type: GQL.StatusMessageType.CLONING,
            message: 'Currently cloning repositories...',
        }
        test('as non-site admin', () => {
            expect(renderer.create(<StatusMessagesNavItem messages={[message]} />).toJSON()).toMatchSnapshot()
        })

        test('as site admin', () => {
            expect(
                renderer.create(<StatusMessagesNavItem messages={[message]} isSiteAdmin={true} />).toJSON()
            ).toMatchSnapshot()
        })
    })
})
