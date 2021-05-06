import { Hasher } from './Hasher'
import { BloomFilter } from './BloomFilter'
import { HighlightedTextProps, RangePosition } from './HighlightedText'
import { QueryProps } from './FuzzyFiles'

function isCaseInsensitiveQuery(query: string): boolean {
    return isLowercase(query)
}
function isLowercase(str: string): boolean {
    return str.toLowerCase() === str && str !== str.toUpperCase()
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

const MAX_VALUE_LENGTH = 100
const SMALL_QUERY_SIZE = 3
function isSmallQuery(query: string): boolean {
    return query.length <= SMALL_QUERY_SIZE
}
const SMALL_QUERY_MAGIC = '😅'

export function fuzzyMatchesQuery(query: string, value: string): RangePosition[] {
    const isCaseInsensitive = query === query.toLowerCase()
    return fuzzyMatches(allFuzzyParts(query), value, isCaseInsensitive)
}
export function fuzzyMatches(queries: string[], value: string, isCaseInsensitive: boolean): RangePosition[] {
    const result: RangePosition[] = []
    var queryIndex = 0
    var start = 0
    while (queryIndex < queries.length && start < value.length) {
        const query = queries[queryIndex]
        if (value.startsWith(query, start)) {
            const end = start + query.length
            result.push({
                startOffset: start,
                endOffset: end,
                isExact: end < value.length && isDelimeterOrUppercase(value[end]),
            })
            queryIndex++
        }
        let end = nextFuzzyPart(value, start + 1)
        while (end < value.length && isDelimeter(value[end])) end++
        start = end
    }
    return queryIndex >= queries.length ? result : []
}
export function allFuzzyParts(value: string): string[] {
    const buf: string[] = []
    var start = 0
    for (var end = 0; end < value.length; end = nextFuzzyPart(value, end)) {
        if (end > start) {
            buf.push(value.substring(start, end))
        }
        while (end < value.length && isDelimeter(value[end])) end++
        start = end
        end++
    }
    buf.push(value.substring(start, end))
    return buf
}
export function nextFuzzyPart(value: string, start: number): number {
    var end = start
    while (end < value.length && !isDelimeterOrUppercase(value[end])) end++
    return end
}
function populateBloomFilter(values: string[]): BloomFilter {
    let hashes = new BloomFilter(2 << 17, 8)
    values.forEach(value => {
        if (value.length < MAX_VALUE_LENGTH) {
            updateHashParts(value, hashes)
        }
    })
    return hashes
}
function allQueryHashParts(query: string, isExact: boolean): number[] {
    const fuzzyParts = allFuzzyParts(query)
    const result: number[] = []
    const H = new Hasher()
    for (var i = 0; i < fuzzyParts.length; i++) {
        H.reset()
        const part = fuzzyParts[i]
        for (var j = 0; j < part.length; j++) {
            H.update(part[j])
        }
        const digest = H.digest()
        // console.log(`part=${part} digest=${digest} chars=${chars.join("")}`);
        result.push(digest)
    }
    if (result.length === 1 && fuzzyParts.length === 1 && isExact) {
        result.push(H.update(SMALL_QUERY_MAGIC).digest())
    }
    return result
}

function updateHashParts(value: string, buf: BloomFilter): void {
    let H = new Hasher()
    let size = 0
    for (var i = 0; i < value.length; i++) {
        const ch = value[i]
        if (isDelimeterOrUppercase(ch)) {
            if (size <= SMALL_QUERY_SIZE) {
                H.update(SMALL_QUERY_MAGIC)
                buf.add(H.digest())
            }
            H.reset()
            size = 0
        }
        if (isDelimeter(ch)) continue
        H.update(ch)
        const digest = H.digest()
        // console.log(`chars=${chars.join("")} digest=${digest}`);
        buf.add(digest)
    }
}

interface BucketResult {
    skipped: boolean
    value: HighlightedTextProps[]
}

class Bucket {
    filter: BloomFilter
    id: number
    constructor(readonly files: string[]) {
        // console.log(`files=${files} hashes=${hashes}`);
        this.filter = populateBloomFilter(files)
        this.id = Math.random()
    }

    private matchesMaybe(hashParts: number[]): boolean {
        return hashParts.every(num => this.filter.test(num))
    }
    public matches(query: string, hashParts: number[], isCaseInsensitive: boolean): BucketResult {
        const matchesMaybe = this.matchesMaybe(hashParts)
        // console.log(
        //   `query=${query} matchesMaybe=${matchesMaybe} hashParts=${hashParts}`
        // );
        if (!matchesMaybe) return { skipped: true, value: [] }
        const result: HighlightedTextProps[] = []
        const queryParts = allFuzzyParts(query)
        for (var i = 0; i < this.files.length; i++) {
            const file = this.files[i]
            const positions = fuzzyMatches(queryParts, file, isCaseInsensitive)
            if (positions.length > 0) {
                result.push(new HighlightedTextProps(file, positions))
            }
        }
        return { skipped: false, value: result }
    }
}

export class FuzzySearch {
    buckets: Bucket[]
    BUCKET_SIZE = 500
    constructor(readonly files: string[]) {
        this.buckets = []
        let buffer: string[] = []
        files.forEach(file => {
            buffer.push(file)
            if (buffer.length >= this.BUCKET_SIZE) {
                this.buckets.push(new Bucket(buffer))
                buffer = []
            }
        })
        if (buffer) this.buckets.push(new Bucket(buffer))
    }
    private actualQuery(query: string): string {
        let end = query.length - 1
        while (end > 0 && isDelimeter(query[end])) end--
        return query.substring(0, end + 1)
    }
    public search(query: QueryProps): HighlightedTextProps[] {
        const result: HighlightedTextProps[] = []
        const finalQuery = this.actualQuery(query.value)
        const isExact = isSmallQuery(query.value)
        const isCaseInsensitive = isCaseInsensitiveQuery(query.value)
        const isVisited = new Set<number>()
        const hashParts = allQueryHashParts(finalQuery, isExact)
        this.updateSearchResults(finalQuery, hashParts, isCaseInsensitive, result, isVisited)
        if (isExact && result.length < query.maxResults) {
            const nonExactParts = allQueryHashParts(finalQuery, false)
            this.updateSearchResults(finalQuery, nonExactParts, isCaseInsensitive, result, isVisited)
        }
        result.sort((a, b) => {
            const byLength = a.text.length - b.text.length
            if (byLength !== 0) return byLength
            const byEarliestMatch = a.offsetSum() - b.offsetSum()
            if (byEarliestMatch !== 0) return byEarliestMatch

            return a.text.localeCompare(b.text)
        })
        return result
    }

    private updateSearchResults(
        finalQuery: string,
        hashParts: number[],
        isCaseInsensitive: boolean,
        result: HighlightedTextProps[],
        isVisited: Set<number>
    ): void {
        this.buckets.forEach(bucket => {
            if (!isVisited.has(bucket.id)) {
                const matches = bucket.matches(finalQuery, hashParts, isCaseInsensitive)
                if (!matches.skipped) isVisited.add(bucket.id)
                result.push(...matches.value)
            }
        })
    }
}
