import { CheckCompletion, CheckScope, CheckResult, MarkupKind } from '@sourcegraph/extension-api-classes'
import * as sourcegraph from 'sourcegraph'
import { of, throwError, Unsubscribable, from, NEVER } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { createCheckService, CheckID, observeChecksInformation } from '././checkService'
import { map } from 'rxjs/operators'
import { isDefined } from '../../../util/types'

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

const CHECK_INFO_1: sourcegraph.CheckInformation = {
    state: { completion: CheckCompletion.Completed, result: CheckResult.Success },
    description: { kind: MarkupKind.Markdown, value: 'd1' },
}

const CHECK_INFO_2: sourcegraph.CheckInformation = {
    state: { completion: CheckCompletion.InProgress },
    description: { kind: MarkupKind.Markdown, value: 'd2' },
}

const SCOPE = CheckScope.Global

const BLANK_CHECK_ID: CheckID = { type: '', id: '' }

describe('CheckService', () => {
    describe('observeChecks', () => {
        test('no providers yields empty array', () =>
            scheduler().run(({ expectObservable }) =>
                expectObservable(createCheckService().observeChecks(SCOPE)).toBe('a', {
                    a: [],
                })
            ))

        test('single provider', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const service = createCheckService()
                service.registerCheckProvider('t', () => ({
                    information: cold<sourcegraph.CheckInformation>('-bc-', {
                        b: CHECK_INFO_1,
                        c: CHECK_INFO_2,
                    }),
                }))
                expectObservable(
                    service.observeChecks(SCOPE).pipe(
                        map(checks =>
                            checks.map(check => {
                                delete check.provider
                                return check
                            })
                        )
                    )
                ).toBe('a---', {
                    a: [{ type: 't', id: 'DUMMY' }],
                })
            })
        })

        test('merges results from multiple providers', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const service = createCheckService()
                const unsub1 = service.registerCheckProvider('1', () => ({
                    information: of(CHECK_INFO_1),
                }))
                let unsub2: Unsubscribable
                cold('-bc', {
                    b: () => {
                        unsub2 = service.registerCheckProvider('2', () => ({
                            information: of(CHECK_INFO_2),
                        }))
                    },
                    c: () => {
                        unsub1.unsubscribe()
                        unsub2.unsubscribe()
                    },
                }).subscribe(f => f())
                expectObservable(
                    service.observeChecks(SCOPE).pipe(
                        map(checks =>
                            checks.map(check => {
                                delete check.provider
                                return check
                            })
                        )
                    )
                ).toBe('ab(cd)', {
                    a: [{ type: '1', id: 'DUMMY' }],
                    b: [{ type: '1', id: 'DUMMY' }, { type: '2', id: 'DUMMY' }],
                    c: [{ type: '2', id: 'DUMMY' }],
                    d: [],
                })
            })
        })

        test('propagates errors', () => {
            scheduler().run(({ expectObservable }) => {
                const service = createCheckService()
                service.registerCheckProvider('', () => {
                    throw new Error('x')
                })
                expectObservable(service.observeChecks(SCOPE)).toBe('a', {
                    a: [{ type: '', id: 'DUMMY', error: new Error('x') }],
                })
            })
        })
    })

    describe('observeCheck', () => {
        test('no providers yields null', () =>
            scheduler().run(({ expectObservable }) =>
                expectObservable(createCheckService().observeCheck(SCOPE, BLANK_CHECK_ID)).toBe('a', {
                    a: null,
                })
            ))

        test('single provider', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const service = createCheckService()
                service.registerCheckProvider('', () => ({
                    information: cold<sourcegraph.CheckInformation>('-bc-', {
                        b: CHECK_INFO_1,
                        c: CHECK_INFO_2,
                    }),
                }))
                expectObservable(service.observeCheck(SCOPE, BLANK_CHECK_ID).pipe(map(isDefined))).toBe('a---', {
                    a: true,
                })
            })
        })

        test('propagates errors', () => {
            scheduler().run(({ expectObservable }) => {
                const service = createCheckService()
                service.registerCheckProvider('', () => {
                    throw new Error('x')
                })
                expectObservable(service.observeCheck(SCOPE, BLANK_CHECK_ID)).toBe('#', undefined, new Error('x'))
            })
        })
    })

    describe('registerCheckProvider', () => {
        test('enforces unique registration names', () => {
            const service = createCheckService()
            service.registerCheckProvider('a', () => ({ information: NEVER }))
            expect(() => service.registerCheckProvider('a', () => ({ information: NEVER }))).toThrowError(
                /already registered/
            )
        })
    })
})

describe('observeChecksInformation', () => {
    test('no providers yields empty array', () =>
        scheduler().run(({ expectObservable }) =>
            expectObservable(from(observeChecksInformation(createCheckService(), SCOPE))).toBe('a', {
                a: [],
            })
        ))

    test('single provider', () => {
        scheduler().run(({ cold, expectObservable }) => {
            const service = createCheckService()
            service.registerCheckProvider('t', () => ({
                information: cold<sourcegraph.CheckInformation>('-bc-', {
                    b: CHECK_INFO_1,
                    c: CHECK_INFO_2,
                }),
            }))
            expectObservable(from(observeChecksInformation(service, SCOPE))).toBe('-bc-', {
                b: [{ type: 't', id: 'DUMMY', information: CHECK_INFO_1 }],
                c: [{ type: 't', id: 'DUMMY', information: CHECK_INFO_2 }],
            })
        })
    })

    test('merges results from multiple providers', () => {
        scheduler().run(({ cold, expectObservable }) => {
            const service = createCheckService()
            const unsub1 = service.registerCheckProvider('1', () => ({
                information: of(CHECK_INFO_1),
            }))
            let unsub2: Unsubscribable
            cold('-bc', {
                b: () => {
                    unsub2 = service.registerCheckProvider('2', () => ({
                        information: of(CHECK_INFO_2),
                    }))
                },
                c: () => {
                    unsub1.unsubscribe()
                    unsub2.unsubscribe()
                },
            }).subscribe(f => f())
            expectObservable(from(observeChecksInformation(service, SCOPE))).toBe('ab(cd)', {
                a: [{ type: '1', id: 'DUMMY', information: CHECK_INFO_1 }],
                b: [
                    { type: '1', id: 'DUMMY', information: CHECK_INFO_1 },
                    { type: '2', id: 'DUMMY', information: CHECK_INFO_2 },
                ],
                c: [{ type: '2', id: 'DUMMY', information: CHECK_INFO_2 }],
                d: [],
            })
        })
    })

    test('propagates errors', () => {
        scheduler().run(({ expectObservable }) => {
            const service = createCheckService()
            service.registerCheckProvider('1', () => {
                throw new Error('x1')
            })
            service.registerCheckProvider('2', () => ({
                information: throwError(new Error('x2')),
            }))
            expectObservable(from(observeChecksInformation(service, SCOPE))).toBe('a', {
                a: [
                    { type: '1', id: 'DUMMY', error: new Error('x1') },
                    { type: '2', id: 'DUMMY', error: new Error('x2') },
                ],
            })
        })
    })
})
