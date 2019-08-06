import React from 'react'
import renderer from 'react-test-renderer'
import { of, queueScheduler, Observable } from 'rxjs'
import { setLinkComponent } from '../../../shared/src/components/Link'
import * as GQL from '../../../shared/src/graphql/schema'
import { StatusMessagesNavItem } from './StatusMessagesNavItem'

describe('StatusMessagesNavItem', () => {
    setLinkComponent(props => <a {...props} />)
    afterAll(() => setLinkComponent(() => null)) // reset global env for other tests

    test('no messages', () => {
        const fetchMessages = (): Observable<GQL.IStatusMessage[]> => of([])
        expect(
            renderer.create(<StatusMessagesNavItem scheduler={queueScheduler} fetchMessages={fetchMessages} />).toJSON()
        ).toMatchSnapshot()
    })

    describe('one CLONING message', () => {
        const message: GQL.IStatusMessage = {
            __typename: 'StatusMessage',
            type: GQL.StatusMessageType.CLONING,
            message: 'Currently cloning repositories...',
        }

        const fetchMessages = (): Observable<GQL.IStatusMessage[]> => of([message])
        test('as non-site admin', () => {
            expect(
                renderer
                    .create(<StatusMessagesNavItem scheduler={queueScheduler} fetchMessages={fetchMessages} />)
                    .toJSON()
            ).toMatchSnapshot()
        })

        test('as site admin', () => {
            expect(
                renderer
                    .create(
                        <StatusMessagesNavItem
                            scheduler={queueScheduler}
                            fetchMessages={fetchMessages}
                            isSiteAdmin={true}
                        />
                    )
                    .toJSON()
            ).toMatchSnapshot()
        })
    })
})
