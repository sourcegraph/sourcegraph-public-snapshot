import { of, Subscribable } from 'rxjs'
import { CodeEditorData, getActiveCodeEditorPosition, ReadonlyEditorService } from './editorService'

export function createTestEditorService(
    editors: Subscribable<readonly CodeEditorData[]> = of([])
): ReadonlyEditorService {
    return { editors }
}

describe('getActiveCodeEditorPosition', () => {
    test('null if code editor is empty', () => {
        expect(getActiveCodeEditorPosition([])).toBeNull()
    })

    test('null if no code editors are active', () => {
        expect(
            getActiveCodeEditorPosition([
                {
                    type: 'CodeEditor',
                    isActive: false,
                    selections: [],
                    item: { uri: 'u', text: 't', languageId: 'l' },
                },
            ])
        ).toBeNull()
    })

    test('null if active code editor has no selection', () => {
        expect(
            getActiveCodeEditorPosition([
                {
                    type: 'CodeEditor',
                    isActive: true,
                    selections: [],
                    item: { uri: 'u', text: 't', languageId: 'l' },
                },
            ])
        ).toBeNull()
    })

    test('null if active code editor has empty selection', () => {
        expect(
            getActiveCodeEditorPosition([
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
            ])
        ).toBeNull()
    })

    test('equivalent params', () => {
        expect(
            getActiveCodeEditorPosition([
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
            ])
        ).toEqual({ textDocument: { uri: 'u', text: 't', languageId: 'l' }, position: { line: 3, character: 2 } })
    })
})
