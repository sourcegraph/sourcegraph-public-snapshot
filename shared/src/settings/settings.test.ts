import { createAggregateError, isErrorLike } from '../util/errors'
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
    test('converts a value', () => {
        const expected: SettingsCascade = {
            subjects: [
                { subject: FIXTURE_ORG, settings: { a: 1 }, lastID: 1 },
                { subject: FIXTURE_USER, settings: { b: 2 }, lastID: 2 },
            ],
            final: { a: 1, b: 2 },
        }
        expect(
            gqlToCascade({
                subjects: [FIXTURE_ORG, FIXTURE_USER],
            })
        ).toEqual(expected)
    })
    test('preserves errors', () => {
        const value = gqlToCascade({
            subjects: [FIXTURE_ORG, FIXTURE_USER_WITH_SETTINGS_ERROR, FIXTURE_USER],
        })
        expect(isErrorLike(value.final) && value.final.message).toBe(SETTINGS_ERROR_FOR_FIXTURE_USER.message)
        expect(
            value.subjects &&
                !isErrorLike(value.subjects) &&
                isErrorLike(value.subjects[1].settings) &&
                value.subjects[1].settings.message
        ).toBe(SETTINGS_ERROR_FOR_FIXTURE_USER.message)
    })
})

describe('mergeSettings', () => {
    test('handles an empty array', () => expect(mergeSettings([])).toBe(null))
    test('merges multiple values', () =>
        expect(mergeSettings<{ a?: number; b?: number } & Settings>([{ a: 1 }, { b: 2 }, { a: 3 }])).toEqual({
            a: 3,
            b: 2,
        }))
})

describe('merge', () => {
    function assertMerged(base: any, add: any, expected: any, custom?: CustomMergeFunctions): void {
        merge(base, add, custom)
        expect(base).toEqual(expected)
    }

    test('merges with empty', () => {
        assertMerged({ a: 1 }, {}, { a: 1 })
        assertMerged({}, { a: 1 }, { a: 1 })
    })
    test('merges top-level objects deeply', () => assertMerged({ a: 1 }, { b: 2 }, { a: 1, b: 2 }))
    test('merges nested objects deeply', () => assertMerged({ a: { b: 1 } }, { a: { c: 2 } }, { a: { b: 1, c: 2 } }))
    test('overwrites arrays', () => assertMerged({ a: [1] }, { a: [2] }, { a: [2] }))
    test('uses custom merge functions', () =>
        assertMerged({ a: [1] }, { a: [2] }, { a: [1, 2] }, { a: (base, add) => [...base, ...add] }))
})
