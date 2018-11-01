import { assert } from 'chai'
import { cloneDeep } from 'lodash-es'
import { createAggregateError, ErrorLike, isErrorLike } from './errors'
import {
    ConfigurationSubject,
    CustomMergeFunctions,
    gqlToCascade,
    merge,
    mergeSettings,
    Settings,
    SubjectConfigurationContents,
} from './settings'

const FIXTURE_ORG: ConfigurationSubject & SubjectConfigurationContents = {
    __typename: 'Org',
    name: 'n',
    displayName: 'n',
    id: 'a',
    settingsURL: 'u',
    viewerCanAdminister: true,
    latestSettings: { configuration: { contents: '{"a":1}' } },
}

const FIXTURE_USER: ConfigurationSubject & SubjectConfigurationContents = {
    __typename: 'User',
    username: 'n',
    displayName: 'n',
    id: 'b',
    settingsURL: 'u',
    viewerCanAdminister: true,
    latestSettings: { configuration: { contents: '{"b":2}' } },
}

const FIXTURE_USER_WITH_SETTINGS_ERROR: ConfigurationSubject & SubjectConfigurationContents = {
    ...FIXTURE_USER,
    id: 'c',
    latestSettings: { configuration: { contents: '.' } },
}

const SETTINGS_ERROR_FOR_FIXTURE_USER = createAggregateError([
    new Error('Configuration parse error, code: 0 (offset: 0, length: 1)'),
])

describe('gqlToCascade', () => {
    it('converts a value', () =>
        assert.deepEqual(
            gqlToCascade({
                subjects: [FIXTURE_ORG, FIXTURE_USER],
            }),
            {
                subjects: [{ subject: FIXTURE_ORG, settings: { a: 1 } }, { subject: FIXTURE_USER, settings: { b: 2 } }],
                merged: { a: 1, b: 2 },
            }
        ))
    it('preserves errors', () => {
        const value = gqlToCascade({
            subjects: [FIXTURE_ORG, FIXTURE_USER_WITH_SETTINGS_ERROR, FIXTURE_USER],
        })
        assert.strictEqual(isErrorLike(value.merged) && value.merged.message, SETTINGS_ERROR_FOR_FIXTURE_USER.message)
        assert.strictEqual(
            value.subjects &&
                !isErrorLike(value.subjects) &&
                isErrorLike(value.subjects[1].settings) &&
                (value.subjects[1].settings as ErrorLike).message,
            SETTINGS_ERROR_FOR_FIXTURE_USER.message
        )
    })
})

describe('mergeSettings', () => {
    it('handles an empty array', () => assert.strictEqual(mergeSettings([]), null))
    it('merges multiple values', () =>
        assert.deepEqual(mergeSettings<{ a?: number; b?: number } & Settings>([{ a: 1 }, { b: 2 }, { a: 3 }]), {
            a: 3,
            b: 2,
        }))
})

describe('merge', () => {
    function assertMerged(base: any, add: any, expected: any, custom?: CustomMergeFunctions): void {
        const origBase = cloneDeep(base)
        merge(base, add, custom)
        assert.deepEqual(
            base,
            expected,
            `merge ${JSON.stringify(origBase)} into ${JSON.stringify(add)}:\ngot:  ${JSON.stringify(
                base
            )}\nwant: ${JSON.stringify(expected)}`
        )
    }

    it('merges with empty', () => {
        assertMerged({ a: 1 }, {}, { a: 1 })
        assertMerged({}, { a: 1 }, { a: 1 })
    })
    it('merges top-level objects deeply', () => assertMerged({ a: 1 }, { b: 2 }, { a: 1, b: 2 }))
    it('merges nested objects deeply', () => assertMerged({ a: { b: 1 } }, { a: { c: 2 } }, { a: { b: 1, c: 2 } }))
    it('overwrites arrays', () => assertMerged({ a: [1] }, { a: [2] }, { a: [2] }))
    it('uses custom merge functions', () =>
        assertMerged({ a: [1] }, { a: [2] }, { a: [1, 2] }, { a: (base, add) => [...base, ...add] }))
})
