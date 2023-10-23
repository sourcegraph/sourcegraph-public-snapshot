import { describe, expect, it } from '@jest/globals'

import { prettyInterval } from './utils'

describe('prettyInterval', () => {
    it('should format 10 seconds as 10 seconds', () => {
        expect(prettyInterval(10)).toEqual('10 seconds')
    })

    it('should format 60 seconds as 1 minute', () => {
        expect(prettyInterval(60)).toEqual('1 minute')
    })

    it('should format 60*60 seconds as 1 hour', () => {
        expect(prettyInterval(60 * 60)).toEqual('1 hour')
    })

    it('should format 24*60*60 seconds as 1 day', () => {
        expect(prettyInterval(24 * 60 * 60)).toEqual('1 day')
    })

    it('should format 24*60*60 + 60*60 + 60 + 5 seconds as 1 day 1 hour 1 minute 5 seconds', () => {
        expect(prettyInterval(24 * 60 * 60 + 60 * 60 + 60 + 5)).toEqual('1 day 1 hour 1 minute 5 seconds')
    })

    it('should format 0 seconds as an empty string', () => {
        expect(prettyInterval(0)).toEqual('')
    })

    it('should format multiple days, hours and minutes correctly', () => {
        expect(prettyInterval(2 * 24 * 60 * 60 + 5 * 60 * 60 + 15 * 60)).toEqual('2 days 5 hours 15 minutes')
    })

    it('should handle plurals correctly', () => {
        expect(prettyInterval(2 * 60)).toEqual('2 minutes')
        expect(prettyInterval(3 * 60 * 60)).toEqual('3 hours')
        expect(prettyInterval(5 * 24 * 60 * 60)).toEqual('5 days')
    })
})
