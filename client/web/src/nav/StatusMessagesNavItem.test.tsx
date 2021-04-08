import React from 'react'
import { of, Observable } from 'rxjs'
import { StatusMessagesNavItem } from './StatusMessagesNavItem'
import { createMemoryHistory } from 'history'
import { mount } from 'enzyme'
import { StatusMessagesResult, StatusMessageFields } from '../graphql-operations'

describe('StatusMessagesNavItem', () => {
    test('no messages', () => {
        const fetchMessages = (): Observable<StatusMessagesResult['statusMessages']> => of([])
        expect(
            mount(
                <StatusMessagesNavItem
                    fetchMessages={fetchMessages}
                    user={{ isSiteAdmin: false, id: '1', createdAt: '1970-01-01T00:00:00', username: 'alice' }}
                    history={createMemoryHistory()}
                />
            )
        ).toMatchSnapshot()
    })

    describe('one CloningProgress message', () => {
        const message: StatusMessageFields = {
            type: 'CloningProgress',
            message: 'Currently cloning repositories...',
        }

        const fetchMessages = () => of([message])
        test('as non-site admin', () => {
            expect(
                mount(
                    <StatusMessagesNavItem
                        fetchMessages={fetchMessages}
                        user={{ isSiteAdmin: false, id: '1', createdAt: '1970-01-01T00:00:00', username: 'alice' }}
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
                        user={{ isSiteAdmin: false, id: '1', createdAt: '1970-01-01T00:00:00', username: 'alice' }}
                        history={createMemoryHistory()}
                    />
                )
            ).toMatchSnapshot()
        })
    })

    describe('one ExternalServiceSyncError message', () => {
        const message: StatusMessageFields = {
            type: 'ExternalServiceSyncError',
            message: 'failed to list organization kubernetes repos: request returned status 404: Not Found',
            externalService: {
                id: 'abcd',
                displayName: 'GitHub.com',
            },
        }

        const fetchMessages = () => of([message])
        test('as non-site admin', () => {
            expect(
                mount(
                    <StatusMessagesNavItem
                        fetchMessages={fetchMessages}
                        user={{ isSiteAdmin: false, id: '1', createdAt: '1970-01-01T00:00:00', username: 'alice' }}
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
                        user={{ isSiteAdmin: false, id: '1', createdAt: '1970-01-01T00:00:00', username: 'alice' }}
                        history={createMemoryHistory()}
                    />
                )
            ).toMatchSnapshot()
        })
    })

    describe('one SyncError message', () => {
        const message: StatusMessageFields = {
            type: 'SyncError',
            message: 'syncer.sync.store.upsert-repos: pg: unique constraint foobar',
        }

        const fetchMessages = () => of([message])
        test('as non-site admin', () => {
            expect(
                mount(
                    <StatusMessagesNavItem
                        fetchMessages={fetchMessages}
                        user={{ isSiteAdmin: false, id: '1', createdAt: '1970-01-01T00:00:00', username: 'alice' }}
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
                        user={{ isSiteAdmin: false, id: '1', createdAt: '1970-01-01T00:00:00', username: 'alice' }}
                        history={createMemoryHistory()}
                    />
                )
            ).toMatchSnapshot()
        })
    })
})
