import { MarkupKind, CheckResult, CheckCompletion, CheckScope } from '@sourcegraph/extension-api-classes'
import { take, toArray } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { integrationTestContext } from './testHelpers'
import { of, from } from 'rxjs'
import { observeChecksInformation } from '../client/services/checkService'

const CHECK_INFO_1: sourcegraph.CheckInformation = {
    description: { kind: MarkupKind.Markdown, value: 'd' },
    state: { completion: CheckCompletion.Completed, result: CheckResult.Success },
}

describe('Checks (integration)', () => {
    describe('checks.registerCheckProvider', () => {
        test('provides', async () => {
            const { services, extensionAPI } = await integrationTestContext()

            const subscription = extensionAPI.checks.registerCheckProvider(
                't',
                (): sourcegraph.CheckProvider => {
                    return {
                        information: of<sourcegraph.CheckInformation>(CHECK_INFO_1),
                    }
                }
            )
            await extensionAPI.internal.sync()

            expect(
                await from(observeChecksInformation(services.checks, CheckScope.Global))
                    .pipe(
                        take(2),
                        toArray()
                    )
                    .toPromise()
            ).toEqual([CHECK_INFO_1])

            subscription.unsubscribe()
        })
    })
})
