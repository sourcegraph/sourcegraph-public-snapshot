export class BloomFilter {
    public buckets: Int32Array
    constructor(value: any, hashFunctionCount: number)
    add(hash: number): void
    test(hash: number): boolean
}
