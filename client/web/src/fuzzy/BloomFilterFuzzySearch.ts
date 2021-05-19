import { BloomFilter } from './BloomFilter'
import { FuzzySearch, FuzzySearchParameters, FuzzySearchResult } from './FuzzySearch'
import { Hasher } from './Hasher'
import { HighlightedTextProps, offsetSum, RangePosition } from './HighlightedText'

/**
 * We don't index filenames with length larger than this value.
 */
const MAX_VALUE_LENGTH = 100

const DEFAULT_BLOOM_FILTER_HASH_FUNCTION_COUNT = 1
const DEFAULT_BLOOM_FILTER_SIZE = 2 << 17
const DEFAULT_BUCKET_SIZE = 50

/**
 * Returns true if the given query fuzzy matches the given value.
 */
export function fuzzyMatchesQuery(query: string, value: string): RangePosition[] {
    return fuzzyMatches(allFuzzyParts(query, true), value)
}

export interface SearchValue {
    text: string
}

export interface Indexing {
    key: 'indexing'
    indexedFileCount: number
    totalFileCount: number
    partialValue: BloomFilterFuzzySearch
    continue: () => Promise<FuzzySearchLoader>
}
export interface Ready {
    key: 'ready'
    value: BloomFilterFuzzySearch
}
export type FuzzySearchLoader = Indexing | Ready

/**
 * Fuzzy search that uses bloom filters to improve performance in very large repositories.
 *
 * NOTE(olafur): this is a reimplementation of the fuzzy finder in the Scala
 * language server that's documented in this blog post here
 * https://scalameta.org/metals/blog/2019/01/22/bloom-filters.html#fuzzy-symbol-search
 *
 * In a nutshell, bloom filters improve performance by allowing us to skip a
 * "bucket" of candidate files if we know that bucket does not match any words
 * in that query. For example, the query "SymPro" is split into the words "Sym"
 * and "Pro". If a bucket of 500 words is guaranteed to have to appearances of
 * the words "Sym" and "Pro", then we can skip those 500 words and move on to
 * the next bucket.
 *
 * One downside of the bloom filter approach is that it requires an indexing
 * phase that can take a couple of seconds to complete on a large input size
 * (>100k filenames). The indexing phase can take a while to complete because we
 * need to compute all possible words that the user may query. For example,
 * given the filename "SymbolProvider", we create a bloom filter with all
 * possible prefixes of "Symbol" and "Provider". Fortunately, bloom filters can be
 * serialized so that the indexing step only runs once per repoName/commitID pair.
 */
export class BloomFilterFuzzySearch extends FuzzySearch {
    public totalFileCount = 0
    constructor(public readonly buckets: Bucket[]) {
        super()
        for (const bucket of buckets) {
            this.totalFileCount += bucket.files.length
        }
    }

    public static fromSearchValuesAsync(
        files: SearchValue[],
        bucketSize: number = DEFAULT_BUCKET_SIZE
    ): FuzzySearchLoader {
        files.sort((a, b) => a.text.length - b.text.length)
        const indexer = new Indexer(files, bucketSize)
        function loop(): FuzzySearchLoader {
            if (indexer.isDone()) {
                return { key: 'ready', value: indexer.complete() }
            }
            indexer.processBuckets(25000)
            return {
                key: 'indexing',
                indexedFileCount: indexer.indexedFileCount(),
                totalFileCount: indexer.totalFileCount(),
                partialValue: indexer.complete(),
                continue: () => new Promise(resolve => resolve(loop())),
            }
        }
        return loop()
    }

    public static fromSearchValues(
        files: SearchValue[],
        bucketSize: number = DEFAULT_BUCKET_SIZE
    ): BloomFilterFuzzySearch {
        const indexer = new Indexer(files, bucketSize)
        while (!indexer.isDone()) {
            indexer.processBuckets(bucketSize)
        }
        return indexer.complete()
    }

    public search(query: FuzzySearchParameters): FuzzySearchResult {
        if (query.value.length === 0) {
            return this.emptyResult(query)
        }
        let falsePositives = 0
        const result: HighlightedTextProps[] = []
        const hashParts = allQueryHashParts(query.value)
        const queryParts = allFuzzyParts(query.value, true)
        const complete = (isComplete: boolean): FuzzySearchResult =>
            this.sorted({ values: result, isComplete, falsePositiveRatio: falsePositives / this.buckets.length })
        for (const bucket of this.buckets) {
            const matches = bucket.matches(query, queryParts, hashParts)
            if (!matches.skipped && matches.value.length === 0) {
                falsePositives++
            }
            for (const value of matches.value) {
                if (result.length >= query.maxResults) {
                    return complete(false)
                }
                result.push(value)
            }
        }
        return complete(true)
    }

    private sorted(result: FuzzySearchResult): FuzzySearchResult {
        result.values.sort((a, b) => {
            const byLength = a.text.length - b.text.length
            if (byLength !== 0) {
                return byLength
            }

            const byEarliestMatch = offsetSum(a) - offsetSum(b)
            if (byEarliestMatch !== 0) {
                return byEarliestMatch
            }

            return a.text.localeCompare(b.text)
        })
        return result
    }

    private emptyResult(query: FuzzySearchParameters): FuzzySearchResult {
        const result: HighlightedTextProps[] = []
        const complete = (isComplete: boolean): FuzzySearchResult => this.sorted({ values: result, isComplete })

        for (const bucket of this.buckets) {
            if (result.length > query.maxResults) {
                return complete(false)
            }
            for (const value of bucket.files) {
                result.push({
                    text: value.text,
                    positions: [],
                    url: query.createUrl ? query.createUrl(value.text) : undefined,
                })
                if (result.length > query.maxResults) {
                    return complete(false)
                }
            }
        }
        return complete(true)
    }
}

export function allFuzzyParts(value: string, includeDelimeters: boolean): string[] {
    const buf: string[] = []
    let start = 0
    let end = 0
    while (end < value.length) {
        if (end > start) {
            buf.push(value.slice(start, end))
        }
        while (end < value.length && isDelimeter(value[end])) {
            if (includeDelimeters) {
                buf.push(value[end])
            }
            end++
        }
        start = end
        end = nextFuzzyPart(value, end + 1)
    }

    if (start < value.length && end > start) {
        buf.push(value.slice(start, end))
    }

    return buf
}

function isDigit(value: string): boolean {
    return value >= '0' && value <= '9'
}
function isLowercaseCharacter(value: string): boolean {
    return isLowercase(value) && !isDelimeter(value)
}
function isLowercase(value: string): boolean {
    return isDigit(value) || (value.toLowerCase() === value && value !== value.toUpperCase())
}

function isUppercaseCharacter(value: string): boolean {
    return isUppercase(value) && !isDelimeter(value)
}
function isUppercase(value: string): boolean {
    return value.toUpperCase() === value && value !== value.toLowerCase()
}

function isDelimeterOrUppercase(character: string): boolean {
    return isDelimeter(character) || isUppercase(character)
}

function isDelimeter(character: string): boolean {
    switch (character) {
        case '/':
        case '_':
        case '-':
        case '.':
        case ' ':
            return true
        default:
            return false
    }
}

function fuzzyMatches(queries: string[], value: string): RangePosition[] {
    const result: RangePosition[] = []
    const matcher = new FuzzyMatcher(queries, value)
    while (!matcher.isDone()) {
        const isCurrentQueryDelimeter = matcher.isQueryDelimeter()
        while (!matcher.isQueryDelimeter() && matcher.isStartDelimeter()) {
            matcher.start++
        }
        if (matcher.matchesFromStart()) {
            result.push(matcher.rangePositionFromStart())
            matcher.queryIndex++
        }
        matcher.start = matcher.nextStart(isCurrentQueryDelimeter)
    }
    return matcher.queryIndex >= queries.length ? result : []
}

class FuzzyMatcher {
    public queryIndex = 0
    public start = 0
    private lowercaseValue: string
    constructor(private readonly queries: string[], private readonly value: string) {
        this.lowercaseValue = value.toLowerCase()
    }
    public nextStart(isCurrentQueryDelimeter: boolean): number {
        const offset = isCurrentQueryDelimeter ? this.start : this.start + 1
        let end = this.isQueryDelimeter()
            ? this.indexOfDelimeter(this.query(), offset)
            : nextFuzzyPart(this.value, offset)
        while (end < this.value.length && !this.isQueryDelimeter() && isDelimeter(this.value[end])) {
            end++
        }
        return end
    }
    public rangePositionFromStart(): RangePosition {
        const end = this.start + this.query().length
        return {
            startOffset: this.start,
            endOffset: end,
            isExact: end >= this.value.length || startsNewWord(this.value, end),
        }
    }
    public matchesFromStart(): boolean {
        const caseInsensitive = this.isCaseInsensitive()
        const compareValue = caseInsensitive ? this.lowercaseValue : this.value
        return (
            compareValue.startsWith(this.query(), this.start) &&
            (!caseInsensitive || isCapitalizedPart(this.value, this.start, this.query()))
        )
    }
    public isStartDelimeter(): boolean {
        return isDelimeter(this.value[this.start])
    }
    public isDone(): boolean {
        return this.queryIndex >= this.queries.length || this.start >= this.value.length
    }
    public query(): string {
        return this.queries[this.queryIndex]
    }
    public isCaseInsensitive(): boolean {
        return isLowercase(this.query())
    }
    public isQueryDelimeter(): boolean {
        return isDelimeter(this.query())
    }
    public indexOfDelimeter(delim: string, start: number): number {
        const index = this.value.indexOf(delim, start)
        return index < 0 ? this.value.length : index
    }
}

function startsNewWord(value: string, index: number): boolean {
    return (
        isDelimeterOrUppercase(value[index]) ||
        (isLowercaseCharacter(value[index]) && !isLowercaseCharacter(value[index - 1]))
    )
}

/**
 * Returns true if value.substring(start, start + query.length) is "properly capitalized".
 *
 * The string is properly capitalized as long it contains no lowercase character
 * that is followed by an uppercase character.  For example:
 *
 * - Not properly capitalized: "InnerClasses" "innerClasses"
 * - Properly capitalized: "Innerclasses" "INnerclasses"
 */
function isCapitalizedPart(value: string, start: number, query: string): boolean {
    let previousIsLowercase = false
    for (let index = start; index < value.length && index - start < query.length; index++) {
        const nextIsLowercase = isLowercase(value[index])
        if (previousIsLowercase && !nextIsLowercase) {
            return false
        }
        previousIsLowercase = nextIsLowercase
    }
    return true
}

function nextFuzzyPart(value: string, start: number): number {
    let end = start
    while (end < value.length && !isDelimeterOrUppercase(value[end])) {
        end++
    }
    return end
}

function populateBloomFilter(values: SearchValue[]): BloomFilter {
    const hashes = new BloomFilter(DEFAULT_BLOOM_FILTER_SIZE, DEFAULT_BLOOM_FILTER_HASH_FUNCTION_COUNT)
    for (const value of values) {
        if (value.text.length < MAX_VALUE_LENGTH) {
            updateHashParts(value.text, hashes)
        }
    }
    return hashes
}

function allQueryHashParts(query: string): number[] {
    const fuzzyParts = allFuzzyParts(query, false)
    const result: number[] = []
    const hasher = new Hasher()
    for (const part of fuzzyParts) {
        hasher.reset()
        for (const character of part) {
            hasher.update(character)
            result.push(hasher.digest())
        }
    }
    return result
}

function updateHashParts(value: string, buf: BloomFilter): void {
    const words = new Hasher()
    const lowercaseWords = new Hasher()

    for (let index = 0; index < value.length; index++) {
        const character = value[index]
        if (isDelimeterOrUppercase(character)) {
            words.reset()
            lowercaseWords.reset()
            if (isUppercaseCharacter(character) && (index === 0 || !isUppercaseCharacter(value[index - 1]))) {
                let uppercaseWordIndex = index
                const upper = []
                while (uppercaseWordIndex < value.length && isUppercaseCharacter(value[uppercaseWordIndex])) {
                    upper.push(value[uppercaseWordIndex])
                    lowercaseWords.update(value[uppercaseWordIndex].toLowerCase())
                    buf.add(lowercaseWords.digest())
                    uppercaseWordIndex++
                }
                lowercaseWords.reset()
            }
        }
        if (isDelimeter(character)) {
            continue
        }
        words.update(character)
        lowercaseWords.update(character.toLowerCase())

        buf.add(words.digest())
        if (words.digest() !== lowercaseWords.digest()) {
            buf.add(lowercaseWords.digest())
        }
    }
}

interface BucketResult {
    skipped: boolean
    value: HighlightedTextProps[]
}

class Bucket {
    constructor(
        public readonly files: SearchValue[],
        public readonly filter: BloomFilter,
        public readonly id: number
    ) {}
    public static fromSearchValues(files: SearchValue[]): Bucket {
        files.sort((a, b) => a.text.length - b.text.length)
        return new Bucket(files, populateBloomFilter(files), Math.random())
    }

    private matchesMaybe(hashParts: number[]): boolean {
        for (const part of hashParts) {
            if (!this.filter.test(part)) {
                return false
            }
        }
        return true
    }
    public matches(query: FuzzySearchParameters, queryParts: string[], hashParts: number[]): BucketResult {
        const matchesMaybe = this.matchesMaybe(hashParts)
        if (!matchesMaybe) {
            return { skipped: true, value: [] }
        }
        const result: HighlightedTextProps[] = []
        for (const file of this.files) {
            const positions = fuzzyMatches(queryParts, file.text)
            if (positions.length > 0) {
                result.push({
                    text: file.text,
                    positions,
                    url: query.createUrl ? query.createUrl(file.text) : undefined,
                })
            }
        }
        return { skipped: false, value: result }
    }
}

class Indexer {
    private buffer: SearchValue[] = []
    private buckets: Bucket[] = []
    private index = 0
    constructor(private readonly files: SearchValue[], private readonly bucketSize: number) {
        this.files.sort((a, b) => a.text.length - b.text.length)
    }

    public complete(): BloomFilterFuzzySearch {
        return new BloomFilterFuzzySearch(this.buckets)
    }

    public isDone(): boolean {
        return this.index >= this.files.length
    }
    public totalFileCount(): number {
        return this.files.length
    }
    public indexedFileCount(): number {
        return this.index
    }
    public processBuckets(fileCount: number): void {
        let bucketCount = fileCount / this.bucketSize
        while (bucketCount > 0 && !this.isDone()) {
            const endIndex = Math.min(this.files.length, this.index + this.bucketSize)
            while (this.index < endIndex) {
                this.buffer.push(this.files[this.index])
                this.index++
            }
            if (this.buffer) {
                this.buckets.push(Bucket.fromSearchValues(this.buffer))
                this.buffer = []
            }
            bucketCount--
        }
    }
}
