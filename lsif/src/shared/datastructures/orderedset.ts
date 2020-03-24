/** A set of values that maintains insertion order. */
export class OrderedSet<T> {
    private seen = new Set<string>()
    private order: T[] = []

    /**
     * Create a new ordered set.
     *
     * @param makeKey The function to generate a unique string key from a value.
     * @param values A set of values used to seed the set.
     * @param trusted Whether the given values are already deduplicated.
     */
    constructor(private makeKey: (value: T) => string, values?: T[], trusted = false) {
        if (!values) {
            return
        }

        if (trusted) {
            this.order = Array.from(values)
            this.seen = new Set(this.order.map(makeKey))
            return
        }

        for (const location of values) {
            this.push(location)
        }
    }

    /** The deduplicated values in insertion order. */
    public get values(): T[] {
        return this.order
    }

    /** Insert a value into the set if it hasn't been seen before. */
    public push(value: T): void {
        const key = this.makeKey(value)
        if (this.seen.has(key)) {
            return
        }

        this.seen.add(key)
        this.order.push(value)
    }
}
