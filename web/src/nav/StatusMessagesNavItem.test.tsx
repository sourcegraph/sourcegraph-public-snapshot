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
        const fetchMessages = (): Observable<GQL.StatusMessage[]> => of([])
        expect(
            renderer.create(<StatusMessagesNavItem scheduler={queueScheduler} fetchMessages={fetchMessages} />).toJSON()
        ).toMatchSnapshot()
    })

    describe('one CLONING message', () => {
        const message: GQL.StatusMessage = {
            __typename: 'CloningStatusMessage',
            message: 'Currently cloning repositories...',
        }

        const fetchMessages = (): Observable<GQL.StatusMessage[]> => of([message])
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

    describe('one SYNCERROR message', () => {
        const message: GQL.StatusMessage = {
            __typename: 'SyncErrorStatusMessage',
            message: 'failed to list organization kubernetes repos: request returned status 404: Not Found',
            externalServiceId: '4',
            externalServiceDisplayName: 'Github Enterprise',
            externalServiceKind: 'GITHUB',
        }

        const fetchMessages = () => of([message])
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
