import { describe, expect, it } from 'vitest'

import { formatInviteDate } from './TeamMemberList'

describe('formatInviteDate', () => {
    it('shows relative descriptions of time in the desired format', () => {
        // ISO-8601 (RFC3339) strings, just the way the backend returns them
        const inputDates = [
            '2024-05-22T15:59:55.000000+00:00',
            '2024-05-22T14:17:55.000000+00:00',
            '2024-05-21T14:17:55.000000+00:00',
            '2024-05-15T14:17:55.000000+00:00',
        ]

        const now = new Date('2024-05-22T16:00:00.000000+00:00')

        const expectedOutput = ['5 seconds ago', '1 hour ago', 'yesterday', 'last week']

        const outputDates = inputDates.map(date => formatInviteDate(date, now))

        expect(outputDates).toEqual(expectedOutput)
    })

    it('handles malformed input', () => {
        // These are in the format of the normal output of our Go back end
        const inputDates = [null, '', '1T14:17:55.000000+00:00', '2024-05-15T14:17:55']

        const now = new Date('2024-05-22T16:00:00.000000+00:00')

        const expectedOutput = ['', '', '', 'last week']

        const outputDates = inputDates.map(date => formatInviteDate(date, now))

        expect(outputDates).toEqual(expectedOutput)
    })
})
