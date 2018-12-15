import assert from 'assert'
import { cloneDeep } from 'lodash'
import { createAggregateError, ErrorLike, isErrorLike } from '../util/errors'
import {
    CustomMergeFunctions,
    gqlToCascade,
    merge,
    mergeSettings,
    Settings,
    SettingsCascade,
    SettingsSubject,
    SubjectSettingsContents,
} from './settings'

const FIXTURE_ORG: SettingsSubject & SubjectSettingsContents = {
    __typename: 'Org',
    name: 'n',
    displayName: 'n',
    id: 'a',
    settingsURL: 'u',
    viewerCanAdminister: true,
    latestSettings: { id: 1, contents: '{"a":1}' },
}

const FIXTURE_USER: SettingsSubject & SubjectSettingsContents = {
    __typename: 'User',
    username: 'n',
    displayName: 'n',
    id: 'b',
    settingsURL: 'u',
    viewerCanAdminister: true,
    latestSettings: { id: 2, contents: '{"b":2}' },
}

const FIXTURE_USER_WITH_SETTINGS_ERROR: SettingsSubject & SubjectSettingsContents = {
    ...FIXTURE_USER,
    id: 'c',
    latestSettings: { id: 3, contents: '.' },
}

const SETTINGS_ERROR_FOR_FIXTURE_USER = createAggregateError([new Error('parse error (code: 0, offset: 0, length: 1)')])

describe('gqlToCascade', () => {
    it('converts a value', () =>
        assert.deepStrictEqual(
            gqlToCascade({
                subjects: [FIXTURE_ORG, FIXTURE_USER],
            }),
            {
                subjects: [
                    { subject: FIXTURE_ORG, settings: { a: 1 }, lastID: 1 },
                    { subject: FIXTURE_USER, settings: { b: 2 }, lastID: 2 },
                ],
                final: { a: 1, b: 2 },
            } as SettingsCascade
        ))
    it('preserves errors', () => {
        const value = gqlToCascade({
            subjects: [FIXTURE_ORG, FIXTURE_USER_WITH_SETTINGS_ERROR, FIXTURE_USER],
        })
        assert.strictEqual(isErrorLike(value.final) && value.final.message, SETTINGS_ERROR_FOR_FIXTURE_USER.message)
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
        assert.deepStrictEqual(mergeSettings<{ a?: number; b?: number } & Settings>([{ a: 1 }, { b: 2 }, { a: 3 }]), {
            a: 3,
            b: 2,
        }))
})

describe('merge', () => {
    function assertMerged(base: any, add: any, expected: any, custom?: CustomMergeFunctions): void {
        const origBase = cloneDeep(base)
        merge(base, add, custom)
        assert.deepStrictEqual(
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
