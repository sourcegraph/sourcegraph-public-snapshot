import { redactSensitiveInfoFromURL } from './util'

describe('tracking/analyticsUtils', () => {
    describe(`${redactSensitiveInfoFromURL.name}()`, () => {
        it('removes search queries from URLs', () => {
            expect(redactSensitiveInfoFromURL('https://sourcegraph.com/search?q=test+query')).toEqual(
                'https://sourcegraph.com/redacted?q=redacted'
            )
        })

        it('removes search queries from URLs but maintains other query params', () => {
            expect(
                redactSensitiveInfoFromURL(
                    'https://sourcegraph.com/search?q=test+query&utm_source=test&utm_campaign=test'
                )
            ).toEqual('https://sourcegraph.com/redacted?q=redacted&utm_source=test&utm_campaign=test')
        })

        it('removes repo information from URLs', () => {
            expect(
                redactSensitiveInfoFromURL(
                    'https://sourcegraph.com/github.com/test/test?utm_source=test&utm_campaign=test'
                )
            ).toEqual('https://sourcegraph.com/redacted?utm_source=test&utm_campaign=test')
        })

        it('removes repo and file information from URLs', () => {
            expect(
                redactSensitiveInfoFromURL(
                    'https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/test?utm_source=test&utm_campaign=test'
                )
            ).toEqual('https://sourcegraph.com/redacted?utm_source=test&utm_campaign=test')
        })
    })
})
