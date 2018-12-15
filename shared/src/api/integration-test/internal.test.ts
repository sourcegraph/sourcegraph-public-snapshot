import assert from 'assert'
import { integrationTestContext } from './helpers.test'

describe('Internal (integration)', () => {
    it('constant values', async () => {
        const { extensionHost } = await integrationTestContext()
        assert.deepEqual(extensionHost.internal.sourcegraphURL.toString(), 'https://example.com')
        assert.deepEqual(extensionHost.internal.clientApplication, 'sourcegraph')
    })
})
