import { Location } from '@sourcegraph/extension-api-types'
import { Observable } from 'rxjs'
import { bufferCount, take } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { languages as sourcegraphLanguages } from 'sourcegraph'
import { Services } from '../client/services'
import { assertToJSON } from '../extension/types/testHelpers'
import { URI } from '../extension/types/uri'
import { createBarrier, integrationTestContext } from './testHelpers'

describe('LanguageFeatures (integration)', () => {
    testLocationProvider<sourcegraph.HoverProvider>(
        'registerHoverProvider',
        extensionHost => extensionHost.languages.registerHoverProvider,
        label => ({
            provideHover: (doc: sourcegraph.TextDocument, pos: sourcegraph.Position) => ({
                contents: { value: label, kind: sourcegraph.MarkupKind.PlainText },
            }),
        }),
        labels => ({
            contents: labels.map(label => ({ value: label, kind: sourcegraph.MarkupKind.PlainText })),
        }),
        run => ({ provideHover: run } as sourcegraph.HoverProvider),
        services =>
            services.textDocumentHover.getHover({
                textDocument: { uri: 'file:///f' },
                position: { line: 1, character: 2 },
            })
    )
    testLocationProvider<sourcegraph.DefinitionProvider>(
        'registerDefinitionProvider',
        extensionHost => extensionHost.languages.registerDefinitionProvider,
        label => ({
            provideDefinition: (doc: sourcegraph.TextDocument, pos: sourcegraph.Position) => [
                { uri: new URI(`file:///${label}`) },
            ],
        }),
        labeledDefinitionResults,
        run => ({ provideDefinition: run } as sourcegraph.DefinitionProvider),
        services =>
            services.textDocumentDefinition.getLocations({
                textDocument: { uri: 'file:///f' },
                position: { line: 1, character: 2 },
            })
    )
    // tslint:disable deprecation The tests must remain until they are removed.
    testLocationProvider(
        'registerTypeDefinitionProvider',
        extensionHost => extensionHost.languages.registerTypeDefinitionProvider,
        label => ({
            provideTypeDefinition: (doc: sourcegraph.TextDocument, pos: sourcegraph.Position) => [
                { uri: new URI(`file:///${label}`) },
            ],
        }),
        labeledDefinitionResults,
        run => ({ provideTypeDefinition: run } as sourcegraph.TypeDefinitionProvider),
        services =>
            services.textDocumentTypeDefinition.getLocations({
                textDocument: { uri: 'file:///f' },
                position: { line: 1, character: 2 },
            })
    )
    testLocationProvider<sourcegraph.ImplementationProvider>(
        'registerImplementationProvider',
        extensionHost => extensionHost.languages.registerImplementationProvider,
        label => ({
            provideImplementation: (doc: sourcegraph.TextDocument, pos: sourcegraph.Position) => [
                { uri: new URI(`file:///${label}`) },
            ],
        }),
        labeledDefinitionResults,
        run => ({ provideImplementation: run } as sourcegraph.ImplementationProvider),
        services =>
            services.textDocumentImplementation.getLocations({
                textDocument: { uri: 'file:///f' },
                position: { line: 1, character: 2 },
            })
    )
    // tslint:enable deprecation
    testLocationProvider<sourcegraph.ReferenceProvider>(
        'registerReferenceProvider',
        extensionHost => extensionHost.languages.registerReferenceProvider,
        label => ({
            provideReferences: (
                doc: sourcegraph.TextDocument,
                pos: sourcegraph.Position,
                context: sourcegraph.ReferenceContext
            ) => [{ uri: new URI(`file:///${label}`) }],
        }),
        labels => labels.map(label => ({ uri: `file:///${label}`, range: undefined })),
        run =>
            ({
                provideReferences: (
                    doc: sourcegraph.TextDocument,
                    pos: sourcegraph.Position,
                    _context: sourcegraph.ReferenceContext
                ) => run(doc, pos),
            } as sourcegraph.ReferenceProvider),
        services =>
            services.textDocumentReferences.getLocations({
                textDocument: { uri: 'file:///f' },
                position: { line: 1, character: 2 },
                context: { includeDeclaration: true },
            })
    )
    testLocationProvider<sourcegraph.LocationProvider>(
        'registerLocationProvider',
        extensionHost => (selector, provider) =>
            extensionHost.languages.registerLocationProvider('x', selector, provider),
        label => ({
            provideLocations: (doc: sourcegraph.TextDocument, pos: sourcegraph.Position) => [
                { uri: new URI(`file:///${label}`) },
            ],
        }),
        labels => labels.map(label => ({ uri: `file:///${label}`, range: undefined })),
        run =>
            ({
                provideLocations: (doc: sourcegraph.TextDocument, pos: sourcegraph.Position) => run(doc, pos),
            } as sourcegraph.LocationProvider),
        services =>
            services.textDocumentLocations.getLocations('x', {
                textDocument: { uri: 'file:///f' },
                position: { line: 1, character: 2 },
            })
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
    getResult: (services: Services) => Observable<any>
): void {
    describe(`languages.${name}`, () => {
        test('registers and unregisters a single provider', async () => {
            const { services, extensionHost } = await integrationTestContext()

            // Register the provider and call it.
            const unsubscribe = registerProvider(extensionHost)(['*'], labeledProvider('a'))
            await extensionHost.internal.sync()
            expect(
                await getResult(services)
                    .pipe(take(1))
                    .toPromise()
            ).toEqual(labeledProviderResults(['a']))

            // Unregister the provider and ensure it's removed.
            unsubscribe.unsubscribe()
            expect(
                await getResult(services)
                    .pipe(take(1))
                    .toPromise()
            ).toEqual(null)
        })

        test('supplies params to the provideXyz method', async () => {
            const { services, extensionHost } = await integrationTestContext()
            const { wait, done } = createBarrier()
            registerProvider(extensionHost)(
                ['*'],
                providerWithImpl((doc, pos) => {
                    assertToJSON(doc, { uri: 'file:///f', languageId: 'l', text: 't' })
                    assertToJSON(pos, { line: 1, character: 2 })
                    done()
                })
            )
            await extensionHost.internal.sync()
            await getResult(services)
                .pipe(take(1))
                .toPromise()
            await wait
        })

        test('supports multiple providers', async () => {
            const { services, extensionHost } = await integrationTestContext()

            // Register 2 providers with different results.
            registerProvider(extensionHost)(['*'], labeledProvider('a'))
            registerProvider(extensionHost)(['*'], labeledProvider('b'))
            await extensionHost.internal.sync()

            // Expect it to emit the first provider's result first (and not block on both providers being ready).
            expect(
                await getResult(services)
                    .pipe(
                        take(2),
                        bufferCount(2)
                    )
                    .toPromise()
            ).toEqual([labeledProviderResults(['a']), labeledProviderResults(['a', 'b'])])
        })
    })
}

function labeledDefinitionResults(labels: string[]): Location | Location[] {
    return labels.map(label => ({ uri: `file:///${label}`, range: undefined }))
}
