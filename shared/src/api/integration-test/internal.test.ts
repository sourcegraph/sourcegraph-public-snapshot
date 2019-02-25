import { integrationTestContext } from './testHelpers'

describe('Internal (integration)', () => {
    test('constant values', async () => {
        const { extensionAPI } = await integrationTestContext()
        expect(extensionAPI.internal.sourcegraphURL.toString()).toEqual('https://example.com/')
        expect(extensionAPI.internal.clientApplication).toEqual('sourcegraph')
    })
})
