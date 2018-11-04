import * as assert from 'assert'
import { take } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { languages as sourcegraphLanguages } from 'sourcegraph'
import { Controller } from '../client/controller'
import { assertToJSON } from '../extension/types/common.test'
import { URI } from '../extension/types/uri'
import { Definition } from '../protocol/plainTypes'
import { createBarrier, integrationTestContext } from './helpers.test'

describe('LanguageFeatures (integration)', () => {
    testLocationProvider(
        'registerHoverProvider',
        extensionHost => extensionHost.languages.registerHoverProvider,
        label =>
            ({
                provideHover: (doc, pos) => ({ contents: { value: label, kind: sourcegraph.MarkupKind.PlainText } }),
            } as sourcegraph.HoverProvider),
        labels => ({
            contents: labels.map(label => ({ value: label, kind: sourcegraph.MarkupKind.PlainText })),
        }),
        run => ({ provideHover: run } as sourcegraph.HoverProvider),
        clientController =>
            clientController.registries.textDocumentHover
                .getHover({ textDocument: { uri: 'file:///f' }, position: { line: 1, character: 2 } })
                .pipe(take(1))
                .toPromise()
    )
    testLocationProvider(
        'registerDefinitionProvider',
        extensionHost => extensionHost.languages.registerDefinitionProvider,
        label =>
            ({
                provideDefinition: (doc, pos) => [{ uri: new URI(`file:///${label}`) }],
            } as sourcegraph.DefinitionProvider),
        labeledDefinitionResults,
        run => ({ provideDefinition: run } as sourcegraph.DefinitionProvider),
        clientController =>
            clientController.registries.textDocumentDefinition
                .getLocation({ textDocument: { uri: 'file:///f' }, position: { line: 1, character: 2 } })
                .pipe(take(1))
                .toPromise()
    )
    testLocationProvider(
        'registerTypeDefinitionProvider',
        extensionHost => extensionHost.languages.registerTypeDefinitionProvider,
        label =>
            ({
                provideTypeDefinition: (doc, pos) => [{ uri: new URI(`file:///${label}`) }],
            } as sourcegraph.TypeDefinitionProvider),
        labeledDefinitionResults,
        run => ({ provideTypeDefinition: run } as sourcegraph.TypeDefinitionProvider),
        clientController =>
            clientController.registries.textDocumentTypeDefinition
                .getLocation({ textDocument: { uri: 'file:///f' }, position: { line: 1, character: 2 } })
                .pipe(take(1))
                .toPromise()
    )
    testLocationProvider(
        'registerImplementationProvider',
        extensionHost => extensionHost.languages.registerImplementationProvider,
        label =>
            ({
                provideImplementation: (doc, pos) => [{ uri: new URI(`file:///${label}`) }],
            } as sourcegraph.ImplementationProvider),
        labeledDefinitionResults,
        run => ({ provideImplementation: run } as sourcegraph.ImplementationProvider),
        clientController =>
            clientController.registries.textDocumentImplementation
                .getLocation({ textDocument: { uri: 'file:///f' }, position: { line: 1, character: 2 } })
                .pipe(take(1))
                .toPromise()
    )
    testLocationProvider(
        'registerReferenceProvider',
        extensionHost => extensionHost.languages.registerReferenceProvider,
        label =>
            ({
                provideReferences: (doc, pos, context) => [{ uri: new URI(`file:///${label}`) }],
            } as sourcegraph.ReferenceProvider),
        labels => labels.map(label => ({ uri: `file:///${label}`, range: undefined })),
        run => ({ provideReferences: run } as sourcegraph.ReferenceProvider),
        clientController =>
            clientController.registries.textDocumentReferences
                .getLocation({
                    textDocument: { uri: 'file:///f' },
                    position: { line: 1, character: 2 },
                    context: { includeDeclaration: true },
                })
                .pipe(take(1))
                .toPromise()
    )
})

/**
 * Generates test cases for sourcegraph.languages.registerXyzProvider functions and their associated
 * XyzProviders, for providers that return a list of locations.
 */
function testLocationProvider<P>(
    name: keyof typeof sourcegraphLanguages,
    registerProvider: (
        extensionHost: typeof sourcegraph
    ) => (selector: sourcegraph.DocumentSelector, provider: P) => sourcegraph.Unsubscribable,
    labeledProvider: (label: string) => P,
    labeledProviderResults: (labels: string[]) => any,
    providerWithImpl: (run: (doc: sourcegraph.TextDocument, pos: sourcegraph.Position) => void) => P,
    getResult: (clientController: Controller<any, any>) => Promise<any>
): void {
    describe(`languages.${name}`, () => {
        it('registers and unregisters a single provider', async () => {
            const { clientController, extensionHost, ready } = await integrationTestContext()

            // Register the provider and call it.
            const unsubscribe = registerProvider(extensionHost)(['*'], labeledProvider('a'))
            await ready
            assert.deepStrictEqual(await getResult(clientController), labeledProviderResults(['a']))

            // Unregister the provider and ensure it's removed.
            unsubscribe.unsubscribe()
            assert.deepStrictEqual(await getResult(clientController), null)
        })

        it('supplies params to the provideHover method', async () => {
            const { clientController, extensionHost, ready } = await integrationTestContext()
            const { wait, done } = createBarrier()
            registerProvider(extensionHost)(
                ['*'],
                providerWithImpl((doc, pos) => {
                    assertToJSON(doc, { uri: 'file:///f', languageId: 'l', text: 't' })
                    assertToJSON(pos, { line: 1, character: 2 })
                    done()
                })
            )
            await ready
            await getResult(clientController)
            await wait
        })

        it('supports multiple providers', async () => {
            const { clientController, extensionHost, ready } = await integrationTestContext()

            // Register 2 providers with different results.
            registerProvider(extensionHost)(['*'], labeledProvider('a'))
            registerProvider(extensionHost)(['*'], labeledProvider('b'))
            await ready

            assert.deepStrictEqual(await getResult(clientController), labeledProviderResults(['a', 'b']))
        })
    })
}

function labeledDefinitionResults(labels: string[]): Definition {
    const results = labels.map(label => ({ uri: `file:///${label}`, range: undefined }))
    return labels.length <= 1 ? results[0] : results
}
