import { describe, expect, test } from '@jest/globals'
import { SearchMatch } from '@sourcegraph/shared/src/search/stream'
import { Observable, Subscriber, Subscription, NextObserver } from 'rxjs'
import { first, materialize, publish, scan, share, take, tap } from 'rxjs/operators'
import { createService, Config, SourcegraphService, PaginatedResult, PageInfo, SearchMatches, SearchEvent } from '../../client/SourcegraphClient'

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
    // No logs get printed
    const result: SearchMatches = new Array<SearchMatch>()
    const obs: Observable<SearchMatches> = sourcegraphService.Search.searchStream(`repo:.*`)
    obs.subscribe((data) => console.log("data", data), (e) => console.error("err", e), () => console.log("complete!"))
    obs.pipe(share())

    expect(result).toHaveLength(50)
  })
})

