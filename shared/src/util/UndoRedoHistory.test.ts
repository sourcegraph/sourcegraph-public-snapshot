import { UndoRedoHistory } from './UndoRedoHistory'

describe('UndoRedoHistory', () => {
    test('undo()', () => {
        const history = new UndoRedoHistory<string>({
            current: 'undone',
            onChange: () => null,
        }).push('')
        expect(history.current).toBe('')
        history.undo()
        expect(history.current).toBe('undone')
    })

    test('redo()', () => {
        const history = new UndoRedoHistory<string>({
            current: '',
            onChange: () => null,
        })
            .push('redone')
            .undo()
        expect(history.current).toBe('')
        history.redo()
        expect(history.current).toBe('redone')
    })

    test('onUpdate()', () => {
        new UndoRedoHistory<string>({
            current: 'undone',
            onChange: value => expect(value).toBe('undone'),
        })
            .push('')
            .undo()
    })

    it('maintains the correct amount of items in history (historyLength prop)', () => {
        const history = new UndoRedoHistory<string>({
            current: 'a',
            historyLength: 2,
            onChange: () => null,
        })
            .push('b')
            .push('c')
        expect(history.current).toBe('c')
        history.undo().undo()
        expect(history.current).toBe('b')
    })
})
