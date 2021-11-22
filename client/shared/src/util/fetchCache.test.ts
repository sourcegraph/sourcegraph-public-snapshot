import fetch from 'jest-fetch-mock'
import MockDate from 'mockdate'

import { fetchCache, FetchCacheReturnType } from './fetchCache'

const EXPECTED_DATA = { foo: { bar: 'baz' } }
const TEST_URL = '/test/api'

const expectResponses = (responses: [FetchCacheReturnType<any>, FetchCacheReturnType<any>]): void => {
    for (const actualResponse of responses) {
        expect(actualResponse.data).toEqual(EXPECTED_DATA)
    }
}

describe('memoizedFetch', () => {
    beforeAll(() => {
        fetch.enableMocks()
        MockDate.reset()
    })

    beforeEach(() => {
        fetch.mockClear()
        fetchCache.clear()
        fetch.mockResponse(JSON.stringify(EXPECTED_DATA))
    })

    afterAll(() => {
        fetch.disableMocks()
    })

    it('makes single request for similar [...args]', async () => {
        const responses = await Promise.all([fetchCache(1, TEST_URL), fetchCache(1, TEST_URL)])

        expectResponses(responses)
        expect(fetch).toHaveBeenCalledTimes(1)
    })

    it('makes single request for similar [...args] with different maxAge', async () => {
        const responses = await Promise.all([fetchCache(100, TEST_URL), fetchCache(1, TEST_URL)])

        expectResponses(responses)
        expect(fetch).toHaveBeenCalledTimes(1)
    })

    it('makes multiple requests when cache item is expired', async () => {
        const responseOne = await fetchCache(1, TEST_URL)
        await new Promise(resolve => setTimeout(resolve, 3))
        const responseTwo = await fetchCache(1, TEST_URL)

        expectResponses([responseOne, responseTwo])
        expect(fetch).toHaveBeenCalledTimes(2)
    })

    it('makes multiple requests for different [...args]', async () => {
        const responses = await Promise.all([fetchCache(1, '/test/api-1'), fetchCache(1, '/test/api-2')])

        expectResponses(responses)
        expect(fetch).toHaveBeenCalledTimes(2)
    })

    it('makes multiple requests when (timeout = 0)', async () => {
        const responses = await Promise.all([fetchCache(0, TEST_URL), fetchCache(0, TEST_URL)])

        expectResponses(responses)
        expect(fetch).toHaveBeenCalledTimes(2)
    })

    it('makes multiple requests when caching is disabled', async () => {
        fetchCache.disable()
        const responses = await Promise.all([fetchCache(1, TEST_URL), fetchCache(1, TEST_URL)])

        expectResponses(responses)
        expect(fetch).toHaveBeenCalledTimes(2)
        fetchCache.enable()
    })
})
