import { describe, expect, test } from '@jest/globals'
import { SearchRepoQuery, SearchResults } from '../../client/Query'
import { createService, Config, SourcegraphService } from '../../client/SourcegraphClient'

const sgConf: Config = {
  endpoint: 'https://scaletesting.sgdev.org',
  token: process.env.SG_TOKEN ?? '',
}

let client: SourcegraphService
async function getClient(): Promise<SourcegraphService> {
  if (!client) {
    client = await createService(sgConf)
  }

  return client
}

describe('basic api check', () => {
  test('check current user', async () => {
    const sourcegraphService: SourcegraphService = await getClient()
    const username: string = await sourcegraphService.Users.currentUsername()
    console.log(username)
    expect(username).toBe('burmudar')
  })
})

describe('pagination tests', () => {
  test('get 50 repos', async () => {
    const sourcegraphService: SourcegraphService = await getClient()
    const query: SearchRepoQuery = new SearchRepoQuery("repo:.* count:50")
    const results: SearchResults = await sourcegraphService.Search.doQuery(query)

    expect(results).toHaveLength(50)
  })
})
