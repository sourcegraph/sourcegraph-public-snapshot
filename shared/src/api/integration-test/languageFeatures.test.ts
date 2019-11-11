import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { Location } from '@sourcegraph/extension-api-types'
import { asyncScheduler, Observable, of } from 'rxjs'
import { observeOn, switchMap, take, toArray } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { Services } from '../client/services'
import { assertToJSON, createBarrier, integrationTestContext } from './testHelpers'

describe('LanguageFeatures (integration)', () => {
    testLocationProvider<sourcegraph.HoverProvider>({
        name: 'registerHoverProvider',
        registerProvider: extensionAPI => (s, p) => extensionAPI.languages.registerHoverProvider(s, p),
        labeledProvider: label => ({
            provideHover: (doc: sourcegraph.TextDocument, pos: sourcegraph.Position) =>
                of({
                    contents: { value: label, kind: MarkupKind.PlainText },
                }).pipe(observeOn(asyncScheduler)),
        }),
        labeledProviderResults: labels => ({
            contents: labels.map(label => ({ value: label, kind: MarkupKind.PlainText })),
        }),
        providerWithImplementation: run => ({ provideHover: run } as sourcegraph.HoverProvider),
        getResult: (services, uri) =>
            services.textDocumentHover.getHover({
                textDocument: { uri },
                position: { line: 1, character: 2 },
            }),
    })
    testLocationProvider<sourcegraph.DefinitionProvider>({
        name: 'registerDefinitionProvider',
        registerProvider: extensionAPI => extensionAPI.languages.registerDefinitionProvider,
        labeledProvider: label => ({
            provideDefinition: (doc: sourcegraph.TextDocument, pos: sourcegraph.Position) =>
                of([{ uri: new URL(`file:///${label}`) }]).pipe(observeOn(asyncScheduler)),
        }),
        labeledProviderResults: labeledDefinitionResults,
        providerWithImplementation: run => ({ provideDefinition: run } as sourcegraph.DefinitionProvider),
        getResult: (services, uri) =>
            services.textDocumentDefinition
                .getLocations({
                    textDocument: { uri },
                    position: { line: 1, character: 2 },
                })
                .pipe(switchMap(locations => locations)),
    })
    testLocationProvider<sourcegraph.ReferenceProvider>({
        name: 'registerReferenceProvider',
        registerProvider: extensionAPI => extensionAPI.languages.registerReferenceProvider,
        labeledProvider: label => ({
            provideReferences: (
                doc: sourcegraph.TextDocument,
                pos: sourcegraph.Position,
                context: sourcegraph.ReferenceContext
            ) => of([{ uri: new URL(`file:///${label}`) }]).pipe(observeOn(asyncScheduler)),
        }),
        labeledProviderResults: labels => labels.map(label => ({ uri: `file:///${label}`, range: undefined })),
        providerWithImplementation: run =>
            ({
                provideReferences: (
                    doc: sourcegraph.TextDocument,
                    pos: sourcegraph.Position,
                    _context: sourcegraph.ReferenceContext
                ) => run(doc, pos),
            } as sourcegraph.ReferenceProvider),
        getResult: (services, uri) =>
            services.textDocumentReferences
                .getLocations({
                    textDocument: { uri },
                    position: { line: 1, character: 2 },
                    context: { includeDeclaration: true },
                })
                .pipe(switchMap(locations => locations)),
    })
    testLocationProvider<sourcegraph.LocationProvider>({
        name: 'registerLocationProvider',
        registerProvider: extensionAPI => (selector, provider) =>
            extensionAPI.languages.registerLocationProvider('x', selector, provider),
        labeledProvider: label => ({
            provideLocations: (doc: sourcegraph.TextDocument, pos: sourcegraph.Position) =>
                of([{ uri: new URL(`file:///${label}`) }]).pipe(observeOn(asyncScheduler)),
        }),
        labeledProviderResults: labels => labels.map(label => ({ uri: `file:///${label}`, range: undefined })),
        providerWithImplementation: run =>
            ({
                provideLocations: (doc: sourcegraph.TextDocument, pos: sourcegraph.Position) => run(doc, pos),
            } as sourcegraph.LocationProvider),
        getResult: (services, uri) =>
            services.textDocumentLocations
                .getLocations('x', {
                    textDocument: { uri },
                    position: { line: 1, character: 2 },
                })
                .pipe(switchMap(x => x)),
    })
})

/**
 * Generates test cases for sourcegraph.languages.registerXyzProvider functions and their associated
 * XyzProviders, for providers that return a list of locations.
 */
function testLocationProvider<P>({
    name,
    registerProvider,
    labeledProvider,
    labeledProviderResults,
    providerWithImplementation,
    getResult,
}: {
    name: keyof typeof sourcegraph.languages
    registerProvider: (
        extensionAPI: typeof sourcegraph
    ) => (selector: sourcegraph.DocumentSelector, provider: P) => sourcegraph.Unsubscribable
    labeledProvider: (label: string) => P
    labeledProviderResults: (labels: string[]) => any
    providerWithImplementation: (run: (doc: sourcegraph.TextDocument, pos: sourcegraph.Position) => void) => P
    getResult: (services: Services, uri: string) => Observable<any>
}): void {
    describe(`languages.${name}`, () => {
        test('registers and unregisters a single provider', async () => {
            const { services, extensionAPI } = await integrationTestContext()

            // Register the provider and call it.
            const subscription = registerProvider(extensionAPI)(['*'], labeledProvider('a'))
            await extensionAPI.internal.sync()
            expect(
                await getResult(services, 'file:///f')
                    .pipe(take(1))
                    .toPromise()
            ).toEqual(labeledProviderResults(['a']))

            // Unregister the provider and ensure it's removed.
            subscription.unsubscribe()
            expect(
                await getResult(services, 'file:///f')
                    .pipe(take(1))
                    .toPromise()
            ).toEqual(null)
        })

        test('syncs with models', async () => {
            const { services, extensionAPI } = await integrationTestContext()

            const subscription = registerProvider(extensionAPI)(['*'], labeledProvider('a'))
            await extensionAPI.internal.sync()

            services.model.addModel({ uri: 'file:///f2', languageId: 'l1', text: 't1' })
            services.editor.addEditor({
                type: 'CodeEditor',
                resource: 'file:///f2',
                selections: [],
                isActive: true,
            })

            expect(
                await getResult(services, 'file:///f2')
                    .pipe(take(1))
                    .toPromise()
            ).toEqual(labeledProviderResults(['a']))

            subscription.unsubscribe()
        })

        test('supplies params to the provideXyz method', async () => {
            const { services, extensionAPI } = await integrationTestContext()
            const { wait, done } = createBarrier()
            registerProvider(extensionAPI)(
                ['*'],
                providerWithImplementation((doc, pos) => {
                    assertToJSON(doc, { uri: 'file:///f', languageId: 'l', text: 't' })
                    assertToJSON(pos, { line: 1, character: 2 })
                    done()
                })
            )
            await extensionAPI.internal.sync()
            await getResult(services, 'file:///f')
                .pipe(take(1))
                .toPromise()
            await wait
        })

        test('supports multiple providers', async () => {
            const { services, extensionAPI } = await integrationTestContext()

            // Register 2 providers with different results.
            registerProvider(extensionAPI)(['*'], labeledProvider('a'))
            registerProvider(extensionAPI)(['*'], labeledProvider('b'))
            await extensionAPI.internal.sync()

            // Expect it to emit the first provider's result first (and not block on both providers being ready).
            expect(
                await getResult(services, 'file:///f')
                    .pipe(take(2), toArray())
                    .toPromise()
            ).toEqual([labeledProviderResults(['a']), labeledProviderResults(['a', 'b'])])
        })
    })
}

function labeledDefinitionResults(labels: string[]): Location[] {
    return labels.map(label => ({ uri: `file:///${label}`, range: undefined }))
}
