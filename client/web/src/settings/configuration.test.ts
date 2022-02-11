import settingsSchemaJSON from '../../../../schema/settings.schema.json'

import { mergeSettingsSchemas } from './configuration'

describe('mergeSettingsSchemas', () => {
    test('handles empty', () =>
        expect(mergeSettingsSchemas([])).toEqual({
            allOf: [{ $ref: settingsSchemaJSON.$id }],
        }))

    test('overwrites additionalProperties and required', () =>
        expect(
            mergeSettingsSchemas([
                {
                    manifest: {
                        contributes: {
                            configuration: { additionalProperties: false, properties: { a: { type: 'string' } } },
                        },
                    },
                },
                {
                    manifest: {
                        contributes: {
                            configuration: { required: ['b'], properties: { b: { type: 'string' } } },
                        },
                    },
                },
            ])
        ).toEqual({
            allOf: [
                { $ref: settingsSchemaJSON.$id },
                { additionalProperties: true, required: [], properties: { a: { type: 'string' } } },
                { additionalProperties: true, required: [], properties: { b: { type: 'string' } } },
            ],
        }))

    test('handles error and null configuration', () =>
        expect(
            mergeSettingsSchemas([
                {
                    manifest: {
                        contributes: {
                            configuration: { additionalProperties: false, properties: { a: { type: 'string' } } },
                        },
                    },
                },
                {
                    manifest: new Error('x'),
                },
                {
                    manifest: null,
                },
                {
                    manifest: {},
                },
                {
                    manifest: { contributes: {} },
                },
            ])
        ).toEqual({
            allOf: [
                { $ref: settingsSchemaJSON.$id },
                { additionalProperties: true, required: [], properties: { a: { type: 'string' } } },
            ],
        }))
})
