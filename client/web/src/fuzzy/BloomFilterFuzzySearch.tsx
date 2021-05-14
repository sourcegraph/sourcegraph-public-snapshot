/* eslint-disable @typescript-eslint/no-unsafe-call */
/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-unsafe-assignment */
/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unsafe-return */
import { BloomFilter } from './BloomFilter'
import { FuzzySearch, FuzzySearchParameters, FuzzySearchResult } from './FuzzySearch'
import { Hasher } from './Hasher'
import { HighlightedTextProps, RangePosition } from './HighlightedText'

/**
 * We don't index filenames with length larger than this value.
 */
const MAX_VALUE_LENGTH = 100

const DEFAULT_BLOOM_FILTER_HASH_FUNCTION_COUNT = 8
const DEFAULT_BLOOM_FILTER_SIZE = 2 << 17

/**
 * Returns true if the given query fuzzy matches the given value.
 */
export function fuzzyMatchesQuery(query: string, value: string): RangePosition[] {
    return fuzzyMatches(allFuzzyParts(query, true), value)
}

export interface SearchValue {
    value: string
    url?: string
}

const DEFAULT_BUCKET_SIZE = 500

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
    constructor(public readonly buckets: Bucket[], public readonly BUCKET_SIZE: number = DEFAULT_BUCKET_SIZE) {
        super()
    }
    public static fromSearchValues(
        files: SearchValue[],
        BUCKET_SIZE: number = DEFAULT_BUCKET_SIZE
    ): BloomFilterFuzzySearch {
        files.sort((a, b) => a.value.length - b.value.length)
        const buckets = []
        let buffer: SearchValue[] = []
        for (const file of files) {
            buffer.push(file)
            if (buffer.length >= BUCKET_SIZE) {
                buckets.push(Bucket.fromSearchValues(buffer))
                buffer = []
            }
        }
        if (buffer) {
            buckets.push(Bucket.fromSearchValues(buffer))
        }
        return new BloomFilterFuzzySearch(buckets, BUCKET_SIZE)
    }

    public serialize(): any {
        return {
            buckets: this.buckets.map(b => b.serialize()),
            BUCKET_SIZE: this.BUCKET_SIZE,
        }
    }

    public static fromSerializedString(text: string): BloomFilterFuzzySearch {
        const json = JSON.parse(text)
        return new BloomFilterFuzzySearch(
            json.buckets.map((bucket: any) => Bucket.fromSerializedString(bucket)),
            json.BUCKET_SIZE
        )
    }

    public search(query: FuzzySearchParameters): FuzzySearchResult {
        if (query.value.length === 0) {
            return this.emptyResult(query)
        }
        const result: HighlightedTextProps[] = []
        const hashParts = allQueryHashParts(query.value)
        const complete = (isComplete: boolean): FuzzySearchResult => this.sorted({ values: result, isComplete })
        for (const bucket of this.buckets) {
            const matches = bucket.matches(query.value, hashParts)
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

            const byEarliestMatch = a.offsetSum() - b.offsetSum()
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
                result.push(new HighlightedTextProps(value.value, [], value.url))
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

function isLowercaseCharacter(value: string): boolean {
    return isLowercase(value) && !isDelimeter(value)
}
function isLowercase(value: string): boolean {
    return value.toLowerCase() === value && value !== value.toUpperCase()
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
    const lowercaseValue = value.toLowerCase()
    const result: RangePosition[] = []
    let queryIndex = 0
    let start = 0
    function query(): string {
        return queries[queryIndex]
    }
    function isCaseInsensitive(): boolean {
        return isLowercase(query())
    }
    function isQueryDelimeter(): boolean {
        return isDelimeter(query())
    }
    function indexOfDelimeter(delim: string, start: number): number {
        const index = value.indexOf(delim, start)
        return index < 0 ? value.length : index
    }
    while (queryIndex < queries.length && start < value.length) {
        const isCurrentQueryDelimeter = isQueryDelimeter()
        while (!isQueryDelimeter() && isDelimeter(value[start])) {
            start++
        }
        const caseInsensitive = isCaseInsensitive()
        const compareValue = caseInsensitive ? lowercaseValue : value
        if (compareValue.startsWith(query(), start) && (!caseInsensitive || isCapitalizedPart(value, start, query()))) {
            const end = start + query().length
            result.push({
                startOffset: start,
                endOffset: end,
                isExact: end >= value.length || startsNewWord(value, end),
            })
            queryIndex++
        }
        const nextStart = isCurrentQueryDelimeter ? start : start + 1
        let end = isQueryDelimeter() ? indexOfDelimeter(query(), nextStart) : nextFuzzyPart(value, nextStart)
        while (end < value.length && !isQueryDelimeter && isDelimeter(value[end])) {
            end++
        }
        start = end
    }
    return queryIndex >= queries.length ? result : []
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
        if (value.value.length < MAX_VALUE_LENGTH) {
            updateHashParts(value.value, hashes)
        }
    }
    return hashes
}

function allQueryHashParts(query: string): number[] {
    const fuzzyParts = allFuzzyParts(query, false)
    const result: number[] = []
    const H = new Hasher()
    for (const part of fuzzyParts) {
        H.reset()
        for (const character of part) {
            H.update(character)
        }
        const digest = H.digest()
        result.push(digest)
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
        files.sort((a, b) => a.value.length - b.value.length)
        return new Bucket(files, populateBloomFilter(files), Math.random())
    }
    public static fromSerializedString(json: any): Bucket {
        return new Bucket(json.files, new BloomFilter(json.filter, DEFAULT_BLOOM_FILTER_HASH_FUNCTION_COUNT), json.id)
    }
    public serialize(): any {
        return {
            files: this.files,
            filter: [].slice.call(this.filter.buckets),
        }
    }

    private matchesMaybe(hashParts: number[]): boolean {
        return hashParts.every(number => this.filter.test(number))
    }
    public matches(query: string, hashParts: number[]): BucketResult {
        const matchesMaybe = this.matchesMaybe(hashParts)
        if (!matchesMaybe) {
            return { skipped: true, value: [] }
        }
        const result: HighlightedTextProps[] = []
        const queryParts = allFuzzyParts(query, true)
        for (const file of this.files) {
            const positions = fuzzyMatches(queryParts, file.value)
            if (positions.length > 0) {
                result.push(new HighlightedTextProps(file.value, positions, file.url))
            }
        }
        return { skipped: false, value: result }
    }
}
