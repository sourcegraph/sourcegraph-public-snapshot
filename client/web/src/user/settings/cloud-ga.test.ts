import { cleanup } from '@testing-library/react'

import { gitlabTokenExpired } from './cloud-ga'

describe('gitlabTokenExpired tests', () => {
    const DATE_NOW = '1646922320064'

    afterAll(cleanup)

    beforeEach(() => {
        jest.useFakeTimers()
        jest.setSystemTime(Number(DATE_NOW))
    })

    afterEach(() => {
        jest.runOnlyPendingTimers()
        jest.useRealTimers()
    })

    describe('happy path', () => {
        it('returns false if the token is not expired', () => {
            const config = `{"token.type": "oauth", "token.oauth.expiry": ${Number(DATE_NOW) / 1000 + 3600}}`
            expect(gitlabTokenExpired(config)).toBeFalsy()
        })

        it('returns true if the token is expired', () => {
            const config = `{"token.type": "oauth", "token.oauth.expiry": ${Number(DATE_NOW) / 1000 - 100}}`
            expect(gitlabTokenExpired(config)).toBeTruthy()
        })

        it('returns true if the expiry time is not defined', () => {
            const config = '{"token.type": "oauth"}'
            expect(gitlabTokenExpired(config)).toBeTruthy()
        })
    })

    describe('non happy path', () => {
        it('returns false if the token is not an oauth token', () => {
            const config = '{"token.type": "pat"}'
            expect(gitlabTokenExpired(config)).toBeFalsy()
        })

        it('returns false if the token type is missing', () => {
            const config = '{"token.oauth.expiry": "123"}'
            expect(gitlabTokenExpired(config)).toBeFalsy()
        })

        it('returns false if the config is undefined', () => {
            expect(gitlabTokenExpired(undefined)).toBeFalsy()
        })

        it('returns false if the json config cannot be parsed', () => {
            const invalidJSON = '{{{{{}'
            expect(gitlabTokenExpired(invalidJSON)).toBeFalsy()
        })
    })
})
