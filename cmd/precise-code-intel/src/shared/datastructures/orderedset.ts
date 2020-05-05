/** A set of values that maintains insertion order. */
export class OrderedSet<T> {
    private set = new Map<string, T>()

    /**
     * Create a new ordered set.
     *
     * @param makeKey The function to generate a unique string key from a value.
     * @param values A set of values used to seed the set.
     */
    constructor(private makeKey: (value: T) => string, values?: T[]) {
        if (!values) {
            return
        }

        for (const value of values) {
            this.push(value)
        }
    }

    /** The deduplicated values in insertion order. */
    public get values(): T[] {
        return Array.from(this.set.values())
    }

    /** Insert a value into the set if it hasn't been seen before. */
    public push(value: T): void {
        this.set.set(this.makeKey(value), value)
    }
}
