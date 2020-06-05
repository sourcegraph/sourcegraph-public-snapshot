import React from 'react'
import { of, Observable } from 'rxjs'
import * as GQL from '../../../shared/src/graphql/schema'
import { StatusMessagesNavItem } from './StatusMessagesNavItem'
import { createMemoryHistory } from 'history'
import { mount } from 'enzyme'

jest.mock('mdi-react/CloudAlertIcon', () => 'CloudAlertIcon')
jest.mock('mdi-react/CloudCheckIcon', () => 'CloudCheckIcon')
jest.mock('mdi-react/CloudSyncIcon', () => 'CloudSyncIcon')

describe('StatusMessagesNavItem', () => {
    test('no messages', () => {
        const fetchMessages = (): Observable<GQL.StatusMessage[]> => of([])
        expect(
            mount(
                <StatusMessagesNavItem
                    fetchMessages={fetchMessages}
                    isSiteAdmin={false}
                    history={createMemoryHistory()}
                />
            )
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
                mount(
                    <StatusMessagesNavItem
                        fetchMessages={fetchMessages}
                        isSiteAdmin={false}
                        history={createMemoryHistory()}
                    />
                )
            ).toMatchSnapshot()
        })

        test('as site admin', () => {
            expect(
                mount(
                    <StatusMessagesNavItem
                        fetchMessages={fetchMessages}
                        isSiteAdmin={true}
                        history={createMemoryHistory()}
                    />
                )
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
                mount(
                    <StatusMessagesNavItem
                        fetchMessages={fetchMessages}
                        isSiteAdmin={false}
                        history={createMemoryHistory()}
                    />
                )
            ).toMatchSnapshot()
        })

        test('as site admin', () => {
            expect(
                mount(
                    <StatusMessagesNavItem
                        fetchMessages={fetchMessages}
                        isSiteAdmin={true}
                        history={createMemoryHistory()}
                    />
                )
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
                mount(
                    <StatusMessagesNavItem
                        fetchMessages={fetchMessages}
                        isSiteAdmin={false}
                        history={createMemoryHistory()}
                    />
                )
            ).toMatchSnapshot()
        })

        test('as site admin', () => {
            expect(
                mount(
                    <StatusMessagesNavItem
                        fetchMessages={fetchMessages}
                        isSiteAdmin={true}
                        history={createMemoryHistory()}
                    />
                )
            ).toMatchSnapshot()
        })
    })
})
