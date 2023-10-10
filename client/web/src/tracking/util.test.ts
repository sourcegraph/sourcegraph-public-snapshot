import { getPreviousMonday, redactSensitiveInfoFromAppURL } from './util'

describe('tracking/util', () => {
    describe(`${redactSensitiveInfoFromAppURL.name}()`, () => {
        it('removes search queries from URLs when originating from cloud instance', () => {
            expect(redactSensitiveInfoFromAppURL('https://sourcegraph.sourcegraph.com/search?q=test+query')).toEqual(
                'https://sourcegraph.sourcegraph.com/search/redacted?q=redacted'
            )
        })

        it('removes search queries from URLs but maintains marketing query params from cloud instances', () => {
            expect(
                redactSensitiveInfoFromAppURL(
                    'https://sourcegraph.sourcegraph.com/search?q=test+query&utm_source=test&utm_campaign=test&utm_cid=test'
                )
            ).toEqual(
                'https://sourcegraph.sourcegraph.com/search/redacted?q=redacted&utm_source=test&utm_campaign=test&utm_cid=test'
            )
        })

        it('removes all query params from URLs but maintains marketing query params from cloud instances', () => {
            expect(
                redactSensitiveInfoFromAppURL(
                    'https://sourcegraph.sourcegraph.com/search?some_query_param=test+query&utm_source=test&utm_campaign=test&utm_content=test&utm_medium=test&utm_medium=test'
                )
            ).toEqual(
                'https://sourcegraph.sourcegraph.com/search/redacted?some_query_param=redacted&utm_source=test&utm_campaign=test&utm_content=test&utm_medium=test&utm_medium=test'
            )
        })

        it('removes repo information from URLs from cloud instances', () => {
            expect(
                redactSensitiveInfoFromAppURL(
                    'https://sourcegraph.sourcegraph.com/github.com/test/test?utm_source=test&utm_campaign=test'
                )
            ).toEqual('https://sourcegraph.sourcegraph.com/github.com/redacted?utm_source=test&utm_campaign=test')
        })

        it('removes repo and file information from URLs from cloud instances', () => {
            expect(
                redactSensitiveInfoFromAppURL(
                    'https://sourcegraph.sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/test?utm_source=test&utm_campaign=test'
                )
            ).toEqual('https://sourcegraph.sourcegraph.com/github.com/redacted?utm_source=test&utm_campaign=test')
        })

        it('redact pathnames from unknown app URLs from cloud instances', () => {
            expect(
                redactSensitiveInfoFromAppURL('https://sourcegraph.sourcegraph.com/notebooks/Tm90ZWJvb2s6MTMwMA==')
            ).toEqual('https://sourcegraph.sourcegraph.com/redacted')
        })

        it('does not redact from sourcegraph subdomins (signup)', () => {
            expect(
                redactSensitiveInfoFromAppURL(
                    'https://signup.sourcegraph.com/case-studies?utm_source=test&utm_campaign=test'
                )
            ).toEqual('https://signup.sourcegraph.com/case-studies?utm_source=test&utm_campaign=test')
        })

        it('does not redact from sourcegraph subdomains (about)', () => {
            expect(
                redactSensitiveInfoFromAppURL(
                    'https://about.sourcegraph.com/case-studies?utm_source=test&utm_campaign=test'
                )
            ).toEqual('https://about.sourcegraph.com/case-studies?utm_source=test&utm_campaign=test')
        })

        it('does not redact url when request is from sourcegraph', () => {
            expect(redactSensitiveInfoFromAppURL('https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/test?utm_source=test&utm_campaign=test')).toEqual(
                'https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/test?utm_source=test&utm_campaign=test'
            )
        })

    })

    describe(`${getPreviousMonday.name}()`, () => {
        it('gets the current day if it is a Monday', () => {
            const date = new Date(2021, 5, 14) // June 14, 2021 is a Monday
            const monday = getPreviousMonday(date)
            expect(monday).toBe('2021-06-14')
        })

        it('gets the previous Monday if it is not a Monday', () => {
            const date = new Date(2021, 5, 13) // June 13, 2021 is a Sunday
            const monday = getPreviousMonday(date)
            expect(monday).toBe('2021-06-07')
        })

        it('gets the previous Monday if it is in a different month', () => {
            const date = new Date(2021, 5, 3) // June 3, 2021 is a Thursday
            const monday = getPreviousMonday(date)
            expect(monday).toBe('2021-05-31')
        })

        it('gets the previous Monday if it is in a different year', () => {
            const date = new Date(2021, 0, 2) // Jan 2, 2021 is a Saturday
            const monday = getPreviousMonday(date)
            expect(monday).toBe('2020-12-28')
        })
    })
})
