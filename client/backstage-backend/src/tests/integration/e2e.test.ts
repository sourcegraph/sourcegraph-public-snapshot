import { describe, expect, test } from '@jest/globals'
import { createService, Config, SourcegraphService } from '../../client/SourcegraphClient'

const sgConf: Config = {
    endpoint: 'https://scaletesting.sgdev.org',
    token: process.env.SG_TOKEN ?? '',
}

const sourcegraphService: SourcegraphService = createService(sgConf)

describe('integration test against dotCom', () => {
    console.log('using config', sgConf)
    test('check current user', async () => {
        const username: string = await sourcegraphService.Users.currentUsername()
        console.log(username)
        expect(username).toBe('burmudar')
    })
})
