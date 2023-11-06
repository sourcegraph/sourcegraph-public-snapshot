import { afterAll, beforeAll, beforeEach, describe, expect, it } from '@jest/globals'
import fetch from 'jest-fetch-mock'
import MockDate from 'mockdate'

import { clearFetchCache, disableFetchCache, enableFetchCache, fetchCache, type FetchCacheResponse } from './fetchCache'

const EXPECTED_DATA = { foo: { bar: 'baz' } }
const TEST_URL = '/test/api'

const expectResponses = (responses: FetchCacheResponse<any>[]): void => {
    for (const actualResponse of responses) {
        expect(actualResponse.data).toEqual(EXPECTED_DATA)
    }
}

describe('fetchCache', () => {
    beforeAll(() => {
        fetch.enableMocks()
        MockDate.reset()
    })

    beforeEach(() => {
        fetch.resetMocks()
        clearFetchCache()
        fetch.mockResponse(JSON.stringify(EXPECTED_DATA))
    })

    afterAll(() => {
        fetch.disableMocks()
    })

    it('makes single request for similar [...args]', async () => {
        const responses = await Promise.all([
            fetchCache({ cacheMaxAge: 100, url: TEST_URL }),
            fetchCache({ cacheMaxAge: 500, url: TEST_URL }),
            fetchCache({ cacheMaxAge: 1000, url: TEST_URL }),
        ])

        expectResponses(responses)
        expect(fetch).toHaveBeenCalledTimes(1)
    })

    it('makes multiple requests for different [...args]', async () => {
        const responses = await Promise.all([
            fetchCache({ cacheMaxAge: 100, url: '/test/api-1' }),
            fetchCache({ cacheMaxAge: 100, url: '/test/api-2' }),
        ])

        expectResponses(responses)
        expect(fetch).toHaveBeenCalledTimes(2)
    })

    it('makes multiple requests in case of fetch error', async () => {
        expect.assertions(3)

        fetch.mockRejectOnce(new Error('Some error'))
        await fetchCache({ cacheMaxAge: 100, url: TEST_URL }).catch(error => expect(error).toBeTruthy())

        await new Promise(resolve => setTimeout(resolve, 0))

        const response = await fetchCache({ cacheMaxAge: 100, url: TEST_URL })
        expect(response.data).toEqual(EXPECTED_DATA)

        expect(fetch).toHaveBeenCalledTimes(2)
    })

    it('makes multiple requests in case of cache expiration', async () => {
        const responseOne = await fetchCache({ cacheMaxAge: 1, url: TEST_URL })
        MockDate.set(Date.now() + 3)
        const responseTwo = await fetchCache({ cacheMaxAge: 1, url: TEST_URL })

        expectResponses([responseOne, responseTwo])
        expect(fetch).toHaveBeenCalledTimes(2)
    })

    it('makes multiple requests if [cacheMaxAge=0]', async () => {
        const responses = await Promise.all([
            fetchCache({ cacheMaxAge: 0, url: TEST_URL }),
            fetchCache({ cacheMaxAge: 0, url: TEST_URL }),
        ])

        expectResponses(responses)
        expect(fetch).toHaveBeenCalledTimes(2)
    })

    it('makes multiple requests when caching is disabled', async () => {
        disableFetchCache()
        const responses = await Promise.all([
            fetchCache({ cacheMaxAge: 100, url: TEST_URL }),
            fetchCache({ cacheMaxAge: 100, url: TEST_URL }),
        ])

        expectResponses(responses)
        expect(fetch).toHaveBeenCalledTimes(2)
        enableFetchCache()
    })
})
