import { Range } from '@sourcegraph/extension-api-classes'
import { WorkspaceEdit } from './workspaceEdit'

const URL_1 = new URL('file:///1')
const URL_2 = new URL('file:///2')

describe('WorkspaceEdit', () => {
    test('replace', () => {
        const edits = new WorkspaceEdit()
        edits.replace(URL_1, new Range(1, 2, 3, 4), 'xy')
        expect(Array.from(edits.textEdits())).toEqual([[URL_1, [{ range: new Range(1, 2, 3, 4), newText: 'xy' }]]])
    })

    test('toJSON/fromJSON', () => {
        const edits = new WorkspaceEdit()
        edits.replace(URL_1, new Range(1, 2, 3, 4), 'xy')
        edits.renameFile(URL_1, URL_2, { ignoreIfExists: true })
        edits.createFile(URL_1, { overwrite: true })
        expect(WorkspaceEdit.fromJSON(edits.toJSON())).toEqual(edits)
    })
})
