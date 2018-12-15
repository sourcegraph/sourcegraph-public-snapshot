import { integrationTestContext } from './helpers.test'

describe('Internal (integration)', () => {
    test('constant values', async () => {
        const { extensionHost } = await integrationTestContext()
        expect(extensionHost.internal.sourcegraphURL.toString()).toEqual('https://example.com')
        expect(extensionHost.internal.clientApplication).toEqual('sourcegraph')
    })
})
