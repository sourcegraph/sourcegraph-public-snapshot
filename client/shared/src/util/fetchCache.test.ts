import fetch from 'jest-fetch-mock'

import { fetchCache } from './fetchCache'

describe('memoizedFetch', () => {
    beforeAll(() => {
        fetch.enableMocks()
    })

    beforeEach(() => {
        fetch.mockClear()
    })

    afterAll(() => {
        fetch.disableMocks()
    })

    it('makes only single request for similar requests', async () => {
        const expectedData = { foo: { bar: 'baz' } }
        fetch.mockResponseOnce(JSON.stringify(expectedData))

        const testUrl = '/test/api'
        const responses = await Promise.all([fetchCache(testUrl), fetchCache(testUrl)])

        for (const actualResponse of responses) {
            expect(actualResponse.data).toEqual(expectedData)
        }
        expect(fetch).toHaveBeenCalledTimes(1)
    })

    it('makes multiple requests for different requests', async () => {
        const expectedData = { foo: { bar: 'baz' } }

        fetch.mockResponse(JSON.stringify(expectedData))

        const responses = await Promise.all([fetchCache('/test/one'), fetchCache('/test/two')])

        for (const actualResponse of responses) {
            expect(actualResponse.data).toEqual(expectedData)
        }
        expect(fetch).toHaveBeenCalledTimes(2)
    })
})
