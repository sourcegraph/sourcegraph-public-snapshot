import { Range } from '@sourcegraph/extension-api-classes'
import { TestScheduler } from 'rxjs/testing'
import { Diagnostic } from '@sourcegraph/extension-api-types'
import * as sourcegraph from 'sourcegraph'
import { DiagnosticCollection } from './diagnosticCollection'
import { from, merge } from 'rxjs'
import { tap, switchMapTo } from 'rxjs/operators'

const FIXTURE_DIAGNOSTIC_1: Diagnostic = {
    message: 'm',
    severity: 2 as sourcegraph.DiagnosticSeverity.Information,
    range: new Range(1, 2, 3, 4),
}

const FIXTURE_DIAGNOSTICS: Diagnostic[] = [FIXTURE_DIAGNOSTIC_1]

const FIXTURE_DIAGNOSTIC_2: Diagnostic = {
    message: 'm2',
    severity: 3 as sourcegraph.DiagnosticSeverity.Information,
    range: new Range(5, 6, 7, 8),
}

const URL_1 = new URL('file:///1')
const URL_2 = new URL('file:///2')

const scheduler = () => new TestScheduler((a, b) => expect(a).toEqual(b))

describe('DiagnosticCollection', () => {
    test('', () => {
        const c = new DiagnosticCollection('a')
        expect(c.name).toBe('a')

        c.set(URL_1, FIXTURE_DIAGNOSTICS)
        expect(Array.from(c.entries())).toEqual([[URL_1, FIXTURE_DIAGNOSTICS]])
        expect(c.get(URL_1)).toEqual(FIXTURE_DIAGNOSTICS)
        expect(c.get(URL_2)).toBe(undefined)
        expect(c.has(URL_1)).toBe(true)
        expect(c.has(URL_2)).toBe(false)

        c.delete(URL_1)
        expect(Array.from(c.entries())).toEqual([])
        expect(c.get(URL_1)).toBe(undefined)
        expect(c.has(URL_1)).toBe(false)
    })

    test('set merges', () => {
        const c = new DiagnosticCollection('a')
        c.set(URL_1, [FIXTURE_DIAGNOSTIC_2])
        c.set([[URL_1, FIXTURE_DIAGNOSTICS], [URL_1, FIXTURE_DIAGNOSTICS]])
        expect(Array.from(c.entries())).toEqual([[URL_1, [...FIXTURE_DIAGNOSTICS, ...FIXTURE_DIAGNOSTICS]]])
    })

    test('changes', () => {
        scheduler().run(({ cold, expectObservable }) => {
            const c = new DiagnosticCollection('a')
            expectObservable(
                merge(
                    from(c.changes),
                    cold<[URL, Diagnostic[]]>('a-b', {
                        a: [URL_1, FIXTURE_DIAGNOSTICS],
                        b: [URL_2, [FIXTURE_DIAGNOSTIC_2]],
                    }).pipe(
                        tap(([uri, diagnostics]) => c.set(uri, diagnostics)),
                        switchMapTo([])
                    )
                )
            ).toBe('a-b', {
                a: [URL_1],
                b: [URL_2],
            })
        })
    })
})
