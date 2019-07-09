import { CheckCompletion, CheckScope, CheckResult, Range, MarkupKind } from '@sourcegraph/extension-api-classes'
import * as sourcegraph from 'sourcegraph'
import { of, throwError, Unsubscribable, combineLatest, from, NEVER } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import { createCheckService, CheckInformationWithID } from '././checkService'
import { map, switchMap } from 'rxjs/operators'
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

describe('CheckService', () => {
    describe('observeChecksInformation', () => {
        test('no providers yields empty array', () =>
            scheduler().run(({ expectObservable }) =>
                expectObservable(createCheckService(false).observeChecksInformation(SCOPE)).toBe('a', {
                    a: [],
                })
            ))

        test('single provider', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const service = createCheckService(false)
                service.registerCheckProvider('t', () => ({
                    information: cold<sourcegraph.CheckInformation>('-bc-', {
                        b: CHECK_INFO_1,
                        c: CHECK_INFO_2,
                    }),
                }))
                expectObservable(service.observeChecksInformation(SCOPE)).toBe('-bc-', {
                    b: [{ ...CHECK_INFO_1, type: 't', id: 'DUMMY' }],
                    c: [{ ...CHECK_INFO_2, type: 't', id: 'DUMMY' }],
                })
            })
        })

        test('merges results from multiple providers', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const service = createCheckService(false)
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
                expectObservable(service.observeChecksInformation(SCOPE)).toBe('ab(cd)', {
                    a: [{ ...CHECK_INFO_1, type: '1', id: 'DUMMY' }],
                    b: [{ ...CHECK_INFO_1, type: '1', id: 'DUMMY' }, { ...CHECK_INFO_2, type: '2', id: 'DUMMY' }],
                    c: [{ ...CHECK_INFO_2, type: '2', id: 'DUMMY' }],
                    d: [],
                })
            })
        })

        test('suppresses errors', () => {
            scheduler().run(({ expectObservable }) => {
                const service = createCheckService(false)
                service.registerCheckProvider('a', () => ({
                    information: throwError(new Error('x')),
                }))
                expectObservable(service.observeChecksInformation(SCOPE)).toBe('a', {
                    a: [],
                })
            })
        })
    })

    describe('observeCheck', () => {
        test('no providers yields null', () =>
            scheduler().run(({ expectObservable }) =>
                expectObservable(createCheckService(false).observeCheck('', SCOPE, '')).toBe('a', {
                    a: null,
                })
            ))

        test('single provider', () => {
            scheduler().run(({ cold, expectObservable }) => {
                const service = createCheckService(false)
                service.registerCheckProvider('', () => ({
                    information: cold<sourcegraph.CheckInformation>('-bc-', {
                        b: CHECK_INFO_1,
                        c: CHECK_INFO_2,
                    }),
                }))
                expectObservable(service.observeCheck('', SCOPE, '').pipe(map(isDefined))).toBe('a---', {
                    a: true,
                })
            })
        })

        test('propagates errors', () => {
            scheduler().run(({ expectObservable }) => {
                const service = createCheckService(false)
                service.registerCheckProvider('', () => {
                    throw new Error('x')
                })
                expectObservable(service.observeCheck('', SCOPE, '')).toBe('#', undefined, new Error('x'))
            })
        })
    })

    describe('registerCheckProvider', () => {
        test('enforces unique registration names', () => {
            const service = createCheckService(false)
            service.registerCheckProvider('a', () => ({ information: NEVER }))
            expect(() => service.registerCheckProvider('a', () => ({ information: NEVER }))).toThrowError(
                /already registered/
            )
        })
    })
})
