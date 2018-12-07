import * as assert from 'assert'
import { integrationTestContext } from './helpers.test'

describe('Internal (integration)', () => {
    it('constant values', async () => {
        const { extensionHost } = await integrationTestContext()
        assert.deepStrictEqual(extensionHost.internal.sourcegraphURL.toString(), 'https://example.com')
        assert.deepStrictEqual(extensionHost.internal.clientApplication, 'sourcegraph')
    })
})
