import { Range } from '@sourcegraph/extension-api-classes'
import { TextEdit } from './textEdit'

describe('TextEdit', () => {
    test('replace', () => {
        const edit = TextEdit.replace(new Range(1, 2, 3, 4), 'xy')
        expect(edit.toJSON()).toEqual({ range: new Range(1, 2, 3, 4), newText: 'xy' })
    })

    test('toJSON/fromJSON', () => {
        const edit = TextEdit.replace(new Range(1, 2, 3, 4), 'xy')
        expect(TextEdit.fromJSON(edit.toJSON())).toEqual(edit)
    })
})
