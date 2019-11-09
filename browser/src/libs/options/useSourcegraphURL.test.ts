import { TestScheduler } from 'rxjs/testing'
import * as sinon from 'sinon'
import { observeSourcegraphURLEdition } from './useSourcegraphURL'
import { NEVER, of, Subject } from 'rxjs'
import { InvalidSourcegraphURLError } from '../../shared/util/context'
import { AuthRequiredError } from '../../../../shared/src/backend/errors'

const scheduler = (): TestScheduler => new TestScheduler((a, b) => expect(a).toEqual(b))

describe('observeConnectionStatus()', () => {
    const defaultProps = {
        changes: NEVER,
        submits: NEVER,
        persistSourcegraphURL: sinon.spy((sourcegraphURL: string) => NEVER),
        urlHasPermissions: sinon.spy((url: string) => of(true)),
        connectToSourcegraphInstance: sinon.spy((url: string) => NEVER),
        observeSourcegraphURL: sinon.spy(() => NEVER),
    }
    test('Returns an error for an invalid persisted Sourcegraph URL', () => {
        scheduler().run(({ cold, expectObservable }) => {
            expectObservable(
                observeSourcegraphURLEdition({
                    ...defaultProps,
                    observeSourcegraphURL: () => cold('a', { a: 'sourcegraph.com' }),
                })
            ).toBe('a', {
                a: {
                    sourcegraphURL: 'sourcegraph.com',
                    connectionStatus: {
                        type: 'error',
                        error: new InvalidSourcegraphURLError('sourcegraph.com'),
                    },
                },
            })
        })
    })
    test('Checks the connection status of a valid persisted Sourcegraph URL (successful connection)', () => {
        scheduler().run(({ cold, expectObservable }) => {
            expectObservable(
                observeSourcegraphURLEdition({
                    ...defaultProps,
                    connectToSourcegraphInstance: () => cold('-b'),
                    observeSourcegraphURL: () => of('https://sourcegraph.com'),
                })
            ).toBe('ab', {
                a: {
                    sourcegraphURL: 'https://sourcegraph.com',
                    connectionStatus: {
                        type: 'connecting',
                    },
                },
                b: {
                    sourcegraphURL: 'https://sourcegraph.com',
                    connectionStatus: {
                        type: 'connected',
                    },
                },
            })
        })
    })
    test('Checks the connection status of a valid persisted Sourcegraph URL (error on connection)', () => {
        scheduler().run(({ cold, expectObservable }) => {
            expectObservable(
                observeSourcegraphURLEdition({
                    ...defaultProps,
                    connectToSourcegraphInstance: () =>
                        cold('-#', undefined, new AuthRequiredError('https://sourcegraph.com')),
                    observeSourcegraphURL: () => cold('a', { a: 'https://sourcegraph.com' }),
                })
            ).toBe('ab', {
                a: {
                    sourcegraphURL: 'https://sourcegraph.com',
                    connectionStatus: {
                        type: 'connecting',
                    },
                },
                b: {
                    sourcegraphURL: 'https://sourcegraph.com',
                    connectionStatus: {
                        type: 'error',
                        error: new AuthRequiredError('https://sourcegraph.com'),
                        urlHasPermissions: true,
                    },
                },
            })
        })
    })
    test('Checks the connection status of a valid persisted Sourcegraph URL (error on checking permissions)', () => {
        scheduler().run(({ cold, expectObservable }) => {
            expectObservable(
                observeSourcegraphURLEdition({
                    ...defaultProps,
                    urlHasPermissions: () => cold('#', undefined, 'Test error'),
                    connectToSourcegraphInstance: () =>
                        cold('-#', undefined, new AuthRequiredError('https://sourcegraph.com')),
                    observeSourcegraphURL: () => cold('a', { a: 'https://sourcegraph.com' }),
                })
            ).toBe('01', [
                {
                    sourcegraphURL: 'https://sourcegraph.com',
                    connectionStatus: {
                        type: 'connecting',
                    },
                },
                {
                    sourcegraphURL: 'https://sourcegraph.com',
                    connectionStatus: {
                        type: 'error',
                        error: new AuthRequiredError('https://sourcegraph.com'),
                    },
                },
            ])
        })
    })
    test('Emits changes, then perists the Sourcegraph URL on explicit submit', () => {
        scheduler().run(({ cold, expectObservable }) => {
            const submits = new Subject<string>()
            expectObservable(
                observeSourcegraphURLEdition({
                    ...defaultProps,
                    changes: cold('abc', {
                        a: 'https://sourcegraph.co',
                        b: 'https://sourcegraph.',
                        c: 'https://sourcegraph.org',
                    }),
                    submits: cold('---d'),
                    persistSourcegraphURL: (sourcegraphURL: string) => {
                        submits.next(sourcegraphURL)
                        return of(undefined)
                    },
                })
            ).toBe('abc-', {
                a: {
                    sourcegraphURL: 'https://sourcegraph.co',
                    connectionStatus: undefined,
                },
                b: {
                    sourcegraphURL: 'https://sourcegraph.',
                    connectionStatus: undefined,
                },
                c: {
                    sourcegraphURL: 'https://sourcegraph.org',
                    connectionStatus: undefined,
                },
            })
            expectObservable(submits).toBe('---a', {
                a: 'https://sourcegraph.org',
            })
        })
    })
    test('Emits changes, then persists the Sourcegraph URL after 2s', () => {
        scheduler().run(({ cold, expectObservable }) => {
            const submits = new Subject<string>()
            expectObservable(
                observeSourcegraphURLEdition({
                    ...defaultProps,
                    changes: cold('abc 2s de', {
                        a: 'https://sourcegraph.co',
                        b: 'https://sourcegraph.',
                        c: 'https://sourcegraph.org',
                        d: 'https://sourcegraph.ed',
                        e: 'https://sourcegraph.edu',
                    }),
                    persistSourcegraphURL: (sourcegraphURL: string) => {
                        submits.next(sourcegraphURL)
                        return of(undefined)
                    },
                })
            ).toBe('abc 2s de', {
                a: {
                    sourcegraphURL: 'https://sourcegraph.co',
                    connectionStatus: undefined,
                },
                b: {
                    sourcegraphURL: 'https://sourcegraph.',
                    connectionStatus: undefined,
                },
                c: {
                    sourcegraphURL: 'https://sourcegraph.org',
                    connectionStatus: undefined,
                },
                d: {
                    sourcegraphURL: 'https://sourcegraph.ed',
                    connectionStatus: undefined,
                },
                e: {
                    sourcegraphURL: 'https://sourcegraph.edu',
                    connectionStatus: undefined,
                },
            })
            expectObservable(submits).toBe('-- 2s a - 2s b', {
                a: 'https://sourcegraph.org',
                b: 'https://sourcegraph.edu',
            })
        })
    })
    test('Does not persist an invalid Sourcegraph URL', () => {
        scheduler().run(({ cold, expectObservable }) => {
            const submits = new Subject<string>()
            expectObservable(
                observeSourcegraphURLEdition({
                    ...defaultProps,
                    changes: cold('abc', {
                        a: 'https://sourcegraph.co',
                        b: 'https://sourcegraph.',
                        c: 'https://sourcegraph.org',
                    }),
                    submits: cold('---d'),
                    persistSourcegraphURL: (sourcegraphURL: string) => {
                        submits.next(sourcegraphURL)
                        return of(undefined)
                    },
                })
            ).toBe('abc-', {
                a: {
                    sourcegraphURL: 'https://sourcegraph.co',
                    connectionStatus: undefined,
                },
                b: {
                    sourcegraphURL: 'https://sourcegraph.',
                    connectionStatus: undefined,
                },
                c: {
                    sourcegraphURL: 'https://sourcegraph.org',
                    connectionStatus: undefined,
                },
            })
            expectObservable(submits).toBe('---a', {
                a: 'https://sourcegraph.org',
            })
        })
    })
    test('Emits changes, then persists the Sourcegraph URL after 2s', () => {
        scheduler().run(({ cold, expectObservable }) => {
            const submits = new Subject<string>()
            expectObservable(
                observeSourcegraphURLEdition({
                    ...defaultProps,
                    changes: cold('a', {
                        a: 'sourcegraph.com',
                    }),
                    submits: cold('-b'),
                    persistSourcegraphURL: (sourcegraphURL: string) => {
                        submits.next(sourcegraphURL)
                        return of(undefined)
                    },
                })
            ).toBe('ab', {
                a: {
                    sourcegraphURL: 'sourcegraph.com',
                    connectionStatus: undefined,
                },
                b: {
                    sourcegraphURL: 'sourcegraph.com',
                    connectionStatus: {
                        type: 'error',
                        error: new InvalidSourcegraphURLError('sourcegraph.com'),
                    },
                },
            })
            expectObservable(submits).toBe('--')
        })
    })
    test('Handles errors when persisting the Sourcegraph URL', () => {
        scheduler().run(({ cold, expectObservable }) => {
            expectObservable(
                observeSourcegraphURLEdition({
                    ...defaultProps,
                    changes: cold('a', {
                        a: 'https://sourcegraph.com',
                    }),
                    submits: cold('-b'),
                    persistSourcegraphURL: (sourcegraphURL: string) => cold('#', undefined, 'Test error'),
                })
            ).toBe('ab', {
                a: {
                    sourcegraphURL: 'https://sourcegraph.com',
                    connectionStatus: undefined,
                },
                b: {
                    sourcegraphURL: 'https://sourcegraph.com',
                    connectionStatus: {
                        type: 'error',
                        error: new Error('Error setting Sourcegraph URL'),
                    },
                },
            })
        })
    })
})
