import { describe, expect, it } from 'vitest'
import { formatInviteDate } from './TeamMemberList'

// This is a bit weird test, but I wanted to give the design some examples for how the formatting looks,
// and this was the easiest way to do that. Rather than throwing away the test, I'm leaving it here.
describe('getFilterFnsFromCodyContextFilters', () => {
    it('looks good for typical input', async () => {
        // These are in the format of the normal output of our Go back end
        const inputDates = [
            '2024-05-22T15:59:55.000000+00:00',
            '2024-05-22T14:17:55.000000+00:00',
            '2024-05-21T14:17:55.000000+00:00',
            '2024-05-15T14:17:55.000000+00:00',
            ]

        const now = new Date('2024-05-22T16:00:00.000000+00:00')

        const expectedOutput = [
            '5 seconds ago',
            '1 hour ago',
            'yesterday',
            'last week',
            ]

        const outputDates = inputDates.map(date => formatInviteDate(date, now))

        expect(outputDates).toEqual(expectedOutput)
    })

    it('does not fail too badly with atypical input', async () => {
        // These are in the format of the normal output of our Go back end
        const inputDates = [
            null,
            '',
            '1T14:17:55.000000+00:00',
            '2024-05-15T14:17:55',
            ]

        const now = new Date('2024-05-22T16:00:00.000000+00:00')

        const expectedOutput = [
            '',
            '',
            '',
            'last week',
        ]

        const outputDates = inputDates.map(date => formatInviteDate(date, now))

        expect(outputDates).toEqual(expectedOutput)
    })
})
