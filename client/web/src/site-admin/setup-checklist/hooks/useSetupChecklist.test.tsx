import { add, endOfYesterday, sub } from 'date-fns'
import { cloneDeep, merge } from 'lodash'

import { getLicenseSetupStatus } from './useSetupChecklist'

type DeepPartial<T> = T extends object
    ? {
          [P in keyof T]?: DeepPartial<T[P]>
      }
    : T

describe('useSetupChecklist', () => {
    describe('getLicenseSetupStatus', () => {
        const now = new Date(2023, 8, 10)
        const args: Parameters<typeof getLicenseSetupStatus>[0] = {
            productSubscription: {
                license: {
                    __typename: 'ProductLicenseInfo',
                    isValid: true,
                    expiresAt: add(now, { days: 100 }).toUTCString(),
                    tags: [],
                    userCount: 10,
                },
            },
            users: {
                totalCount: 9,
            },
        }
        const toArgs = (newArgs: DeepPartial<typeof args>) => merge(cloneDeep(args), newArgs)

        it('returns undefined if license is valid', () => {
            expect(getLicenseSetupStatus(args, now)).toBeUndefined()
        })

        it('returns reason if license is invalid', () => {
            const reason = getLicenseSetupStatus(toArgs({ productSubscription: { license: { isValid: false } } }), now)
            expect(reason).toBeDefined()
            expect(reason).toMatchSnapshot()
        })

        it('returns reason if free plan license', () => {
            const reason = getLicenseSetupStatus(
                toArgs({ productSubscription: { license: { tags: ['plan:free-1'] } } }),
                now
            )
            expect(reason).toBeDefined()
            expect(reason).toMatchSnapshot()
        })

        it('returns reason if exceeded allowed user count', () => {
            const reason = getLicenseSetupStatus(toArgs({ users: { totalCount: 11 } }), now)
            expect(reason).toBeDefined()
            expect(reason).toMatchSnapshot()
        })

        it('returns undefined if exceeded allowed user count but true-up', () => {
            const reason = getLicenseSetupStatus(
                toArgs({ users: { totalCount: 11 }, productSubscription: { license: { tags: ['true-up'] } } }),
                now
            )
            expect(reason).toBeUndefined()
        })

        it('returns reason if expired ', () => {
            const reason = getLicenseSetupStatus(
                toArgs({ productSubscription: { license: { expiresAt: sub(now, { days: 1 }).toUTCString() } } }),
                now
            )
            expect(reason).toBeDefined()
            expect(reason).toMatchSnapshot()
        })

        it('returns reason if expiring soon ', () => {
            const reason = getLicenseSetupStatus(
                toArgs({ productSubscription: { license: { expiresAt: add(now, { days: 6 }).toUTCString() } } }),
                now
            )
            expect(reason).toBeDefined()
            expect(reason).toMatchSnapshot()
        })

        it('returns undefined if dev plan ', () => {
            const reason = getLicenseSetupStatus(
                toArgs({
                    productSubscription: { license: { expiresAt: endOfYesterday().toUTCString(), tags: ['dev'] } },
                }),
                now
            )
            expect(reason).toBeUndefined()
        })
    })
})
