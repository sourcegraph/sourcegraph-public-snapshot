import delay from 'delay'
import { combineLatest, of } from 'rxjs'
import sinon, { SinonSpy } from 'sinon'

import { requestGraphQL } from '../../backend/graphql'
import type { FeatureFlagName } from '../featureFlags'

import { setFeatureFlagOverride } from './feature-flag-local-overrides'
import { FeatureFlagClient } from './FeatureFlagClient'

describe('FeatureFlagClient', () => {
    const ENABLED_FLAG = 'enabled-flag' as FeatureFlagName
    const DISABLED_FLAG = 'disabled-flag' as FeatureFlagName
    const NON_EXISTING_FLAG = 'non-existing-flag' as FeatureFlagName

    const mockRequestGraphQL = sinon.spy((query, variables) =>
        of({
            data: {
                evaluateFeatureFlag:
                    variables.flagName === ENABLED_FLAG ? true : variables.flagName === DISABLED_FLAG ? false : null,
            },
            errors: [],
        })
    ) as typeof requestGraphQL & SinonSpy

    beforeEach(() => mockRequestGraphQL.resetHistory())

    it('does not make initial API call ', () => {
        new FeatureFlagClient(mockRequestGraphQL)
        sinon.assert.notCalled(mockRequestGraphQL)
    })

    it('returns [true] response from API call for feature flag evaluation', done => {
        const client = new FeatureFlagClient(mockRequestGraphQL)
        expect.assertions(1)

        client.get(ENABLED_FLAG).subscribe(value => {
            expect(value).toBe(true)
            sinon.assert.calledOnce(mockRequestGraphQL)
            done()
        })
    })

    it('returns [false] response from API call for feature flag evaluation', done => {
        const client = new FeatureFlagClient(mockRequestGraphQL)
        expect.assertions(1)

        client.get(DISABLED_FLAG).subscribe(value => {
            expect(value).toBe(false)
            sinon.assert.calledOnce(mockRequestGraphQL)
            done()
        })
    })

    it('returns [defaultValue] correctly', done => {
        const client = new FeatureFlagClient(mockRequestGraphQL)
        expect.assertions(1)

        client.get(NON_EXISTING_FLAG).subscribe(value => {
            expect(value).toBeNull()
            sinon.assert.calledOnce(mockRequestGraphQL)
            done()
        })
    })

    it('makes only single API call per feature flag evaluation', done => {
        const client = new FeatureFlagClient(mockRequestGraphQL)
        expect.assertions(2)

        combineLatest([client.get(ENABLED_FLAG), client.get(ENABLED_FLAG)]).subscribe(([value1, value2]) => {
            expect(value1).toBe(true)
            expect(value2).toBe(true)
            sinon.assert.calledOnce(mockRequestGraphQL)
            done()
        })
    })

    it('makes only single API call per feature flag if cache is still active', async () => {
        const cacheTimeToLive = 10
        const client = new FeatureFlagClient(mockRequestGraphQL, cacheTimeToLive)
        expect.assertions(3)

        const value1 = await client.get(ENABLED_FLAG).toPromise()
        expect(value1).toBe(true)

        await delay(5)

        const value2 = await client.get(ENABLED_FLAG).toPromise()
        expect(value2).toBe(true)
        sinon.assert.calledOnce(mockRequestGraphQL)

        await delay(5)

        const value3 = await client.get(ENABLED_FLAG).toPromise()
        expect(value3).toBe(true)
        sinon.assert.calledTwice(mockRequestGraphQL)
    })

    it('updates on new/different value after cache TTL', async () => {
        let index = -1
        const mockRequestGraphQL = sinon.spy((query, variables) => {
            index++
            return of({
                data: { evaluateFeatureFlag: [ENABLED_FLAG].includes(variables.flagName) && index === 0 },
                errors: [],
            })
        }) as typeof requestGraphQL & SinonSpy

        const cacheTimeToLive = 1
        const client = new FeatureFlagClient(mockRequestGraphQL, cacheTimeToLive)
        expect.assertions(2)

        const value1 = await client.get(ENABLED_FLAG).toPromise()
        expect(value1).toBe(true)

        await delay(cacheTimeToLive)

        const value2 = await client.get(ENABLED_FLAG).toPromise()
        expect(value2).toBe(false)
    })

    describe('local feature flag overrides', () => {
        beforeEach(() => {
            // remove local overrides
            localStorage.clear()
        })
        it('returns [false] override if it exists', done => {
            const client = new FeatureFlagClient(mockRequestGraphQL)
            setFeatureFlagOverride(ENABLED_FLAG, false)
            expect.assertions(1)

            client.get(ENABLED_FLAG).subscribe(value => {
                expect(value).toBe(false)
                sinon.assert.notCalled(mockRequestGraphQL)
                done()
            })
        })

        it('returns [true] override if it exists', done => {
            const client = new FeatureFlagClient(mockRequestGraphQL)
            setFeatureFlagOverride(DISABLED_FLAG, true)
            expect.assertions(1)

            client.get(DISABLED_FLAG).subscribe(value => {
                expect(value).toBe(true)
                sinon.assert.notCalled(mockRequestGraphQL)
                done()
            })
        })

        it('does not use non-boolean override', done => {
            const client = new FeatureFlagClient(mockRequestGraphQL)
            // eslint-disable-next-line @typescript-eslint/ban-ts-comment
            // @ts-ignore
            setFeatureFlagOverride(DISABLED_FLAG, 'something else')
            expect.assertions(1)

            client.get(DISABLED_FLAG).subscribe(value => {
                expect(value).toBe(false)
                sinon.assert.calledOnce(mockRequestGraphQL)
                done()
            })
        })
    })
})
