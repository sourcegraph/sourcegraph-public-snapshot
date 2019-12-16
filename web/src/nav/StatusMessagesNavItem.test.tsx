import React from 'react'
import renderer from 'react-test-renderer'
import { of, queueScheduler, Observable } from 'rxjs'
import * as GQL from '../../../shared/src/graphql/schema'
import { StatusMessagesNavItem } from './StatusMessagesNavItem'

describe('StatusMessagesNavItem', () => {
    test('no messages', () => {
        const fetchMessages = (): Observable<GQL.StatusMessage[]> => of([])
        expect(
            renderer.create(<StatusMessagesNavItem scheduler={queueScheduler} fetchMessages={fetchMessages} />).toJSON()
        ).toMatchSnapshot()
    })

    describe('one CloningProgress message', () => {
        const message: GQL.StatusMessage = {
            __typename: 'CloningProgress',
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

    describe('one ExternalServiceSyncError message', () => {
        const message: GQL.StatusMessage = {
            __typename: 'ExternalServiceSyncError',
            message: 'failed to list organization kubernetes repos: request returned status 404: Not Found',
            externalService: {
                __typename: 'ExternalService',
                id: 'abcd',
                displayName: 'GitHub.com',
                kind: GQL.ExternalServiceKind.GITHUB,
                config: '{}',
                createdAt: new Date(),
                updatedAt: new Date(),
                warning: '',
            },
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

    describe('one SyncError message', () => {
        const message: GQL.StatusMessage = {
            __typename: 'SyncError',
            message: 'syncer.sync.store.upsert-repos: pg: unique constraint foobar',
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
