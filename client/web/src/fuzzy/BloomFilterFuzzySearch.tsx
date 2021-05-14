import { Hasher } from './Hasher'
import { BloomFilter } from './BloomFilter'
import { HighlightedTextProps, RangePosition } from './HighlightedText'
import { FuzzySearch, FuzzySearchParameters, FuzzySearchResult } from './FuzzySearch'

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
    constructor(readonly buckets: Bucket[], readonly BUCKET_SIZE: number = DEFAULT_BUCKET_SIZE) {
        super()
    }
    public static fromSearchValues(
        files: SearchValue[],
        BUCKET_SIZE: number = DEFAULT_BUCKET_SIZE
    ): BloomFilterFuzzySearch {
        const buckets = []
        let buffer: SearchValue[] = []
        files.forEach(file => {
            buffer.push(file)
            if (buffer.length >= BUCKET_SIZE) {
                buckets.push(Bucket.fromSearchValues(buffer))
                buffer = []
            }
        })
        if (buffer) buckets.push(Bucket.fromSearchValues(buffer))
        return new BloomFilterFuzzySearch(buckets, BUCKET_SIZE)
    }

    public serialize(): any {
        return {
            buckets: this.buckets.map(b => b.serialize()),
            BUCKET_SIZE: this.BUCKET_SIZE,
        }
    }

    public static fromSerializedString(text: string): BloomFilterFuzzySearch {
        const json = JSON.parse(text) as any
        return new BloomFilterFuzzySearch(json.buckets.map(Bucket.fromSerializedString), json.BUCKET_SIZE)
    }

    public search(query: FuzzySearchParameters): FuzzySearchResult {
        if (query.value.length === 0) return this.emptyResult(query)
        const self = this
        const result: HighlightedTextProps[] = []
        const finalQuery = query.value // this.actualQuery(query.value)
        const hashParts = allQueryHashParts(finalQuery)
        function complete(isComplete: boolean) {
            return self.sorted({ values: result, isComplete: isComplete })
        }
        for (var i = 0; i < this.buckets.length; i++) {
            const bucket = this.buckets[i]
            const matches = bucket.matches(finalQuery, hashParts)
            for (var j = 0; j < matches.value.length; j++) {
                if (result.length >= query.maxResults) {
                    return complete(false)
                }
                result.push(matches.value[j])
            }
        }
        return complete(true)
    }

    private sorted(result: FuzzySearchResult): FuzzySearchResult {
        result.values.sort((a, b) => {
            const byLength = a.text.length - b.text.length
            if (byLength !== 0) return byLength
            const byEarliestMatch = a.offsetSum() - b.offsetSum()
            if (byEarliestMatch !== 0) return byEarliestMatch

            return a.text.localeCompare(b.text)
        })
        return result
    }

    private emptyResult(query: FuzzySearchParameters): FuzzySearchResult {
        const result: HighlightedTextProps[] = []
        const self = this
        function complete(isComplete: boolean) {
            return self.sorted({ values: result, isComplete: isComplete })
        }

        for (var i = 0; i < this.buckets.length; i++) {
            const bucket = this.buckets[i]
            if (result.length > query.maxResults) return complete(false)
            for (var j = 0; j < bucket.files.length; j++) {
                const value = bucket.files[j]
                result.push(new HighlightedTextProps(value.value, [], value.url))
                if (result.length > query.maxResults) return complete(false)
            }
        }
        return complete(true)
    }
}

export function allFuzzyParts(value: string, includeDelimeters: boolean): string[] {
    const buf: string[] = []
    var start = 0
    for (var end = 0; end < value.length; end = nextFuzzyPart(value, end)) {
        if (end > start) {
            buf.push(value.substring(start, end))
        }
        while (end < value.length && isDelimeter(value[end])) {
            if (includeDelimeters) {
                buf.push(value[end])
            }
            end++
        }
        start = end
        end++
    }

    if (start < value.length && end > start) {
        buf.push(value.substring(start, end))
    }

    return buf
}

function isLowercase(str: string): boolean {
    return str.toLowerCase() === str && str !== str.toUpperCase()
}

function isUppercaseCharacter(str: string): boolean {
    return isUppercase(str) && !isDelimeter(str)
}
function isUppercase(str: string): boolean {
    return str.toUpperCase() === str && str !== str.toLowerCase()
}

function isDelimeterOrUppercase(ch: string): boolean {
    return isDelimeter(ch) || isUppercase(ch)
}

function isDelimeter(ch: string): boolean {
    switch (ch) {
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
    var queryIndex = 0
    function query(): string {
        return queries[queryIndex]
    }
    function isCaseInsensitive(): boolean {
        return isLowercase(query())
    }
    function isQueryDelimeter(): boolean {
        return isDelimeter(query())
    }
    function indexOfDelimeter(delim: string, i: number) {
        const index = value.indexOf(delim, i)
        return index < 0 ? value.length : index
    }
    var start = 0
    while (queryIndex < queries.length && start < value.length) {
        const isCurrentQueryDelimeter = isQueryDelimeter()
        while (!isQueryDelimeter() && isDelimeter(value[start])) start++
        const caseInsensitive = isCaseInsensitive()
        const compareValue = caseInsensitive ? lowercaseValue : value
        if (compareValue.startsWith(query(), start) && (!caseInsensitive || isCapitalizedPart(value, start, query()))) {
            const end = start + query().length
            result.push({
                startOffset: start,
                endOffset: end,
                isExact: end < value.length && isDelimeterOrUppercase(value[end]),
            })
            queryIndex++
        }
        const nextStart = isCurrentQueryDelimeter ? start : start + 1
        let end = isQueryDelimeter() ? indexOfDelimeter(query(), nextStart) : nextFuzzyPart(value, nextStart)
        while (end < value.length && !isQueryDelimeter && isDelimeter(value[end])) end++
        start = end
    }
    return queryIndex >= queries.length ? result : []
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
function isCapitalizedPart(value: string, start: number, query: string) {
    let previousIsLowercase = false
    for (var i = start; i < value.length && i - start < query.length; i++) {
        const nextIsLowercase = isLowercase(value[i])
        if (previousIsLowercase && !nextIsLowercase) return false
        previousIsLowercase = nextIsLowercase
    }
    return true
}

function nextFuzzyPart(value: string, start: number): number {
    var end = start
    while (end < value.length && !isDelimeterOrUppercase(value[end])) end++
    return end
}

function populateBloomFilter(values: SearchValue[]): BloomFilter {
    let hashes = new BloomFilter(DEFAULT_BLOOM_FILTER_SIZE, DEFAULT_BLOOM_FILTER_HASH_FUNCTION_COUNT)
    values.forEach(value => {
        if (value.value.length < MAX_VALUE_LENGTH) {
            updateHashParts(value.value, hashes)
        }
    })
    return hashes
}

function allQueryHashParts(query: string): number[] {
    const fuzzyParts = allFuzzyParts(query, false)
    const result: number[] = []
    const H = new Hasher()
    for (var i = 0; i < fuzzyParts.length; i++) {
        H.reset()
        const part = fuzzyParts[i]
        for (var j = 0; j < part.length; j++) {
            H.update(part[j])
        }
        const digest = H.digest()
        result.push(digest)
    }
    return result
}

function updateHashParts(value: string, buf: BloomFilter): void {
    let H = new Hasher(),
        L = new Hasher()

    for (var i = 0; i < value.length; i++) {
        const ch = value[i]
        if (isDelimeterOrUppercase(ch)) {
            H.reset()
            L.reset()
            if (isUppercaseCharacter(ch) && (i === 0 || !isUppercaseCharacter(value[i - 1]))) {
                let j = i
                const upper = []
                while (j < value.length && isUppercaseCharacter(value[j])) {
                    upper.push(value[j])
                    L.update(value[j].toLowerCase())
                    buf.add(L.digest())
                    j++
                }
                L.reset()
            }
        }
        if (isDelimeter(ch)) continue
        H.update(ch)
        L.update(ch.toLowerCase())

        buf.add(H.digest())
        if (H.digest() !== L.digest()) {
            buf.add(L.digest())
        }
    }
}

interface BucketResult {
    skipped: boolean
    value: HighlightedTextProps[]
}

class Bucket {
    constructor(readonly files: SearchValue[], readonly filter: BloomFilter, readonly id: number) {}
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
        return hashParts.every(num => this.filter.test(num))
    }
    public matches(query: string, hashParts: number[]): BucketResult {
        const matchesMaybe = this.matchesMaybe(hashParts)
        if (!matchesMaybe) return { skipped: true, value: [] }
        const result: HighlightedTextProps[] = []
        const queryParts = allFuzzyParts(query, true)
        for (var i = 0; i < this.files.length; i++) {
            const file = this.files[i]
            const positions = fuzzyMatches(queryParts, file.value)
            if (positions.length > 0) {
                result.push(new HighlightedTextProps(file.value, positions, file.url))
            }
        }
        return { skipped: false, value: result }
    }
}
