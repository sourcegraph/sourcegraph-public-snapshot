import { parseUrlOverrideFeatureFlags } from './parseUrlOverrideFeatureFlags'

describe('parseUrlOverrideFeatureFlags', () => {
    it('parses single key', () => {
        expect(parseUrlOverrideFeatureFlags('')).toBeUndefined()
        expect(parseUrlOverrideFeatureFlags('?feature-flag-key=foo')).toEqual({ foo: undefined })
        expect(parseUrlOverrideFeatureFlags('feature-flag-key=foo&feature-flag-value=false')).toEqual({ foo: 'false' })
    })

    it('parses multiple keys', () => {
        expect(parseUrlOverrideFeatureFlags('?feature-flag-key=foo,bar')).toEqual({ foo: undefined, bar: undefined })

        expect(parseUrlOverrideFeatureFlags('feature-flag-key=foo,bar&feature-flag-value=,true')).toEqual({
            foo: undefined,
            bar: 'true',
        })

        expect(parseUrlOverrideFeatureFlags('feature-flag-key=foo,bar&feature-flag-value=false,')).toEqual({
            foo: 'false',
            bar: undefined,
        })

        expect(parseUrlOverrideFeatureFlags('feature-flag-key=foo,bar&feature-flag-value=false,true')).toEqual({
            foo: 'false',
            bar: 'true',
        })
    })
})
