import { UndoRedoHistory } from './UndoRedoHistory'

describe('UndoRedoHistory', () => {
    test('undo()', () => {
        new UndoRedoHistory<string>({
            current: 'undone',
            onUpdate: value => expect(value).toBe('undone'),
        })
            .push('')
            .undo()
    })

    test('redo()', () => {
        const history = new UndoRedoHistory<string>({
            current: '',
            onUpdate: () => null,
        })
            .push('redone')
            .undo()
            .redo()
        expect(history.current).toBe('redone')
    })

    it('maintains the correct amount of items in history (historyLength prop)', () => {
        const history = new UndoRedoHistory<string>({
            current: 'a',
            historyLength: 1,
            onUpdate: () => null,
        })
            .push('b')
            .push('c')
            .undo()
            .undo()
        expect(history.current).toBe('b')
    })
})
