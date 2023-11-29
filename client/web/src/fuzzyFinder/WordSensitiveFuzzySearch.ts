import { BloomFilter } from 'bloomfilter'

import { type HighlightedLinkProps, offsetSum, type RangePosition } from '../components/fuzzyFinder/HighlightedLink'

import {
    FuzzySearch,
    type IndexingFSM,
    type FuzzySearchParameters,
    type FuzzySearchResult,
    type createUrlFunction,
    type FuzzySearchConstructorParameters,
} from './FuzzySearch'
import { Hasher } from './Hasher'
import type { SearchValue } from './SearchValue'

/**
 * We don't index filenames with length larger than this value.
 */
const MAX_VALUE_LENGTH = 100

// Normally, you need multiple hash functions to keep the false-positive ratio
// low. However, non-empirical observations indicate that a single hash function
// works fine and provides the fastest indexing time in large repositories like
// Chromium.
const DEFAULT_BLOOM_FILTER_HASH_FUNCTION_COUNT = 1

const DEFAULT_INDEXING_BUCKET_SIZE = 25_000
// The number of filenames to group together in a single bucket, and the number
// string prefixes that each bloom can contain.  Currently, every bucket can
// contain up to 262.144 prefixes (conservatively large number).  With bucket
// size 50, my off-the-napkin calculation is that total memory usage with 400k
// files (Chromium size) may be as large as ~261mb. It's usable on most
// computers, but still a bit high.
// Tracking issue to fine-tune these parameters: https://github.com/sourcegraph/sourcegraph/issues/21201
const DEFAULT_BUCKET_SIZE = 50
const DEFAULT_BLOOM_FILTER_SIZE = 2 << 17

/**
 * Returns true if the given query fuzzy matches the given value.
 */
export function fuzzyMatchesQuery(query: string, value: string): RangePosition[] {
    return fuzzyMatches(allFuzzyParts(query, true), value)
}

/**
 * Word-sensitive fuzzy search that
 *
 * Is specifically designed to support low-latency filtering in large
 * repositories (>100k files).
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
export class WordSensitiveFuzzySearch extends FuzzySearch {
    public totalFileCount = 0
    constructor(public readonly buckets: Bucket[]) {
        super()
        for (const bucket of buckets) {
            this.totalFileCount += bucket.files.length
        }
    }

    public static fromSearchValuesAsync(
        files: SearchValue[],
        params?: FuzzySearchConstructorParameters,
        bucketSize: number = DEFAULT_BUCKET_SIZE
    ): IndexingFSM {
        files.sort((a, b) => a.text.length - b.text.length)
        const indexer = new Indexer(files, bucketSize, params)
        function loop(): IndexingFSM {
            if (indexer.isDone()) {
                return { key: 'ready', value: indexer.complete() }
            }
            indexer.processBuckets(DEFAULT_INDEXING_BUCKET_SIZE)
            let indexingPromise: Promise<IndexingFSM> | undefined
            return {
                key: 'indexing',
                indexedFileCount: indexer.indexedFileCount(),
                totalFileCount: indexer.totalFileCount(),
                partialFuzzy: indexer.complete(),
                isIndexing: () => indexingPromise !== undefined,
                continueIndexing: () => {
                    if (!indexingPromise) {
                        indexingPromise = later().then(() => loop())
                    }
                    return indexingPromise
                },
            }
        }
        return loop()
    }

    public static fromSearchValues(
        files: SearchValue[],
        params?: FuzzySearchConstructorParameters,
        bucketSize: number = DEFAULT_BUCKET_SIZE
    ): WordSensitiveFuzzySearch {
        const indexer = new Indexer(files, bucketSize, params)
        while (!indexer.isDone()) {
            indexer.processBuckets(bucketSize)
        }
        return indexer.complete()
    }

    public search(query: FuzzySearchParameters): FuzzySearchResult {
        if (query.query.length === 0) {
            return this.emptyResult(query)
        }
        let falsePositives = 0
        const result: HighlightedLinkProps[] = []
        const hashParts = allQueryHashParts(query.query)
        const queryParts = allFuzzyParts(query.query, true)
        const complete = (isComplete: boolean): FuzzySearchResult =>
            this.sorted({ links: result, isComplete, falsePositiveRatio: falsePositives / this.buckets.length })
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
        result.links.sort((a, b) => {
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
        const result: HighlightedLinkProps[] = []
        const complete = (isComplete: boolean): FuzzySearchResult => this.sorted({ links: result, isComplete })

        for (const bucket of this.buckets) {
            if (result.length > query.maxResults) {
                return complete(false)
            }
            for (const value of bucket.files) {
                result.push({
                    ...value,
                    positions: [],
                    url: value.url || bucket.createUrl?.(value.text),
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
    return isLowercaseOrDigit(value) && !isDelimeter(value)
}
function isLowercaseOrDigit(value: string): boolean {
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
        case ' ': {
            return true
        }
        default: {
            return false
        }
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
        return isLowercaseOrDigit(this.query())
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
        const nextIsLowercase = isLowercaseOrDigit(value[index])
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
    value: HighlightedLinkProps[]
}

export function mergedHandler(
    firstHandler: undefined | (() => void),
    secondHandler: undefined | (() => void)
): undefined | (() => void) {
    // TODO: avoid this weird merging logic
    if (firstHandler && secondHandler) {
        return () => {
            firstHandler()
            secondHandler()
        }
    }
    return firstHandler || secondHandler
}

class Bucket {
    constructor(
        public readonly files: SearchValue[],
        public readonly filter: BloomFilter,
        public readonly id: number,
        public readonly createUrl: createUrlFunction
    ) {}
    public static fromSearchValues(files: SearchValue[], createUrl: createUrlFunction): Bucket {
        files.sort((a, b) => a.text.length - b.text.length)
        return new Bucket(files, populateBloomFilter(files), Math.random(), createUrl)
    }

    private matchesMaybe(hashParts: number[]): boolean {
        for (const part of hashParts) {
            if (!this.filter.test(part)) {
                return false
            }
        }
        return true
    }
    public matches(parameters: FuzzySearchParameters, queryParts: string[], hashParts: number[]): BucketResult {
        const matchesMaybe = this.matchesMaybe(hashParts)
        if (!matchesMaybe) {
            return { skipped: true, value: [] }
        }
        const result: HighlightedLinkProps[] = []
        for (const file of this.files) {
            const positions = fuzzyMatches(queryParts, file.text)
            if (positions.length > 0) {
                result.push({
                    text: file.text,
                    positions,
                    url: this.createUrl?.(file.text),
                    onClick: file.onClick,
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
    constructor(
        private readonly files: SearchValue[],
        private readonly bucketSize: number,
        private readonly params?: FuzzySearchConstructorParameters
    ) {
        this.files.sort((a, b) => a.text.length - b.text.length)
    }

    public complete(): WordSensitiveFuzzySearch {
        return new WordSensitiveFuzzySearch(this.buckets)
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
                this.buckets.push(Bucket.fromSearchValues(this.buffer, this?.params?.createURL))
                this.buffer = []
            }
            bucketCount--
        }
    }
}

async function later(): Promise<void> {
    return new Promise(resolve => setTimeout(() => resolve(), 0))
}
