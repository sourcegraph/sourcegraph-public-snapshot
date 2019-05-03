import { Selection } from '@sourcegraph/extension-api-classes'
import { ExtDocuments } from './documents'
import { ExtWindow, ExtWindows } from './windows'

describe('ExtWindow', () => {
    const NOOP_PROXY = {} as any

    const DOCUMENTS = new ExtDocuments(async () => void 0)
    DOCUMENTS.$acceptDocumentData([{ uri: 'u', text: 't', languageId: 'l' }])

    test('reuses ExtCodeEditor object when updated', () => {
        const wins = new ExtWindow(NOOP_PROXY, DOCUMENTS, {
            editors: [{ type: 'CodeEditor', editorId: 'editor#0', resource: 'u', isActive: true, selections: [] }],
        })
        const origViewComponent = wins.activeViewComponent
        expect(origViewComponent).toBeTruthy()

        wins.update({
            editors: [
                {
                    type: 'CodeEditor',
                    editorId: 'editor#0',
                    resource: 'u',
                    isActive: true,
                    selections: [new Selection(1, 2, 3, 4)],
                },
            ],
        })
        expect(wins.activeViewComponent).toBe(origViewComponent)
    })

    test('creates new ExtCodeEditor object for a different resource URI', () => {
        const wins = new ExtWindow(NOOP_PROXY, DOCUMENTS, {
            editors: [{ type: 'CodeEditor', editorId: 'editor#0', resource: 'u', isActive: true, selections: [] }],
        })
        const origViewComponent = wins.activeViewComponent
        expect(origViewComponent).toBeTruthy()

        wins.update({
            editors: [{ type: 'CodeEditor', editorId: 'editor#1', resource: 'u2', isActive: true, selections: [] }],
        })
        expect(wins.activeViewComponent).not.toBe(origViewComponent)
    })
})

describe('ExtWindows', () => {
    const NOOP_PROXY = {} as any

    const documents = new ExtDocuments(async () => void 0)
    documents.$acceptDocumentData([{ uri: 'u', text: 't', languageId: 'l' }])

    test('reuses ExtWindow object when updated', () => {
        const wins = new ExtWindows(NOOP_PROXY, documents)
        wins.$acceptWindowData({
            editors: [{ type: 'CodeEditor', editorId: 'editor#0', resource: 'u', isActive: true, selections: [] }],
        })
        const origWin = wins.activeWindow
        expect(origWin).toBeTruthy()

        wins.$acceptWindowData({
            editors: [{ type: 'CodeEditor', editorId: 'editor#0', resource: 'u', isActive: false, selections: [] }],
        })
        expect(wins.activeWindow).toBe(origWin)

        wins.$acceptWindowData({
            editors: [{ type: 'CodeEditor', editorId: 'editor#1', resource: 'u2', isActive: false, selections: [] }],
        })
        expect(wins.activeWindow).toBe(origWin)
    })
})
