import { Unsubscribable } from 'rxjs'

interface Props<T> {
    /**
     * Callback that receives value from undo/redo
     */
    onChange(currentValue: T): void
    /**
     * Current value that the history should start with
     */
    current: T
    /**
     * Max number of items that should be stored in each history (past, future).
     * E.g: If given 10, then past + future + current = 21
     */
    historyLength?: number
}

/**
 * Provides generic undo/redo capabilities
 */
export class UndoRedoHistory<T> implements Unsubscribable {
    public current: T
    /**
     * Can be undefined after unsubscribe
     */
    private onChange?: Props<T>['onChange']

    private past: T[] = []
    private future: T[] = []
    private historyLength = 50

    constructor(props: Props<T>) {
        this.onChange = value => props.onChange(value)
        this.current = props.current
        this.historyLength = props.historyLength ?? this.historyLength
    }

    /**
     * Generic function to append items to this.past or this.future
     */
    private addToHistoryArray = (history: T[], value: T): T[] =>
        history.length === this.historyLength ? history.slice(1).concat(value) : history.concat(value)

    /**
     * (DRY) Function that does the actual undo/redo.
     * if action === 'undo':
     *     push current to future
     *     take from past, set to current
     * if action === 'redo':
     *     push current to past
     *     take from future, set to current
     */
    private moveHistory(action: 'undo' | 'redo'): void {
        const from = action === 'undo' ? 'past' : 'future'
        const to = action === 'undo' ? 'future' : 'past'
        const fromValue = this[from]
        const toValue = this[to]

        // No items to take, exit
        if (fromValue.length === 0) {
            return
        }

        this[from] = fromValue.slice(0, fromValue.length - 1)
        this[to] = this.addToHistoryArray(toValue, this.current)
        this.current = fromValue[fromValue.length - 1]

        if (this.onChange) {
            this.onChange(this.current)
        }
    }

    // All public methods should return `this` for method chaining

    /**
     * Can be used by components and rxjs.Subscription to unsubscribe on unmount
     */
    public unsubscribe(): this {
        this.onChange = undefined
        return this
    }

    /**
     * Standard undo/redo behavior: once a new value is pushed the future
     * history is erased, and will be populated on the next `undo()`
     */
    public push(value: T): this {
        this.past = this.addToHistoryArray(this.past, this.current)
        this.current = value
        this.future = []
        return this
    }

    public undo(): this {
        this.moveHistory('undo')
        return this
    }

    public redo(): this {
        this.moveHistory('redo')
        return this
    }
}
