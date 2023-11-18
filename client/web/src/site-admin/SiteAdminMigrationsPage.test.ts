import { parse } from 'semver'
import { describe, expect, it } from 'vitest'

import type { OutOfBandMigrationFields } from '../graphql-operations'

import { isComplete, isInvalidForVersion } from './SiteAdminMigrationsPage'

describe('isComplete', () => {
    it('should interpret ->0% as incomplete', () => {
        expect(isComplete({ ...zeroFields, progress: 0 })).toBeFalsy()
    })
    it('should interpret ->50% as incomplete', () => {
        expect(isComplete({ ...zeroFields, progress: 0.5 })).toBeFalsy()
    })
    it('should interpret ->100% as complete', () => {
        expect(isComplete({ ...zeroFields, progress: 1 })).toBeTruthy()
    })

    it('should interpret <-0% as complete', () => {
        expect(isComplete({ ...zeroFields, progress: 0, applyReverse: true })).toBeTruthy()
    })
    it('should interpret <-50% as incomplete', () => {
        expect(isComplete({ ...zeroFields, progress: 0.5, applyReverse: true })).toBeFalsy()
    })
    it('should interpret <-100% as incomplete', () => {
        expect(isComplete({ ...zeroFields, progress: 1, applyReverse: true })).toBeFalsy()
    })
})

describe('isInvalidForVersion', () => {
    it('should mark incomplete migrations as invalid after upgrade', () => {
        expect(
            isInvalidForVersion(
                { ...zeroFields, introduced: '3.9', deprecated: '3.10', progress: 0.5 },
                parse('3.10.0')
            )
        ).toBeTruthy()
    })

    it('should not mark incomplete migrations as invalid without deprecation', () => {
        expect(isInvalidForVersion({ ...zeroFields, introduced: '3.9', progress: 0.5 }, parse('3.10.0'))).toBeFalsy()
    })

    it('should mark incomplete migrations as invalid after downgrade', () => {
        expect(
            isInvalidForVersion({ ...zeroFields, introduced: '3.9', deprecated: '3.10', progress: 0.5 }, parse('3.8.0'))
        ).toBeTruthy()
    })

    it('should not mark incomplete migrations as invalid after downgrade if non-destructive', () => {
        expect(
            isInvalidForVersion(
                { ...zeroFields, introduced: '3.9', deprecated: '3.10', progress: 0.5, nonDestructive: true },
                parse('3.8.0')
            )
        ).toBeFalsy()
    })
})

const zeroFields: OutOfBandMigrationFields = {
    id: '',
    team: '',
    component: '',
    description: '',
    introduced: '',
    deprecated: '',
    progress: 0,
    created: '',
    lastUpdated: null,
    nonDestructive: false,
    applyReverse: false,
    errors: [],
}
