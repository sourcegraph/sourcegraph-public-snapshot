import { describe, expect, test } from '@jest/globals'

import { createService, Config, SourcegraphService } from '../../client/SourcegraphClient'

const sgConf: Config = {
    endpoint: 'https://scaletesting.sgdev.org',
    token: process.env.SG_TOKEN ?? '',
}

describe('integration test against dotCom', async () => {
    const sourcegraphService: SourcegraphService = await createService(sgConf)
    test('check current user', async () => {
        const username: string = await sourcegraphService.Users.currentUsername()
        console.log(username)
        expect(username).toBe('burmudar')
    })
})
