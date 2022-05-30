/* eslint-disable @typescript-eslint/ban-ts-comment */
import { combineLatest, of } from 'rxjs'
import sinon from 'sinon'

import type { FeatureFlagName } from '../featureFlags'

import { setFeatureFlagOverride } from './feature-flag-local-overrides'
import { FeatureFlagClient } from './FeatureFlagClient'

const enabledFlag = 'enabled-flag' as FeatureFlagName
const disabledFlag = 'disabled-flag' as FeatureFlagName
describe('FeatureFlagClient', () => {
    const mockRequestGraphQL = sinon.spy((query, variables) =>
        of({ data: { evaluateFeatureFlag: [enabledFlag].includes(variables.flagName) } })
    )
    beforeEach(() => {
        mockRequestGraphQL.resetHistory()
        // remove local overrides
        localStorage.clear()
    })
    it('does not make initial API call ', () => {
        // @ts-ignore
        new FeatureFlagClient(mockRequestGraphQL)
        sinon.assert.notCalled(mockRequestGraphQL)
    })

    it('returns [true] response from API call for feature flag evaluation', done => {
        // @ts-ignore
        const client = new FeatureFlagClient(mockRequestGraphQL)
        expect.assertions(1)

        client.get(enabledFlag).subscribe(value => {
            expect(value).toBe(true)
            sinon.assert.calledOnce(mockRequestGraphQL)
            done()
        })
    })

    it('returns [false] response from API call for feature flag evaluation', done => {
        // @ts-ignore
        const client = new FeatureFlagClient(mockRequestGraphQL, 1000)
        expect.assertions(1)

        client.get(disabledFlag).subscribe({
            next: value => {
                expect(value).toBe(false)
                sinon.assert.calledOnce(mockRequestGraphQL)
                done()
            },
            complete: () => {
                throw new Error('Should not be completed when passing refetch interval');
            },
        })
    })

    it('completes after single fall if no refetch interval passed', done => {
        // @ts-ignore
        const client = new FeatureFlagClient(mockRequestGraphQL)
        expect.assertions(1)

        client.get(enabledFlag).subscribe({
            next: value => {
                expect(value).toBe(true)
            },
            complete: () => {
                sinon.assert.calledOnce(mockRequestGraphQL)
                done()
            },
        })
    })

    it('makes only single API call per feature flag evaluation', done => {
        // @ts-ignore
        const client = new FeatureFlagClient(mockRequestGraphQL)
        expect.assertions(2)

        combineLatest([client.get(enabledFlag), client.get(enabledFlag)]).subscribe(([value1, value2]) => {
            expect(value1).toBe(true)
            expect(value2).toBe(true)
            sinon.assert.calledOnce(mockRequestGraphQL)
            done()
        })
    })

    it('updates on new/different value', done => {
        let index = -1
        const mockRequestGraphQL = sinon.spy((query, variables) => {
            index++
            return of({ data: { evaluateFeatureFlag: [enabledFlag].includes(variables.flagName) && index === 0 } })
        })

        // @ts-ignore
        const client = new FeatureFlagClient(mockRequestGraphQL, 1)
        expect.assertions(2)

        client.get(enabledFlag).subscribe(value => {
            if (index === 0) {
                expect(value).toBe(true)
            } else {
                expect(value).toBe(false)
                sinon.assert.calledTwice(mockRequestGraphQL)
                done()
            }
        })
    })

    describe('local overrides', () => {
        it('returns [false] override if it exists', done => {
            // @ts-ignore
            const client = new FeatureFlagClient(mockRequestGraphQL)
            setFeatureFlagOverride(enabledFlag, false)
            expect.assertions(1)

            client.get(enabledFlag).subscribe(value => {
                expect(value).toBe(false)
                sinon.assert.calledOnce(mockRequestGraphQL)
                done()
            })
        })

        it('returns [true] override if it exists', done => {
            // @ts-ignore
            const client = new FeatureFlagClient(mockRequestGraphQL)
            setFeatureFlagOverride(disabledFlag, true)
            expect.assertions(1)

            client.get(disabledFlag).subscribe(value => {
                expect(value).toBe(true)
                sinon.assert.calledOnce(mockRequestGraphQL)
                done()
            })
        })

        it('does not use non-boolean override', done => {
            // @ts-ignore
            const client = new FeatureFlagClient(mockRequestGraphQL)
            // @ts-ignore
            setFeatureFlagOverride(disabledFlag, 'something else')
            expect.assertions(1)

            client.get(disabledFlag).subscribe(value => {
                expect(value).toBe(false)
                sinon.assert.calledOnce(mockRequestGraphQL)
                done()
            })
        })
    })
})
