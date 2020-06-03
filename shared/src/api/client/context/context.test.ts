import { Selection } from '@sourcegraph/extension-api-types'
import { EMPTY_SETTINGS_CASCADE, SettingsCascadeOrError } from '../../../settings/settings'
import { CodeEditorWithPartialModel } from '../services/viewerService'
import { getComputedContextProperty } from './context'

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
        expect(getComputedContextProperty(undefined, settings, {}, 'config.a')).toBe(1)
        expect(getComputedContextProperty(undefined, settings, {}, 'config.a.b')).toBe(2)
        expect(getComputedContextProperty(undefined, settings, {}, 'config.c.d')).toBe(3)
        expect(getComputedContextProperty(undefined, settings, {}, 'config.x')).toBe(null)
    })

    describe('with code editors', () => {
        const editor: CodeEditorWithPartialModel = {
            viewerId: 'editor2',
            type: 'CodeEditor',
            resource: 'file:///a/b.c',
            model: { languageId: 'l' },
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
        }

        describe('resource', () => {
            test('provides resource.uri', () =>
                expect(getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, 'resource.uri')).toBe(
                    'file:///a/b.c'
                ))
            test('provides resource.basename', () =>
                expect(getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, 'resource.basename')).toBe('b.c'))
            test('provides resource.dirname', () =>
                expect(getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, 'resource.dirname')).toBe(
                    'file:///a'
                ))
            test('provides resource.extname', () =>
                expect(getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, 'resource.extname')).toBe('.c'))
            test('provides resource.language', () =>
                expect(getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, 'resource.language')).toBe('l'))
            test('provides resource.type', () =>
                expect(getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, 'resource.type')).toBe(
                    'textDocument'
                ))

            test('returns null when there are no code editors', () =>
                expect(getComputedContextProperty(undefined, EMPTY_SETTINGS_CASCADE, {}, 'resource.uri')).toBe(null))
        })

        describe('component', () => {
            test('provides component.type', () =>
                expect(getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, 'component.type')).toBe(
                    'CodeEditor'
                ))

            test('returns null when there are no code editors', () =>
                expect(getComputedContextProperty(undefined, EMPTY_SETTINGS_CASCADE, {}, 'component.type')).toBe(null))

            function assertSelection(
                editor: CodeEditorWithPartialModel,
                expression: string,
                expected: Selection
            ): void {
                expect(getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, expression)).toEqual(expected)
                expect(getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, `${expression}.start`)).toEqual(
                    expected.start
                )
                expect(getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, `${expression}.end`)).toEqual(
                    expected.end
                )
                expect(getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, `${expression}.start.line`)).toBe(
                    expected.start.line
                )
                expect(
                    getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, `${expression}.start.character`)
                ).toBe(expected.start.character)
                expect(getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, `${expression}.end.line`)).toBe(
                    expected.end.line
                )
                expect(
                    getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, `${expression}.end.character`)
                ).toBe(expected.end.character)
            }

            test('provides primary selection', () =>
                assertSelection(editor, 'component.selection', {
                    start: { line: 1, character: 2 },
                    end: { line: 3, character: 4 },
                    anchor: { line: 1, character: 2 },
                    active: { line: 3, character: 4 },
                    isReversed: false,
                }))

            test('provides selections', () =>
                expect(getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, 'component.selections')).toEqual([
                    {
                        start: { line: 1, character: 2 },
                        end: { line: 3, character: 4 },
                        anchor: { line: 1, character: 2 },
                        active: { line: 3, character: 4 },
                        isReversed: false,
                    },
                ]))

            function assertNoSelection(editor?: CodeEditorWithPartialModel): void {
                expect(getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, 'component.selection')).toBe(null)
                expect(
                    getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, 'component.selection.start')
                ).toBe(null)
                expect(getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, 'component.selection.end')).toBe(
                    null
                )
                expect(
                    getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, 'component.selection.start.line')
                ).toBe(null)
                expect(
                    getComputedContextProperty(
                        editor,
                        EMPTY_SETTINGS_CASCADE,
                        {},
                        'component.selection.start.character'
                    )
                ).toBe(null)
                expect(
                    getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, 'component.selection.end.line')
                ).toBe(null)
                expect(
                    getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, 'component.selection.end.character')
                ).toBe(null)
            }

            test('returns null when there is no selection', () => {
                assertNoSelection({
                    viewerId: 'editor1',
                    type: 'CodeEditor' as const,
                    resource: 'file:///a/b.c',
                    model: { languageId: 'l' },
                    selections: [],
                    isActive: true,
                })
            })

            test('returns null when there is no component', () => {
                assertNoSelection()
            })

            test('returns undefined for out-of-bounds selection', () =>
                expect(
                    getComputedContextProperty(editor, EMPTY_SETTINGS_CASCADE, {}, 'get(component.selections, 1)')
                ).toBe(undefined))
        })
    })

    describe('panel', () => {
        test('provides panel.activeView.id', () =>
            expect(
                getComputedContextProperty(undefined, EMPTY_SETTINGS_CASCADE, {}, 'panel.activeView.id', {
                    type: 'panelView',
                    id: 'x',
                    hasLocations: true,
                })
            ).toBe('x'))

        test('provides panel.activeView.hasLocations', () =>
            expect(
                getComputedContextProperty(undefined, EMPTY_SETTINGS_CASCADE, {}, 'panel.activeView.hasLocations', {
                    type: 'panelView',
                    id: 'x',
                    hasLocations: true,
                })
            ).toBe(true))

        test('returns null for panel.activeView.id when there is no panel', () =>
            expect(getComputedContextProperty(undefined, EMPTY_SETTINGS_CASCADE, {}, 'panel.activeView.id')).toBe(null))
    })

    test('falls back to the context entries', () => {
        expect(getComputedContextProperty(undefined, EMPTY_SETTINGS_CASCADE, { x: 1 }, 'x')).toBe(1)
        expect(getComputedContextProperty(undefined, EMPTY_SETTINGS_CASCADE, {}, 'y')).toBe(undefined)
    })
})
