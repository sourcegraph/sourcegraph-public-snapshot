import * as assert from 'assert'
import { from } from 'rxjs'
import { filter, map, switchMap, take } from 'rxjs/operators'
import { MarkupKind } from 'sourcegraph'
import * as sourcegraph from 'sourcegraph'
import { Controller } from '../client/controller'
import { Environment } from '../client/environment'
import { HoverMerged } from '../client/types/hover'
import { clientStateIsActive } from '../test/helpers'
import { createMessageTransports } from '../test/integration/helpers'
import { createExtensionHost } from './extensionHost'

const FIXTURE_ENVIRONMENT: Environment<any, any> = {
    component: {
        document: { uri: 'file:///f', languageId: 'l', version: 1, text: '' },
        selections: [],
        visibleRanges: [],
    },
    extensions: [{ id: 'x' }],
    configuration: {},
    context: {},
}

describe('Extension host (integration)', () => {
    interface TestContext {
        clientController: Controller<any, any>
        extensionHost: typeof sourcegraph
    }

    const clientIsActive = (clientController: Controller<any, any>): Promise<void> =>
        clientController.clientEntries
            .pipe(
                filter(entries => entries.length > 0),
                map(entries => entries[0]),
                switchMap(entry => from(clientStateIsActive(entry.client))),
                take(1)
            )
            .toPromise()

    const ready = async ({ clientController, extensionHost }: TestContext): Promise<void> => {
        await extensionHost.internal.sync()
        await clientIsActive(clientController)
    }

    const create = async (): Promise<TestContext & { ready: Promise<void> }> => {
        const [clientTransports, serverTransports] = createMessageTransports()

        const clientController = new Controller({
            clientOptions: () => ({ createMessageTransports: () => clientTransports }),
        })
        clientController.setEnvironment(FIXTURE_ENVIRONMENT)

        const extensionHost = await createExtensionHost(serverTransports)
        return { clientController, extensionHost, ready: ready({ clientController, extensionHost }) }
    }

    it('registers and unregisters a provider', async () => {
        const { clientController, extensionHost, ready } = await create()

        // Register the hover provider and call it.
        const unsubscribe = extensionHost.registerHoverProvider(['*'], {
            provideHover: () => ({ contents: { value: 'a', kind: MarkupKind.PlainText } }),
        })
        await ready
        assert.deepStrictEqual(
            await clientController.registries.textDocumentHover
                .getHover({
                    textDocument: { uri: 'file:///f' },
                    position: { line: 1, character: 2 },
                })
                .pipe(take(1))
                .toPromise(),
            {
                contents: [{ value: 'a', kind: MarkupKind.PlainText }],
            } as HoverMerged
        )

        // Unregister the hover provider and ensure it's removed.
        unsubscribe.unsubscribe()
        await ready
        assert.deepStrictEqual(
            await clientController.registries.textDocumentHover
                .getHover({
                    textDocument: { uri: 'file:///f' },
                    position: { line: 1, character: 2 },
                })
                .pipe(take(1))
                .toPromise(),
            null
        )
    })

    it.skip('supports multiple hover providers', async () => {
        const { clientController, extensionHost, ready } = await create()

        extensionHost.registerHoverProvider(['*'], {
            provideHover: () => ({ contents: { value: 'a', kind: MarkupKind.PlainText } }),
        })
        extensionHost.registerHoverProvider(['*'], {
            provideHover: () => ({ contents: { value: 'b', kind: MarkupKind.PlainText } }),
        })
        await ready

        const hover = await clientController.registries.textDocumentHover
            .getHover({
                textDocument: { uri: 'file:///f' },
                position: { line: 1, character: 2 },
            })
            .pipe(take(1))
            .toPromise()
        assert.deepStrictEqual(hover, {
            contents: [{ value: 'a', kind: MarkupKind.PlainText }, { value: 'b', kind: MarkupKind.PlainText }],
        } as HoverMerged)
    })
})
