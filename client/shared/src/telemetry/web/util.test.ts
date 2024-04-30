import { describe, expect, it } from 'vitest'

import { getPreviousMonday } from './util'

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
