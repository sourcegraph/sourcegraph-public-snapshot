import { redactSensitiveInfoFromURL } from './util'

describe('tracking/util', () => {
    describe(`${redactSensitiveInfoFromURL.name}()`, () => {
        it('removes search queries from URLs', () => {
            expect(redactSensitiveInfoFromURL('https://sourcegraph.com/search?q=test+query')).toEqual(
                'https://sourcegraph.com/redacted?q=redacted'
            )
        })

        it('removes search queries from URLs but maintains marketing query params', () => {
            expect(
                redactSensitiveInfoFromURL(
                    'https://sourcegraph.com/search?q=test+query&utm_source=test&utm_campaign=test'
                )
            ).toEqual('https://sourcegraph.com/redacted?q=redacted&utm_source=test&utm_campaign=test')
        })

        it('removes all query params from URLs but maintains marketing query params', () => {
            expect(
                redactSensitiveInfoFromURL(
                    'https://sourcegraph.com/search?some_query_param=test+query&utm_source=test&utm_campaign=test&utm_content=test&utm_medium=test&utm_medium=test'
                )
            ).toEqual(
                'https://sourcegraph.com/redacted?some_query_param=redacted&utm_source=test&utm_campaign=test&utm_content=test&utm_medium=test&utm_medium=test'
            )
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
