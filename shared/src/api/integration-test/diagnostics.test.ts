import { DiagnosticSeverity, Range } from '@sourcegraph/extension-api-classes'
import { first } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { integrationTestContext } from './testHelpers'
import { of, from } from 'rxjs'

const FIXTURE_DIAGNOSTIC_1: sourcegraph.Diagnostic = {
    resource: new URL('http://a'),
    range: new Range(1, 2, 3, 4),
    message: 'm',
    severity: DiagnosticSeverity.Information,
}

describe('Diagnostics (integration)', () => {
    describe('workspace.registerDiagnosticProvider', () => {
        test('provides', async () => {
            const { services, extensionAPI } = await integrationTestContext()

            const PROVIDER_TYPE = 't'
            const subscription = extensionAPI.workspace.registerDiagnosticProvider(PROVIDER_TYPE, {
                provideDiagnostics: () => of([FIXTURE_DIAGNOSTIC_1]),
            })
            await extensionAPI.internal.sync()

            expect(
                await from(services.diagnostics.observeDiagnostics({}))
                    .pipe(first())
                    .toPromise()
            ).toEqual([{ ...FIXTURE_DIAGNOSTIC_1, type: PROVIDER_TYPE }])

            subscription.unsubscribe()
        })
    })
})
