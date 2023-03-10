import { describe, expect, test } from '@jest/globals'

import { createService, Config } from './SourcegraphClient'

const INTEGRATION_TEST = true

describe.skip('query marshalling', () => {
    if (!INTEGRATION_TEST) {
        return
    }

    const config: Config = {
        // TODO(@burmudar): this should be made configurable, even though it is for an integration test
        endpoint: 'https://sourcegraph.test:3443',
        token: 'SOURCEGRAPH_TOKEN' in process.env ? (process.env['SOURCEGRAPH_TOKEN'] as string) : '',
        sudoUsername: 'sourcegraph',
    }

    const client = createService(config)
    test('check CurrentUser is marshalled', () => {
        expect(client.Users.currentUsername).toBe('sourcegraph')
    })

    test('check AuthenticatedUser is marshalled', () => {
        expect(client.Users.getAuthenticatedUser()).toBe({})
    })
})
