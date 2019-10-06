import { Range } from '@sourcegraph/extension-api-classes'
import { WorkspaceEdit } from '../../types/workspaceEdit'

const URL_1 = new URL('file:///1')

describe('WorkspaceEdit', () => {
    test('replace', () => {
        const edits = new WorkspaceEdit()
        edits.replace(URL_1, new Range(1, 2, 3, 4), 'xy')
        expect(Array.from(edits.textEdits())).toEqual([[URL_1, [{ range: new Range(1, 2, 3, 4), newText: 'xy' }]]])
    })
})
