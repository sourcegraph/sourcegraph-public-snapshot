/* eslint-disable @typescript-eslint/ban-ts-comment */
import { EMPTY, of } from 'rxjs'
import sinon from 'sinon'

import type { FeatureFlagName } from '../featureFlags'

import { EVALUATE_FEATURE_FLAGS_QUERY, EVALUATE_FEATURE_FLAG_QUERY, FeatureFlagClient } from './FeatureFlagClient'

describe('FeatureFlagClient', () => {
    // @ts-ignore
    const mockRequestGraphQL = sinon.spy((query, variables) => {
        if (query === EVALUATE_FEATURE_FLAGS_QUERY) {
            return of({
                data: {
                    evaluatedFeatureFlags: [],
                },
            })
        }

        if (query === EVALUATE_FEATURE_FLAG_QUERY) {
            return of({ data: { evaluateFeatureFlag: ['enabled-flag'].includes(variables.flagName) } })
        }
        return EMPTY
    })
    beforeEach(() => mockRequestGraphQL.resetHistory())
    it('makes initial API call ', () => {
        // @ts-ignore
        new FeatureFlagClient(mockRequestGraphQL)
        // @ts-ignore
        sinon.assert.calledOnceWithExactly(mockRequestGraphQL, EVALUATE_FEATURE_FLAGS_QUERY)
    })

    it('uses [true] response from API call for feature flag evaluation', done => {
        // @ts-ignore
        const client = new FeatureFlagClient(mockRequestGraphQL)

        client.on('enabled-flag' as FeatureFlagName, value => {
            expect(value).toBe(true)
            // @ts-ignore
            sinon.assert.calledWithExactly(mockRequestGraphQL, EVALUATE_FEATURE_FLAG_QUERY, {
                flagName: 'enabled-flag',
            })
            done()
        })
    })

    it('uses [false] response from API call for feature flag evaluation', done => {
        // @ts-ignore
        const client = new FeatureFlagClient(mockRequestGraphQL)

        client.on('other-flag' as FeatureFlagName, value => {
            expect(value).toBe(false)
            // @ts-ignore
            sinon.assert.calledWithExactly(mockRequestGraphQL, EVALUATE_FEATURE_FLAG_QUERY, {
                flagName: 'other-flag',
            })
            done()
        })
    })
})
