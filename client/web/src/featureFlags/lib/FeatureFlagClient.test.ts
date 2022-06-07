import { combineLatest, of } from 'rxjs'
import sinon, { SinonSpy } from 'sinon'

import { requestGraphQL } from '../../backend/graphql'
import type { FeatureFlagName } from '../featureFlags'

import { setFeatureFlagOverride } from './feature-flag-local-overrides'
import { FeatureFlagClient } from './FeatureFlagClient'

describe('FeatureFlagClient', () => {
    const ENABLED_FLAG = 'enabled-flag' as FeatureFlagName
    const DISABLED_FLAG = 'disabled-flag' as FeatureFlagName

    const mockRequestGraphQL = sinon.spy(() =>
        of({
            data: {
                viewerFeatureFlags: [
                    {
                        name: ENABLED_FLAG,
                        value: true,
                    },
                    {
                        name: DISABLED_FLAG,
                        value: false,
                    },
                ],
            },
            errors: [],
        })
    ) as typeof requestGraphQL & SinonSpy

    beforeEach(() => mockRequestGraphQL.resetHistory())

    it('makes initial API call', () => {
        new FeatureFlagClient(mockRequestGraphQL)
        sinon.assert.calledOnce(mockRequestGraphQL)
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

        client.get(DISABLED_FLAG).subscribe({
            next: value => {
                expect(value).toBe(false)
                sinon.assert.calledOnce(mockRequestGraphQL)
                done()
            },
            complete: () => {
                throw new Error('Should not complete when passing refetch interval')
            },
        })
    })

    it('makes only single API call per feature flag evaluation', done => {
        const client = new FeatureFlagClient(mockRequestGraphQL)
        expect.assertions(2)

        combineLatest([client.get(ENABLED_FLAG), client.get(DISABLED_FLAG)]).subscribe(([value1, value2]) => {
            expect(value1).toBe(true)
            expect(value2).toBe(false)
            sinon.assert.calledOnce(mockRequestGraphQL)
            done()
        })
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
                sinon.assert.calledOnce(mockRequestGraphQL)
                done()
            })
        })

        it('returns [true] override if it exists', done => {
            const client = new FeatureFlagClient(mockRequestGraphQL)
            setFeatureFlagOverride(DISABLED_FLAG, true)
            expect.assertions(1)

            client.get(DISABLED_FLAG).subscribe(value => {
                expect(value).toBe(true)
                sinon.assert.calledOnce(mockRequestGraphQL)
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
