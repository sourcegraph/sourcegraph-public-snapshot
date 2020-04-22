import * as React from 'react'
import { render, RenderResult } from '@testing-library/react'
import { noop, Observable, of } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { TestScheduler } from 'rxjs/testing'
import { OptionsContainer, OptionsContainerProps } from './OptionsContainer'

describe('OptionsContainer', () => {
    const stubs: Pick<
        OptionsContainerProps,
        | 'isActivated'
        | 'fetchCurrentTabStatus'
        | 'ensureValidSite'
        | 'toggleExtensionDisabled'
        | 'toggleFeatureFlag'
        | 'featureFlags'
        | 'hasPermissions'
        | 'requestPermissions'
    > = {
        isActivated: true,
        hasPermissions: () => Promise.resolve(true),
        requestPermissions: noop,
        fetchCurrentTabStatus: () => Promise.resolve(undefined),
        ensureValidSite: (url: string) => new Observable<void>(),
        toggleExtensionDisabled: (isActivated: boolean) => Promise.resolve(undefined),
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
                            const ensureValidSite = (url: string): Observable<void> => {
                                observer.next(url)

                                return of(undefined)
                            }

                            render(
                                <OptionsContainer
                                    {...stubs}
                                    sourcegraphURL={url}
                                    ensureValidSite={ensureValidSite}
                                    setSourcegraphURL={() => Promise.resolve()}
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

        const buildRenderer = (): ((ui: React.ReactElement) => void) => {
            let rerender: RenderResult['rerender'] | undefined

            return ui => {
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
                            const ensureValidSite = (url: string): Observable<void> => {
                                observer.next(url)

                                return of(undefined)
                            }

                            renderOrRerender(
                                <OptionsContainer
                                    {...stubs}
                                    sourcegraphURL={url}
                                    ensureValidSite={ensureValidSite}
                                    setSourcegraphURL={() => Promise.resolve()}
                                />
                            )
                        })
                )
            )

            expectObservable(siteFetches).toBe('ab', values)
        })
    })

    test('handles when an error is thrown checking the site connection', () => {
        const ensureValidSite = (): never => {
            throw new Error('no site, woops')
        }

        try {
            render(
                <OptionsContainer
                    {...stubs}
                    sourcegraphURL="https://test.com"
                    ensureValidSite={ensureValidSite}
                    setSourcegraphURL={() => Promise.resolve()}
                />
            )
        } catch (err) {
            throw new Error("shouldn't be hit")
        }
    })
})
