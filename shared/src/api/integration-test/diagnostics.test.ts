import { Range } from '@sourcegraph/extension-api-classes'
import { from } from 'rxjs'
import { distinctUntilChanged } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { collectSubscribableValues, integrationTestContext } from './testHelpers'

const FIXTURE_DIAGNOSTIC_1: sourcegraph.Diagnostic = {
    message: 'm',
    severity: 2 as sourcegraph.DiagnosticSeverity.Information,
    range: new Range(1, 2, 3, 4),
}

const URL_1 = new URL('http://a')

describe('Diagnostics (integration)', () => {
    describe('languages.createDiagnosticCollection', () => {
        test('updates diagnostics', async () => {
            const { services, extensionAPI } = await integrationTestContext()
            const values = collectSubscribableValues(
                from(services.diagnostics.collection.changes).pipe(distinctUntilChanged())
            )

            const c = extensionAPI.languages.createDiagnosticCollection('a')
            c.set(URL_1, [FIXTURE_DIAGNOSTIC_1])
            await extensionAPI.internal.sync()
            expect(values).toEqual([[URL_1]])
            expect(Array.from(services.diagnostics.collection.entries())).toEqual([[URL_1, [FIXTURE_DIAGNOSTIC_1]]])
        })
    })
})
