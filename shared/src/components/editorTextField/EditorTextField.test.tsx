import { of } from 'rxjs'
import { ViewerId, CodeEditorData } from '../../api/client/services/viewerService'
import { EditorTextFieldUtils } from './EditorTextField'
import { Selection } from '@sourcegraph/extension-api-types'
import * as sinon from 'sinon'
import { noop } from 'lodash'
import { TextModel } from '../../api/client/services/modelService'

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
        const setSelections = sinon.spy<(editor: ViewerId, selections: Selection[]) => void>(noop)
        EditorTextFieldUtils.updateEditorSelectionFromElement({ setSelections }, { viewerId: 'e' }, e)
        sinon.assert.calledOnce(setSelections)
        expect(setSelections.args[0][0]).toEqual({ viewerId: 'e' })
        expect(setSelections.args[0][1]).toEqual([
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
        const updateModel = sinon.spy<(uri: string, text: string) => void>(noop)
        EditorTextFieldUtils.updateModelFromElement({ updateModel }, 'u', e)
        sinon.assert.calledOnce(updateModel)
        expect(updateModel.args[0][0]).toEqual('u')
        expect(updateModel.args[0][1]).toEqual('abc')
    })

    describe('updateElementOnEditorOrModelChanges', () => {
        test('forward selection', () => {
            const e = document.createElement('textarea')
            e.value = 'abc'
            const setValue = sinon.spy<(value: string) => void>(noop)
            const subscription = EditorTextFieldUtils.updateElementOnEditorOrModelChanges(
                {
                    observeViewer: () =>
                        of<CodeEditorData>({
                            type: 'CodeEditor',
                            resource: 'u',
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
                {
                    observeModel: () => of<TextModel>({ uri: 'u', languageId: 'l', text: 'xyz' }),
                },
                { viewerId: 'e' },
                setValue,
                { current: e }
            )
            sinon.assert.calledOnce(setValue)
            expect(setValue.args[0][0]).toEqual('xyz')
            expect(e.selectionStart).toBe(2)
            expect(e.selectionEnd).toBe(3)
            expect(e.selectionDirection).toBe('forward')
            subscription.unsubscribe()
        })
        test('backward selection', () => {
            const e = document.createElement('textarea')
            e.value = 'abc'
            const setValue = sinon.spy<(value: string) => void>(noop)
            const subscription = EditorTextFieldUtils.updateElementOnEditorOrModelChanges(
                {
                    observeViewer: () =>
                        of<CodeEditorData>({
                            type: 'CodeEditor',
                            resource: 'u',
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
                {
                    observeModel: () => of<TextModel>({ uri: 'u', languageId: 'l', text: 'xyz' }),
                },
                { viewerId: 'e' },
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
