import React from 'react'
import renderer from 'react-test-renderer'
import { from } from 'rxjs'
import { setLinkComponent } from '../../../shared/src/components/Link'
import * as GQL from '../../../shared/src/graphql/schema'
import { StatusMessagesNavItem } from './StatusMessagesNavItem'

describe('StatusMessagesNavItem', () => {
    setLinkComponent(props => <a {...props} />)
    afterAll(() => setLinkComponent(() => null)) // reset global env for other tests

    test('no messages', () => {
        const fetchMessages = () => from([])
        expect(renderer.create(<StatusMessagesNavItem fetchMessages={fetchMessages} />).toJSON()).toMatchSnapshot()
    })

    describe('one CLONING message', () => {
        const message: GQL.IStatusMessage = {
            __typename: 'StatusMessage',
            type: GQL.StatusMessageType.CLONING,
            message: 'Currently cloning repositories...',
        }

        const fetchMessages = () => {
            console.log('here')
            return from([[message]])
        }
        test('as non-site admin', () => {
            expect(renderer.create(<StatusMessagesNavItem fetchMessages={fetchMessages} />).toJSON()).toMatchSnapshot()
        })

        test('as site admin', () => {
            expect(
                renderer.create(<StatusMessagesNavItem fetchMessages={fetchMessages} isSiteAdmin={true} />).toJSON()
            ).toMatchSnapshot()
        })
    })
})
