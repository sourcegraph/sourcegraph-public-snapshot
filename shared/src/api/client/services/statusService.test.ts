import { StatusCompletion, StatusScope } from '@sourcegraph/extension-api-classes'
import { of, throwError, Unsubscribable } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import * as sourcegraph from 'sourcegraph'
import { createStatusService, WrappedStatus } from './statusService'

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

const STATUS_1: WrappedStatus = { name: '', status: { title: '1', state: StatusCompletion.Completed } }

const STATUS_2: WrappedStatus = { name: '', status: { title: '2', state: StatusCompletion.Completed } }

const SCOPE = StatusScope.Global

describe('StatusService', () => {
    describe('observeStatuses', () => {
        test('no providers yields empty array', () =>
            scheduler().run(({ expectObservable }) =>
                expectObservable(createStatusService(false).observeStatuses(SCOPE)).toBe('a', {
                    a: [],
                })
            ))

        test('single provider', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const service = createStatusService(false)
                service.registerStatusProvider('', {
                    provideStatus: () =>
                        cold<sourcegraph.Status | null>('abcd', {
                            a: null,
                            b: STATUS_1.status,
                            c: STATUS_2.status,
                            d: null,
                        }),
                })
                expectObservable(service.observeStatuses(SCOPE)).toBe('abcd', {
                    a: [],
                    b: [STATUS_1],
                    c: [STATUS_2],
                    d: [],
                })
            })
        })

        test('merges results from multiple providers', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const service = createStatusService(false)
                const unsub1 = service.registerStatusProvider('1', {
                    provideStatus: () => of(STATUS_1.status),
                })
                let unsub2: Unsubscribable
                cold('-bc', {
                    b: () => {
                        unsub2 = service.registerStatusProvider('2', {
                            provideStatus: () => of(STATUS_2.status),
                        })
                    },
                    c: () => {
                        unsub1.unsubscribe()
                        unsub2.unsubscribe()
                    },
                }).subscribe(f => f())
                expectObservable(service.observeStatuses(SCOPE)).toBe('ab(cd)', {
                    a: [STATUS_1],
                    b: [STATUS_1, STATUS_2],
                    c: [STATUS_2],
                    d: [],
                })
            })
        })

        test('suppresses errors', () => {
            scheduler().run(({ expectObservable }) => {
                const service = createStatusService(false)
                service.registerStatusProvider('a', {
                    provideStatus: () => throwError(new Error('x')),
                })
                expectObservable(service.observeStatuses(SCOPE)).toBe('a', {
                    a: [],
                })
            })
        })
    })

    describe('observeStatuse', () => {
        test('no providers yields null', () =>
            scheduler().run(({ expectObservable }) =>
                expectObservable(createStatusService(false).observeStatus('', SCOPE)).toBe('a', {
                    a: null,
                })
            ))

        test('single provider', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const service = createStatusService(false)
                service.registerStatusProvider('', {
                    provideStatus: () =>
                        cold<sourcegraph.Status | null>('abcd', {
                            a: null,
                            b: STATUS_1.status,
                            c: STATUS_2.status,
                            d: null,
                        }),
                })
                expectObservable(service.observeStatus('', SCOPE)).toBe('abcd', {
                    a: null,
                    b: STATUS_1,
                    c: STATUS_2,
                    d: null,
                })
            })
        })

        test('suppresses errors', () => {
            // TODO!(sqs): probably should NOT suppress errors, especially when observing a single status
            scheduler().run(({ expectObservable }) => {
                const service = createStatusService(false)
                service.registerStatusProvider('a', {
                    provideStatus: () => throwError(new Error('x')),
                })
                expectObservable(service.observeStatus('', SCOPE)).toBe('a', {
                    a: null,
                })
            })
        })
    })

    describe('registerStatusProvider', () => {
        test('enforces unique registration names', () => {
            const service = createStatusService(false)
            service.registerStatusProvider('a', {
                provideStatus: () => null,
            })
            expect(() =>
                service.registerStatusProvider('a', {
                    provideStatus: () => null,
                })
            ).toThrowError(/already registered/)
        })
    })
})
