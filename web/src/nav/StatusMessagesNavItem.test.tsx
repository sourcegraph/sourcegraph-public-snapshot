import React from 'react'
import renderer from 'react-test-renderer'
import { setLinkComponent } from '../../../shared/src/components/Link'
import * as GQL from '../../../shared/src/graphql/schema'
import { StatusMessagesNavItem } from './StatusMessagesNavItem'

describe('StatusMessagesNavItem', () => {
    setLinkComponent((props: any) => <a {...props} />)
    afterAll(() => setLinkComponent(null as any)) // reset global env for other tests

    test('no messages', () => {
        expect(renderer.create(<StatusMessagesNavItem messages={[]} />).toJSON()).toMatchSnapshot()
    })
    test('one CLONING message', () => {
        const message: GQL.IStatusMessage = {
            __typename: 'StatusMessage',
            type: GQL.StatusMessageType.CLONING,
            message: 'Currently cloning repositories...',
        }
        expect(renderer.create(<StatusMessagesNavItem messages={[message]} />).toJSON()).toMatchSnapshot()
    })
})
