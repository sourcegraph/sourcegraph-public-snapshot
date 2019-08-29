import { throwError, Unsubscribable, of, NEVER } from 'rxjs'
import { TestScheduler } from 'rxjs/testing'
import * as sourcegraph from 'sourcegraph'
import { createDiagnosticService } from './diagnosticService'
import { DiagnosticSeverity, Range } from '@sourcegraph/extension-api-classes'

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

const DIAGNOSTIC_1: sourcegraph.Diagnostic = {
    resource: new URL('https://example.com/1'),
    range: new Range(1, 2, 3, 4),
    message: 'm1',
    severity: DiagnosticSeverity.Information,
}

const DIAGNOSTIC_2: sourcegraph.Diagnostic = {
    resource: new URL('https://example.com/2'),
    range: new Range(5, 6, 7, 8),
    message: 'm2',
    severity: DiagnosticSeverity.Error,
}

const SCOPE: Parameters<sourcegraph.DiagnosticProvider['provideDiagnostics']>[0] = {}

describe('DiagnosticService', () => {
    test('no providers yields empty array', () =>
        scheduler().run(({ expectObservable }) =>
            expectObservable(createDiagnosticService(false).observeDiagnostics(SCOPE)).toBe('a', {
                a: [],
            })
        ))

    test('single provider', () => {
        scheduler().run(({ cold, expectObservable }) => {
            const service = createDiagnosticService(false)
            service.registerDiagnosticProvider('', {
                provideDiagnostics: () =>
                    cold<sourcegraph.Diagnostic[]>('abcd', {
                        a: [],
                        b: [DIAGNOSTIC_1],
                        c: [DIAGNOSTIC_1, DIAGNOSTIC_2],
                        d: [],
                    }),
            })
            expectObservable(service.observeDiagnostics(SCOPE)).toBe('abcd', {
                a: [],
                b: [DIAGNOSTIC_1],
                c: [DIAGNOSTIC_1, DIAGNOSTIC_2],
                d: [],
            })
        })
    })

    test('merges results from multiple providers', () => {
        scheduler().run(({ cold, expectObservable }) => {
            const service = createDiagnosticService(false)
            const unsub1 = service.registerDiagnosticProvider('1', {
                provideDiagnostics: () => of([DIAGNOSTIC_1]),
            })
            let unsub2: Unsubscribable
            cold('-bc', {
                b: () => {
                    unsub2 = service.registerDiagnosticProvider('2', {
                        provideDiagnostics: () => of([DIAGNOSTIC_2]),
                    })
                },
                c: () => {
                    unsub1.unsubscribe()
                    unsub2.unsubscribe()
                },
            }).subscribe(f => f())
            expectObservable(service.observeDiagnostics(SCOPE)).toBe('ab(cd)', {
                a: [DIAGNOSTIC_1],
                b: [DIAGNOSTIC_1, DIAGNOSTIC_2],
                c: [DIAGNOSTIC_2],
                d: [],
            })
        })
    })

    test('suppresses errors', () => {
        scheduler().run(({ expectObservable }) => {
            const service = createDiagnosticService(false)
            service.registerDiagnosticProvider('a', {
                provideDiagnostics: () => throwError(new Error('x')),
            })
            expectObservable(service.observeDiagnostics(SCOPE)).toBe('a', {
                a: [],
            })
        })
    })

    test('enforces unique registration types', () => {
        const service = createDiagnosticService(false)
        service.registerDiagnosticProvider('a', {
            provideDiagnostics: () => NEVER,
        })
        expect(() =>
            service.registerDiagnosticProvider('a', {
                provideDiagnostics: () => NEVER,
            })
        ).toThrowError(/already registered/)
    })
})
