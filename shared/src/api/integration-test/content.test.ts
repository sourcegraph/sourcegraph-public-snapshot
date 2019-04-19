import { take } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { integrationTestContext } from './testHelpers'

describe('content (integration)', () => {
    test('registers a link preview provider', async () => {
        const { services, extensionAPI } = await integrationTestContext()

        // Register the provider and call it.
        extensionAPI.content.registerLinkPreviewProvider('http://example.com', {
            provideLinkPreview: url => ({
                content: { value: `xyz ${url.toString()}`, kind: sourcegraph.MarkupKind.PlainText },
            }),
        })
        await extensionAPI.internal.sync()
        expect(
            await services.linkPreviews
                .provideLinkPreview('http://example.com/foo')
                .pipe(take(1))
                .toPromise()
        ).toEqual({
            content: [{ value: 'xyz http://example.com/foo', kind: sourcegraph.MarkupKind.PlainText }],
            hover: [],
        })
    })

    test('unregisters a link preview provider', async () => {
        const { services, extensionAPI } = await integrationTestContext()

        // Register the provider and call it.
        const subscription = extensionAPI.content.registerLinkPreviewProvider('http://example.com', {
            provideLinkPreview: () => ({ content: { value: 'bar', kind: sourcegraph.MarkupKind.PlainText } }),
        })
        await extensionAPI.internal.sync()

        // Unregister the provider and ensure it's removed.
        subscription.unsubscribe()
        await extensionAPI.internal.sync()
        expect(
            await services.linkPreviews
                .provideLinkPreview('http://example.com/foo')
                .pipe(take(1))
                .toPromise()
        ).toEqual(null)
    })

    test('supports multiple link preview providers', async () => {
        const { services, extensionAPI } = await integrationTestContext()

        // Register the provider and call it
        extensionAPI.content.registerLinkPreviewProvider('http://example.com', {
            provideLinkPreview: () => ({ content: { value: 'qux', kind: sourcegraph.MarkupKind.PlainText } }),
        })
        extensionAPI.content.registerLinkPreviewProvider('http://example.com', {
            provideLinkPreview: () => ({ content: { value: 'zip', kind: sourcegraph.MarkupKind.PlainText } }),
        })
        await extensionAPI.internal.sync()
        expect(
            await services.linkPreviews
                .provideLinkPreview('http://example.com/foo')
                .pipe(take(1))
                .toPromise()
        ).toEqual({
            content: [
                { value: 'qux', kind: sourcegraph.MarkupKind.PlainText },
                { value: 'zip', kind: sourcegraph.MarkupKind.PlainText },
            ],
            hover: [],
        })
    })
})
