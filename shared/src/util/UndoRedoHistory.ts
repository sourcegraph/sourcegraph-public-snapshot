import { Unsubscribable } from 'rxjs'

interface Props<ItemType> {
    /**
     * Callback that receives current value on undo/redo
     */
    onChange(currentValue: ItemType): void
    /**
     * Current value that the history should start with
     */
    current: ItemType
    /**
     * Maximum number of items to be stored in history
     */
    historyLength?: number
}

/**
 * Provides generic undo/redo capabilities
 */
export class UndoRedoHistory<ItemType> implements Unsubscribable {
    /**
     * See Props['current']
     */
    public get current(): ItemType {
        return this.history[this.currentIndex]
    }
    /**
     * See Props['onChange'].
     * Can be undefined after unsubscribe
     */
    private onChange?: Props<ItemType>['onChange']
    /**
     * Array containing current and pushed values
     */
    private history: ItemType[]
    /**
     * Cursor defining which index is the current value in history
     */
    private currentIndex = 0
    /**
     * See Props['historyLength']
     */
    private historyLength = 50

    constructor(props: Props<ItemType>) {
        this.history = [props.current]
        this.onChange = value => props.onChange(value)
        this.historyLength = props.historyLength ?? this.historyLength
    }

    /**
     * (DRY) Method for actual undo/redo.
     * Moves the currentIndex cursor and emits change.
     */
    private moveHistory(action: 'undo' | 'redo'): this {
        if (action === 'undo' && this.currentIndex > 0) {
            this.currentIndex--
        }
        if (action === 'redo' && this.currentIndex < this.historyLength) {
            this.currentIndex++
        }
        if (this.onChange) {
            this.onChange(this.current)
        }
        return this
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
     * Standard undo/redo behavior: once a new value is pushed the future (redo) history is erased.
     * On a push: concat new value to history and set currentIndex to last item of history
     */
    public push(value: ItemType): this {
        if (this.currentIndex < this.historyLength) {
            this.history = this.history
                // respect this.historyLength, and clear redo values when a new value is pushed
                .slice(this.currentIndex - this.historyLength, this.currentIndex + 1)
                .concat(value)
            this.currentIndex = this.history.length - 1
        }
        return this
    }

    public undo(): this {
        return this.moveHistory('undo')
    }

    public redo(): this {
        return this.moveHistory('redo')
    }
}
