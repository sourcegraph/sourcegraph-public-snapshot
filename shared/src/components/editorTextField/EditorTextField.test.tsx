import { of } from 'rxjs'
import { CodeEditorWithModel } from '../../api/client/services/editorService'
import { EditorTextFieldUtils } from './EditorTextField'

describe('EditorTextFieldUtils', () => {
    describe('getEditorDataFromElement', () => {
        test('empty selection', () => {
            const e = document.createElement('textarea')
            e.value = 'abc'
            e.setSelectionRange(2, 2)
            expect(EditorTextFieldUtils.getEditorDataFromElement(e)).toEqual({
                text: 'abc',
                selections: [
                    {
                        anchor: { line: 0, character: 2 },
                        active: { line: 0, character: 2 },
                        start: { line: 0, character: 2 },
                        end: { line: 0, character: 2 },
                        isReversed: false,
                    },
                ],
            })
        })
        test('forward selection', () => {
            const e = document.createElement('textarea')
            e.value = 'abc'
            e.setSelectionRange(2, 3)
            expect(EditorTextFieldUtils.getEditorDataFromElement(e)).toEqual({
                text: 'abc',
                selections: [
                    {
                        anchor: { line: 0, character: 2 },
                        active: { line: 0, character: 3 },
                        start: { line: 0, character: 2 },
                        end: { line: 0, character: 3 },
                        isReversed: false,
                    },
                ],
            })
        })
        test('backward selection', () => {
            const e = document.createElement('textarea')
            e.value = 'abc'
            e.setSelectionRange(2, 3, 'backward')
            expect(EditorTextFieldUtils.getEditorDataFromElement(e)).toEqual({
                text: 'abc',
                selections: [
                    {
                        anchor: { line: 0, character: 3 },
                        active: { line: 0, character: 2 },
                        start: { line: 0, character: 3 },
                        end: { line: 0, character: 2 },
                        isReversed: true,
                    },
                ],
            })
        })
    })

    test('updateEditorSelectionFromElement', () => {
        const e = document.createElement('textarea')
        e.value = 'abc'
        e.setSelectionRange(2, 3, 'backward')
        const setSelections = jest.fn()
        EditorTextFieldUtils.updateEditorSelectionFromElement({ setSelections }, { editorId: 'e' }, e)
        expect(setSelections.mock.calls.length).toBe(1)
        expect(setSelections.mock.calls[0][0]).toEqual({ editorId: 'e' })
        expect(setSelections.mock.calls[0][1]).toEqual([
            {
                anchor: { line: 0, character: 3 },
                active: { line: 0, character: 2 },
                start: { line: 0, character: 3 },
                end: { line: 0, character: 2 },
                isReversed: true,
            },
        ])
    })

    test('updateModelFromElement', () => {
        const e = document.createElement('textarea')
        e.value = 'abc'
        e.setSelectionRange(2, 3, 'backward')
        const updateModel = jest.fn()
        EditorTextFieldUtils.updateModelFromElement({ updateModel }, 'u', e)
        expect(updateModel.mock.calls.length).toBe(1)
        expect(updateModel.mock.calls[0][0]).toEqual('u')
        expect(updateModel.mock.calls[0][1]).toEqual('abc')
    })

    describe('updateElementOnEditorOrModelChanges', () => {
        test('forward selection', () => {
            const e = document.createElement('textarea')
            e.value = 'abc'
            const setValue = jest.fn()
            const subscription = EditorTextFieldUtils.updateElementOnEditorOrModelChanges(
                {
                    observeEditorAndModel: () =>
                        of<CodeEditorWithModel>({
                            editorId: 'e',
                            type: 'CodeEditor',
                            resource: 'u',
                            model: { uri: 'u', languageId: 'l', text: 'xyz' },
                            selections: [
                                {
                                    anchor: { line: 0, character: 2 },
                                    active: { line: 0, character: 3 },
                                    start: { line: 0, character: 2 },
                                    end: { line: 0, character: 3 },
                                    isReversed: false,
                                },
                            ],
                            isActive: true,
                        }),
                },
                { editorId: 'e' },
                setValue,
                { current: e }
            )
            expect(setValue.mock.calls.length).toBe(1)
            expect(setValue.mock.calls[0][0]).toEqual('xyz')
            expect(e.selectionStart).toBe(2)
            expect(e.selectionEnd).toBe(3)
            expect(e.selectionDirection).toBe('forward')
            subscription.unsubscribe()
        })
        test('backward selection', () => {
            const e = document.createElement('textarea')
            e.value = 'abc'
            const setValue = jest.fn()
            const subscription = EditorTextFieldUtils.updateElementOnEditorOrModelChanges(
                {
                    observeEditorAndModel: () =>
                        of<CodeEditorWithModel>({
                            editorId: 'e',
                            type: 'CodeEditor',
                            resource: 'u',
                            model: { uri: 'u', languageId: 'l', text: 'xyz' },
                            selections: [
                                {
                                    anchor: { line: 0, character: 3 },
                                    active: { line: 0, character: 2 },
                                    start: { line: 0, character: 3 },
                                    end: { line: 0, character: 2 },
                                    isReversed: true,
                                },
                            ],
                            isActive: true,
                        }),
                },
                { editorId: 'e' },
                setValue,
                { current: e }
            )
            expect(e.selectionStart).toBe(2)
            expect(e.selectionEnd).toBe(3)
            expect(e.selectionDirection).toBe('backward')
            subscription.unsubscribe()
        })
    })
})
