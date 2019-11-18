import { UndoRedoHistory } from './UndoRedoHistory'

describe('UndoRedoHistory', () => {
    test('undo()', () => {
        const history = new UndoRedoHistory<string>({
            current: 'undone',
            onUpdate: () => null,
        })
            .push('')
            .undo()
        expect(history.current).toBe('undone')
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

    test('onUpdate()', () => {
        new UndoRedoHistory<string>({
            current: 'undone',
            onUpdate: value => expect(value).toBe('undone'),
        })
            .push('')
            .undo()
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
