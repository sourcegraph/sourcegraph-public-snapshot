import { modelToTextDocumentPositionParams } from './model'

describe('modelToTextDocumentPositionParams', () => {
    test('null if visibleViewComponents is empty', () => {
        expect(modelToTextDocumentPositionParams({ visibleViewComponents: null })).toBeNull()
        expect(modelToTextDocumentPositionParams({ visibleViewComponents: [] })).toBeNull()
    })

    test('null if no visibleViewComponents are active', () => {
        expect(
            modelToTextDocumentPositionParams({
                visibleViewComponents: [
                    {
                        type: 'CodeEditor',
                        isActive: false,
                        selections: [],
                        item: { uri: 'u', text: 't', languageId: 'l' },
                    },
                ],
            })
        ).toBeNull()
    })

    test('null if active visibleViewComponents has no selection', () => {
        expect(
            modelToTextDocumentPositionParams({
                visibleViewComponents: [
                    {
                        type: 'CodeEditor',
                        isActive: true,
                        selections: [],
                        item: { uri: 'u', text: 't', languageId: 'l' },
                    },
                ],
            })
        ).toBeNull()
    })

    test('null if active visibleViewComponents has empty selection', () => {
        expect(
            modelToTextDocumentPositionParams({
                visibleViewComponents: [
                    {
                        type: 'CodeEditor',
                        isActive: true,
                        selections: [
                            {
                                start: { line: 3, character: -1 },
                                end: { line: 3, character: -1 },
                                anchor: { line: 3, character: -1 },
                                active: { line: 3, character: -1 },
                                isReversed: false,
                            },
                        ],
                        item: { uri: 'u', text: 't', languageId: 'l' },
                    },
                ],
            })
        ).toBeNull()
    })

    test('equivalent params', () => {
        expect(
            modelToTextDocumentPositionParams({
                visibleViewComponents: [
                    {
                        type: 'CodeEditor',
                        isActive: true,
                        selections: [
                            {
                                start: { line: 3, character: 2 },
                                end: { line: 3, character: 5 },
                                anchor: { line: 3, character: 2 },
                                active: { line: 3, character: 5 },
                                isReversed: false,
                            },
                        ],
                        item: { uri: 'u', text: 't', languageId: 'l' },
                    },
                ],
            })
        ).toEqual({ textDocument: { uri: 'u', text: 't', languageId: 'l' }, position: { line: 3, character: 2 } })
    })
})
