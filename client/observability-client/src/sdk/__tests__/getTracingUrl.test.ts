import { getTracingURL, TRACING_URL_SUFFIX } from '../getTracingUrl'

const BASE_URL = 'https://sourcegraph.test:3443'
const BASE_URL_WITH_SLASH = `${BASE_URL}/`
const RELATIVE_ENDPOINT = '-/debug/otlp'

describe('getTracingUrl', () => {
    describe('creates tracing URL with absolute endpoint', () => {
        const absoluteEndpoint = new URL(RELATIVE_ENDPOINT, BASE_URL).toString()
        const expectedURL = `${absoluteEndpoint}/${TRACING_URL_SUFFIX}`

        it('with ending slash', () => {
            expect(getTracingURL(`${absoluteEndpoint}/`, BASE_URL)).toBe(expectedURL)
            expect(getTracingURL(`${absoluteEndpoint}/`, BASE_URL_WITH_SLASH)).toBe(expectedURL)
        })

        it('without ending slash', () => {
            expect(getTracingURL(absoluteEndpoint, BASE_URL)).toBe(expectedURL)
            expect(getTracingURL(absoluteEndpoint, BASE_URL_WITH_SLASH)).toBe(expectedURL)
        })
    })

    describe('creates tracing URL with relative endpoint', () => {
        const expectedURL = `${BASE_URL}/${RELATIVE_ENDPOINT}/${TRACING_URL_SUFFIX}`

        it('with ending slash', () => {
            expect(getTracingURL(`${RELATIVE_ENDPOINT}/`, BASE_URL)).toBe(expectedURL)
            expect(getTracingURL(`${RELATIVE_ENDPOINT}/`, BASE_URL_WITH_SLASH)).toBe(expectedURL)
        })

        it('without ending slash', () => {
            expect(getTracingURL(RELATIVE_ENDPOINT, BASE_URL)).toBe(expectedURL)
            expect(getTracingURL(RELATIVE_ENDPOINT, BASE_URL_WITH_SLASH)).toBe(expectedURL)
        })
    })
})
