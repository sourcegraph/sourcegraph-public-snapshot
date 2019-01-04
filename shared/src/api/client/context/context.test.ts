import { Selection } from '@sourcegraph/extension-api-types'
import { EMPTY_SETTINGS_CASCADE, SettingsCascadeOrError } from '../../../settings/settings'
import { EMPTY_MODEL, Model } from '../model'
import { applyContextUpdate, Context, getComputedContextProperty } from './context'

describe('applyContextUpdate', () => {
    test('merges properties', () =>
        expect(applyContextUpdate({ a: 1, b: null, c: 2, d: 3, e: null }, { a: null, b: 1, c: 3 })).toEqual({
            b: 1,
            c: 3,
            d: 3,
            e: null,
        } as Context))
})

describe('getComputedContextProperty', () => {
    test('provides config', () => {
        const settings: SettingsCascadeOrError = {
            final: {
                a: 1,
                'a.b': 2,
                'c.d': 3,
            },
            subjects: [],
        }
        expect(getComputedContextProperty(EMPTY_MODEL, settings, {}, 'config.a')).toBe(1)
        expect(getComputedContextProperty(EMPTY_MODEL, settings, {}, 'config.a.b')).toBe(2)
        expect(getComputedContextProperty(EMPTY_MODEL, settings, {}, 'config.c.d')).toBe(3)
        expect(getComputedContextProperty(EMPTY_MODEL, settings, {}, 'config.x')).toBe(null)
    })

    describe('model with component', () => {
        const model: Model = {
            ...EMPTY_MODEL,
            visibleViewComponents: [
                {
                    type: 'textEditor',
                    item: {
                        uri: 'file:///inactive',
                        languageId: 'inactive',
                        text: 'inactive',
                    },
                    selections: [
                        {
                            start: { line: 11, character: 22 },
                            end: { line: 33, character: 44 },
                            anchor: { line: 11, character: 22 },
                            active: { line: 33, character: 44 },
                            isReversed: false,
                        },
                    ],
                    isActive: false,
                },
                {
                    type: 'textEditor',
                    item: {
                        uri: 'file:///a/b.c',
                        languageId: 'l',
                        text: 't',
                    },
                    selections: [
                        {
                            start: { line: 1, character: 2 },
                            end: { line: 3, character: 4 },
                            anchor: { line: 1, character: 2 },
                            active: { line: 3, character: 4 },
                            isReversed: false,
                        },
                    ],
                    isActive: true,
                },
            ],
        }

        describe('resource', () => {
            test('provides resource.uri', () =>
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, 'resource.uri')).toBe(
                    'file:///a/b.c'
                ))
            test('provides resource.basename', () =>
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, 'resource.basename')).toBe('b.c'))
            test('provides resource.dirname', () =>
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, 'resource.dirname')).toBe(
                    'file:///a'
                ))
            test('provides resource.extname', () =>
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, 'resource.extname')).toBe('.c'))
            test('provides resource.language', () =>
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, 'resource.language')).toBe('l'))
            test('provides resource.type', () =>
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, 'resource.type')).toBe(
                    'textDocument'
                ))

            test('returns null when the model has no component', () =>
                expect(getComputedContextProperty(EMPTY_MODEL, EMPTY_SETTINGS_CASCADE, {}, 'resource.uri')).toBe(null))
        })

        describe('component', () => {
            test('provides component.type', () =>
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, 'component.type')).toBe(
                    'textEditor'
                ))

            test('returns null when the model has no component', () =>
                expect(getComputedContextProperty(EMPTY_MODEL, EMPTY_SETTINGS_CASCADE, {}, 'component.type')).toBe(
                    null
                ))

            function assertSelection(model: Model, expr: string, expected: Selection): void {
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, expr)).toEqual(expected)
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, `${expr}.start`)).toEqual(
                    expected.start
                )
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, `${expr}.end`)).toEqual(
                    expected.end
                )
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, `${expr}.start.line`)).toBe(
                    expected.start.line
                )
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, `${expr}.start.character`)).toBe(
                    expected.start.character
                )
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, `${expr}.end.line`)).toBe(
                    expected.end.line
                )
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, `${expr}.end.character`)).toBe(
                    expected.end.character
                )
            }

            test('provides primary selection', () =>
                assertSelection(model, 'component.selection', {
                    start: { line: 1, character: 2 },
                    end: { line: 3, character: 4 },
                    anchor: { line: 1, character: 2 },
                    active: { line: 3, character: 4 },
                    isReversed: false,
                }))

            test('provides selections', () =>
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, 'component.selections')).toEqual([
                    {
                        start: { line: 1, character: 2 },
                        end: { line: 3, character: 4 },
                        anchor: { line: 1, character: 2 },
                        active: { line: 3, character: 4 },
                        isReversed: false,
                    },
                ]))

            function assertNoSelection(model: Model): void {
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, 'component.selection')).toBe(null)
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, 'component.selection.start')).toBe(
                    null
                )
                expect(getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, 'component.selection.end')).toBe(
                    null
                )
                expect(
                    getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, 'component.selection.start.line')
                ).toBe(null)
                expect(
                    getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, 'component.selection.start.character')
                ).toBe(null)
                expect(
                    getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, 'component.selection.end.line')
                ).toBe(null)
                expect(
                    getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, 'component.selection.end.character')
                ).toBe(null)
            }

            test('returns null when there is no selection', () => {
                assertNoSelection({
                    ...EMPTY_MODEL,
                    visibleViewComponents: [
                        {
                            type: 'textEditor',
                            item: {
                                uri: 'file:///a/b.c',
                                languageId: 'l',
                                text: 't',
                            },
                            selections: [],
                            isActive: true,
                        },
                    ],
                })
            })

            test('returns null when there is no component', () => {
                assertNoSelection({
                    ...EMPTY_MODEL,
                    visibleViewComponents: [],
                })
            })

            test('returns undefined for out-of-bounds selection', () =>
                expect(
                    getComputedContextProperty(model, EMPTY_SETTINGS_CASCADE, {}, 'get(component.selections, 1)')
                ).toBe(undefined))
        })
    })

    describe('panel', () => {
        test('provides panel.activeView.id', () =>
            expect(
                getComputedContextProperty(EMPTY_MODEL, EMPTY_SETTINGS_CASCADE, {}, 'panel.activeView.id', {
                    type: 'panelView',
                    id: 'x',
                })
            ).toBe('x'))

        test('returns null for panel.activeView.id when there is no panel', () =>
            expect(getComputedContextProperty(EMPTY_MODEL, EMPTY_SETTINGS_CASCADE, {}, 'panel.activeView.id')).toBe(
                null
            ))
    })

    describe('location', () => {
        test('scoped context shadows outer context', () =>
            expect(
                getComputedContextProperty(EMPTY_MODEL, EMPTY_SETTINGS_CASCADE, { a: 1 }, 'a', {
                    type: 'location',
                    location: { uri: 'x', context: { a: 2 } },
                })
            ).toBe(2))

        test('provides location.uri', () =>
            expect(
                getComputedContextProperty(EMPTY_MODEL, EMPTY_SETTINGS_CASCADE, {}, 'location.uri', {
                    type: 'location',
                    location: { uri: 'x' },
                })
            ).toBe('x'))

        test('returns null for location.uri when there is no location', () =>
            expect(getComputedContextProperty(EMPTY_MODEL, EMPTY_SETTINGS_CASCADE, {}, 'location.uri')).toBe(null))
    })

    test('falls back to the context entries', () => {
        expect(getComputedContextProperty(EMPTY_MODEL, EMPTY_SETTINGS_CASCADE, { x: 1 }, 'x')).toBe(1)
        expect(getComputedContextProperty(EMPTY_MODEL, EMPTY_SETTINGS_CASCADE, {}, 'y')).toBe(undefined)
    })
})
