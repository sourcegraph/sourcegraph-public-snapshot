import React from 'react'
import renderer from 'react-test-renderer'
import { of, Observable } from 'rxjs'
import * as GQL from '../../../shared/src/graphql/schema'
import { StatusMessagesNavItem } from './StatusMessagesNavItem'
import { createMemoryHistory } from 'history'

jest.mock('mdi-react/CloudAlertIcon', () => 'CloudAlertIcon')
jest.mock('mdi-react/CloudCheckIcon', () => 'CloudCheckIcon')
jest.mock('mdi-react/CloudSyncIcon', () => 'CloudSyncIcon')

describe('StatusMessagesNavItem', () => {
    test('no messages', () => {
        const fetchMessages = (): Observable<GQL.StatusMessage[]> => of([])
        expect(
            renderer
                .create(
                    <StatusMessagesNavItem
                        fetchMessages={fetchMessages}
                        isSiteAdmin={false}
                        history={createMemoryHistory()}
                    />
                )
                .toJSON()
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
                    .create(
                        <StatusMessagesNavItem
                            fetchMessages={fetchMessages}
                            isSiteAdmin={false}
                            history={createMemoryHistory()}
                        />
                    )
                    .toJSON()
            ).toMatchSnapshot()
        })

        test('as site admin', () => {
            expect(
                renderer
                    .create(
                        <StatusMessagesNavItem
                            fetchMessages={fetchMessages}
                            isSiteAdmin={true}
                            history={createMemoryHistory()}
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
                createdAt: new Date().toISOString(),
                updatedAt: new Date().toISOString(),
                webhookURL: null,
                warning: '',
            },
        }

        const fetchMessages = () => of([message])
        test('as non-site admin', () => {
            expect(
                renderer
                    .create(
                        <StatusMessagesNavItem
                            fetchMessages={fetchMessages}
                            isSiteAdmin={false}
                            history={createMemoryHistory()}
                        />
                    )
                    .toJSON()
            ).toMatchSnapshot()
        })

        test('as site admin', () => {
            expect(
                renderer
                    .create(
                        <StatusMessagesNavItem
                            fetchMessages={fetchMessages}
                            isSiteAdmin={true}
                            history={createMemoryHistory()}
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
                    .create(
                        <StatusMessagesNavItem
                            fetchMessages={fetchMessages}
                            isSiteAdmin={false}
                            history={createMemoryHistory()}
                        />
                    )
                    .toJSON()
            ).toMatchSnapshot()
        })

        test('as site admin', () => {
            expect(
                renderer
                    .create(
                        <StatusMessagesNavItem
                            fetchMessages={fetchMessages}
                            isSiteAdmin={true}
                            history={createMemoryHistory()}
                        />
                    )
                    .toJSON()
            ).toMatchSnapshot()
        })
    })
})
