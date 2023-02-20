import { describe, expect, test, jest } from '@jest/globals'
import { SourcegraphClient, SourcegraphService, BaseClient } from './SourcegraphClient'
import { Query } from './Query'

let mockClient: BaseClient = new BaseClient('', '', '')

let sourcegraphService: SourcegraphService = new SourcegraphClient(mockClient)

describe('SourcegraphService', () => {
    test('testquery', async () => {
        // setup
        jest.spyOn(mockClient, 'fetch').mockImplementationOnce(async () => {
            console.log('hello from the mock impl')
            return {
                data: {
                    currentUser: {
                        username: 'william',
                    },
                },
            }
        })
        // test
        const username = await sourcegraphService.Users.currentUsername()
        // verify
        expect(username).toBe('william')
    })
})
