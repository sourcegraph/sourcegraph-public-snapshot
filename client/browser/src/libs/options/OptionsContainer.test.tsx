import * as React from 'react'
import { render, RenderResult } from 'react-testing-library'
import { noop, Observable, of } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { TestScheduler } from 'rxjs/testing'
import sinon from 'sinon'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { AccessToken } from '../../browser/types'
import { OptionsContainer, OptionsContainerProps } from './OptionsContainer'

describe('OptionsContainer', () => {
    const stubs: Pick<
        OptionsContainerProps,
        | 'getAccessToken'
        | 'setAccessToken'
        | 'fetchAccessTokenIDs'
        | 'createAccessToken'
        | 'fetchCurrentUser'
        | 'ensureValidSite'
        | 'toggleFeatureFlag'
        | 'featureFlags'
    > = {
        fetchCurrentUser: () => new Observable<GQL.IUser>(),
        createAccessToken: () => new Observable<AccessToken>(),
        getAccessToken: (url: string) => new Observable<AccessToken | undefined>(),
        setAccessToken: (token: string) => undefined,
        fetchAccessTokenIDs: (url: string) => new Observable<Pick<AccessToken, 'id'>[]>(),
        ensureValidSite: (url: string) => new Observable<void>(),
        toggleFeatureFlag: noop,
        featureFlags: [],
    }

    test('checks the connection status when it mounts', () => {
        const scheduler = new TestScheduler((a, b) => expect(a).toEqual(b))

        scheduler.run(({ cold, expectObservable }) => {
            const values = { a: 'https://test.com' }

            const siteFetches = cold('a', values).pipe(
                switchMap(
                    url =>
                        new Observable<string>(observer => {
                            const ensureValidSite = (url: string) => {
                                observer.next(url)

                                return of(undefined)
                            }

                            render(
                                <OptionsContainer
                                    {...stubs}
                                    sourcegraphURL={url}
                                    ensureValidSite={ensureValidSite}
                                    setSourcegraphURL={noop}
                                />
                            )
                        })
                )
            )

            expectObservable(siteFetches).toBe('a', values)
        })
    })

    test('checks the connection status when it the url updates', () => {
        const scheduler = new TestScheduler((a, b) => expect(a).toEqual(b))

        const buildRenderer = () => {
            let rerender: RenderResult['rerender'] | undefined

            return (ui: React.ReactElement<any>) => {
                if (rerender) {
                    rerender(ui)
                } else {
                    const renderedRes = render(ui)

                    rerender = renderedRes.rerender
                }
            }
        }

        const renderOrRerender = buildRenderer()

        scheduler.run(({ cold, expectObservable }) => {
            const values = { a: 'https://test.com', b: 'https://test1.com' }

            const siteFetches = cold('ab', values).pipe(
                switchMap(
                    url =>
                        new Observable<string>(observer => {
                            const ensureValidSite = (url: string) => {
                                observer.next(url)

                                return of(undefined)
                            }

                            renderOrRerender(
                                <OptionsContainer
                                    {...stubs}
                                    sourcegraphURL={url}
                                    ensureValidSite={ensureValidSite}
                                    setSourcegraphURL={noop}
                                />
                            )
                        })
                )
            )

            expectObservable(siteFetches).toBe('ab', values)
        })
    })

    test('handles when an error is thrown checking the site connection', () => {
        const ensureValidSite = () => {
            throw new Error('no site, woops')
        }

        try {
            render(
                <OptionsContainer
                    {...stubs}
                    sourcegraphURL={'https://test.com'}
                    ensureValidSite={ensureValidSite}
                    setSourcegraphURL={noop}
                />
            )
        } catch (err) {
            throw new Error("shouldn't be hit")
        }
    })

    test('creates a token when no token exists', () => {
        const scheduler = new TestScheduler((a, b) => expect(a).toEqual(b))

        scheduler.run(({ cold, expectObservable }) => {
            const accessToken = {
                id: 'id',
                token: 'token',
            }

            const createdAccessTokens = cold('a').pipe(
                switchMap(
                    () =>
                        new Observable<{ url: string; token: AccessToken }>(observer => {
                            const getAccessToken = () => of(undefined)

                            const createAccessToken = () => of(accessToken)

                            const user = {
                                id: 'user',
                            } as GQL.IUser

                            const ensureValidSite = () => of(undefined)
                            const fetchCurrentUser = () => of(user)
                            const fetchAccessTokenIDs = () => of([])

                            const setAccessToken = (url, token) => {
                                observer.next({ url, token })
                            }

                            render(
                                <OptionsContainer
                                    {...stubs}
                                    fetchCurrentUser={fetchCurrentUser}
                                    fetchAccessTokenIDs={fetchAccessTokenIDs}
                                    ensureValidSite={ensureValidSite}
                                    getAccessToken={getAccessToken}
                                    setAccessToken={setAccessToken}
                                    createAccessToken={createAccessToken}
                                    sourcegraphURL={'https://test.com'}
                                    setSourcegraphURL={noop}
                                />
                            )
                        })
                )
            )

            expectObservable(createdAccessTokens).toBe('a', {
                a: {
                    url: 'https://test.com',
                    token: accessToken,
                },
            })
        })
    })

    test('does not create a new access token when we have a valid token', () => {
        const getAccessToken = () => of({ id: 'valid', token: 'valid' })

        const user = {
            id: 'user',
        } as GQL.IUser

        const ensureValidSite = () => of(undefined)
        const fetchCurrentUser = () => of(user)
        const fetchAccessTokenIDs = () => of([{ id: 'valid' }])

        const createAccessToken = sinon.spy()

        render(
            <OptionsContainer
                {...stubs}
                fetchCurrentUser={fetchCurrentUser}
                fetchAccessTokenIDs={fetchAccessTokenIDs}
                ensureValidSite={ensureValidSite}
                getAccessToken={getAccessToken}
                createAccessToken={createAccessToken}
                sourcegraphURL={'https://test.com'}
                setSourcegraphURL={noop}
            />
        )

        expect(createAccessToken.notCalled).toBe(true)
    })

    test('creates a new access token when the existing token is invalid', () => {
        const getAccessToken = () => of({ id: 'invalid', token: 'invalid' })

        const accessToken = {
            id: 'id',
            token: 'token',
        }

        const createAccessToken = () => of(accessToken)

        const user = {
            id: 'user',
        } as GQL.IUser

        const ensureValidSite = () => of(undefined)
        const fetchCurrentUser = () => of(user)
        const fetchAccessTokenIDs = () => of([{ id: 'other' }])

        const setAccessToken = sinon.spy()

        render(
            <OptionsContainer
                {...stubs}
                fetchCurrentUser={fetchCurrentUser}
                fetchAccessTokenIDs={fetchAccessTokenIDs}
                ensureValidSite={ensureValidSite}
                getAccessToken={getAccessToken}
                setAccessToken={setAccessToken}
                createAccessToken={createAccessToken}
                sourcegraphURL={'https://test.com'}
                setSourcegraphURL={noop}
            />
        )

        expect(setAccessToken.calledOnceWith('https://test.com', accessToken)).toBe(true)
    })
})
