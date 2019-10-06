import { Range } from '@sourcegraph/extension-api-classes'
import * as sourcegraph from 'sourcegraph'
import { ExtDiagnostics } from './diagnostics'

const FIXTURE_DIAGNOSTIC: sourcegraph.Diagnostic = {
    message: 'm',
    severity: 2 as sourcegraph.DiagnosticSeverity.Information,
    range: new Range(1, 2, 3, 4),
}

const FIXTURE_DIAGNOSTICS: sourcegraph.Diagnostic[] = [FIXTURE_DIAGNOSTIC]

const URL_1 = new URL('file:///1')

describe('ExtDiagnostics', () => {
    // TODO!(sqs): failing
    test.skip('$acceptDiagnosticData', () => {
        const extDiagnostics = new ExtDiagnostics({} as any)
        extDiagnostics.$acceptDiagnosticData([[URL_1.toString(), FIXTURE_DIAGNOSTICS]])
        expect(extDiagnostics.getDiagnostics()).toEqual([[URL_1, FIXTURE_DIAGNOSTICS]])
    })

    test('createDiagnosticCollection', () => {
        const extDiagnostics = new ExtDiagnostics({} as any)
        const c = extDiagnostics.createDiagnosticCollection('a')
        c.set([[URL_1, FIXTURE_DIAGNOSTICS]])
        expect(extDiagnostics.getDiagnostics()).toEqual([[URL_1, FIXTURE_DIAGNOSTICS]])
    })
})
