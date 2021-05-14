export class BloomFilter {
    public buckets: Int32Array
    constructor(estimatedSize: number, hashFunctionCount: number)
    constructor(serializedBuckets: number, hashFunctionCount: number)
    add(hash: number): void
    test(hash: number): boolean
}
