import { execFile, spawn } from 'child_process'
import * as path from 'path'

import Assembler from 'stream-json/Assembler'
import StreamValues from 'stream-json/streamers/StreamValues'
import * as vscode from 'vscode'
import winkUtils from 'wink-nlp-utils'

import { ChatClient } from '@sourcegraph/cody-shared/src/chat/chat'
import { Editor } from '@sourcegraph/cody-shared/src/editor'
import { KeywordContextFetcher, ContextResult } from '@sourcegraph/cody-shared/src/local-context'

import { logEvent } from '../event-logger'
import { debug } from '../log'

/**
 * Exclude files without extensions and hidden files (starts with '.')
 * Limits to use 1 thread
 * Exclude files larger than 1MB (based on search.largeFiles)
 * Note: Ripgrep excludes binary files and respects .gitignore by default
 */
const fileExtRipgrepParams = [
    '--ignore-case',
    '-g',
    '*.*',
    '-g',
    '!.*',
    '-g',
    '!*.lock',
    '-g',
    '!*.snap',
    '--max-filesize',
    '10K',
    '--max-depth',
    '10',
]

interface RipgrepStreamData {
    value: {
        type: string
        data: {
            path: {
                text: string
            }
            stats: { bytes_searched: number }
        }
    }
}

/**
 * Term represents a single term in the keyword search.
 * - A term is uniquely defined by its stem.
 * - We keep the originals around for detecting exact matches.
 * - The prefix is the greatest common prefix of the stem and all the originals.
 * For example, if the original is "cody" and the stem is "codi", the prefix is "cod"
 * - The count is the number of times the keyword appears in the document/query.
 */
export interface Term {
    stem: string
    originals: string[]
    prefix: string
    count: number
}

export function regexForTerms(...terms: Term[]): string {
    const inner = terms.map(t => {
        if (t.prefix.length >= 4) {
            return escapeRegex(t.prefix)
        }
        return `${escapeRegex(t.stem)}|${t.originals.map(s => escapeRegex(s)).join('|')}`
    })
    return `(?:${inner.join('|')})`
}

function longestCommonPrefix(s: string, t: string): string {
    let endIdx = 0
    for (let i = 0; i < s.length && i < t.length; i++) {
        if (s[i] !== t[i]) {
            break
        }
        endIdx = i + 1
    }
    return s.slice(0, endIdx)
}

/**
 * A local context fetcher that uses a LLM to generate a keyword query, which is then
 * converted to a regex fed to ripgrep to search for files that are relevant to the
 * user query.
 */
export class LocalKeywordContextFetcher implements KeywordContextFetcher {
    constructor(private rgPath: string, private editor: Editor, private chatClient: ChatClient) {}

    /**
     * Returns pieces of context relevant for the given query. Uses a keyword-search-based
     * approach.
     * @param query user query
     * @param numResults the number of context results to return
     * @returns a list of context results, sorted in *reverse* order (that is,
     * the most important result appears at the bottom)
     */
    public async getContext(query: string, numResults: number): Promise<ContextResult[]> {
        const startTime = performance.now()
        const rootPath = this.editor.getWorkspaceRootPath()
        if (!rootPath) {
            return []
        }

        const filesnamesWithScores = await this.fetchKeywordFiles(rootPath, query)
        const top10 = filesnamesWithScores.slice(0, numResults)

        const messagePairs = await Promise.all(
            top10.map(async ({ filename }) => {
                const uri = vscode.Uri.file(path.join(rootPath, filename))
                try {
                    const content = (await vscode.workspace.openTextDocument(uri)).getText()
                    return [{ fileName: filename, content }]
                } catch (error) {
                    // Handle file reading errors in case of concurrent file deletions or binary files
                    console.error(error)
                    return []
                }
            })
        )
        const searchDuration = performance.now() - startTime
        logEvent('CodyVSCodeExtension:keywordContext:searchDuration', searchDuration, searchDuration)
        debug('LocalKeywordContextFetcher:getContext', JSON.stringify({ searchDuration }))

        return messagePairs.reverse().flat()
    }

    private async userQueryToExpandedKeywords(query: string): Promise<Map<string, Term>> {
        const start = performance.now()
        const keywords = await new Promise<string[]>((resolve, reject) => {
            let responseText = ''
            this.chatClient.chat(
                [
                    {
                        speaker: 'human',
                        text: `Write 3-5 keywords that you would use to search for code snippets that are relevant to answering the following user query: <query>${query}</query> Your response should be only a list of space-delimited keywords and nothing else.`,
                    },
                ],
                {
                    onChange: (text: string) => {
                        responseText = text
                    },
                    onComplete: () => {
                        resolve(responseText.split(/\s+/).filter(e => e.length > 0))
                    },
                    onError: (message: string, statusCode?: number) => {
                        reject(new Error(message))
                    },
                },
                {
                    temperature: 0,
                    fast: true,
                }
            )
        })
        const terms = new Map<string, Term>()
        for (const kw of keywords) {
            const stem = winkUtils.string.stem(kw)
            if (terms.has(stem)) {
                continue
            }
            terms.set(stem, {
                count: 1,
                originals: [kw],
                prefix: longestCommonPrefix(kw.toLowerCase(), stem),
                stem,
            })
        }
        debug(
            'LocalKeywordContextFetcher:userQueryToExpandedKeywords',
            JSON.stringify({ duration: performance.now() - start })
        )
        return terms
    }

    private async userQueryToKeywordQuery(query: string): Promise<Term[]> {
        const terms = new Map<string, Term>()
        const keywordExpansionStartTime = Date.now()
        const expandedTerms = await this.userQueryToExpandedKeywords(query)
        const keywordExpansionDuration = Date.now() - keywordExpansionStartTime
        for (const [stem, term] of expandedTerms) {
            if (terms.has(stem)) {
                continue
            }
            terms.set(stem, term)
        }
        debug(
            'LocalKeywordContextFetcher:userQueryToKeywordQuery',
            'keyword expansion',
            JSON.stringify({
                duration: keywordExpansionDuration,
                expandedTerms: [...expandedTerms.values()].map(v => v.prefix),
            })
        )
        const ret = [...terms.values()]
        return ret
    }

    // Return context results for the Codebase Context Search recipe
    public async getSearchContext(query: string, numResults: number): Promise<ContextResult[]> {
        const rootPath = this.editor.getWorkspaceRootPath()
        if (!rootPath) {
            return []
        }

        const stems = (await this.userQueryToKeywordQuery(query))
            .map(t => (t.prefix.length < 4 ? t.originals[0] : t.prefix))
            .join('|')
        const filesnamesWithScores = await this.fetchKeywordFiles(rootPath, query)
        const topN = filesnamesWithScores.slice(0, numResults)

        const messagePairs = await Promise.all(
            topN.map(async ({ filename }) => {
                try {
                    const uri = vscode.Uri.file(path.join(rootPath, filename))
                    const textDocument = await vscode.workspace.openTextDocument(uri)
                    const snippet = textDocument.getText()
                    const keywordPattern = new RegExp(stems, 'g')
                    // show 5 lines of code only
                    // TODO: Rewrite this to use rg instead @bee
                    const matches = snippet.match(keywordPattern)
                    const keywordIndex = snippet.indexOf(matches ? matches[0] : query)
                    const startLine = Math.max(0, textDocument.positionAt(keywordIndex).line - 2)
                    const endLine = startLine + 5
                    const content = textDocument.getText(new vscode.Range(startLine, 0, endLine, 0))

                    return [{ fileName: filename, content }]
                } catch (error) {
                    console.error(error)
                    return []
                }
            })
        )
        return messagePairs.flat()
    }

    private async fetchFileStats(
        terms: Term[],
        rootPath: string
    ): Promise<{ [filename: string]: { bytesSearched: number } }> {
        const start = performance.now()
        const regexQuery = `\\b${regexForTerms(...terms)}`
        const rgArgs = [...fileExtRipgrepParams, '--json', regexQuery, '.']
        const proc = spawn(this.rgPath, rgArgs, {
            cwd: rootPath,
            stdio: ['ignore', 'pipe', process.stderr],
            windowsHide: true,
        })
        const fileTermCounts: {
            [filename: string]: {
                bytesSearched: number
            }
        } = {}

        // Process the ripgrep JSON output to get the file sizes. We use an object filter to
        // fast-filter out irrelevant lines of output
        const objectFilter = (assembler: Assembler): boolean | undefined => {
            // Each ripgrep JSON line begins with the following format:
            //
            //   {"type":"begin|match|end","data":"...
            //
            // We only care about the "type":"end" lines, which contain the file size in bytes.
            if (assembler.key === null && assembler.stack.length === 0 && assembler.current.type) {
                return assembler.current.type === 'end'
            }
            // return undefined to indicate our uncertainty at this moment
            return undefined
        }
        await new Promise<void>((resolve, reject) => {
            try {
                proc.stdout
                    .pipe(StreamValues.withParser({ objectFilter }))
                    .on('data', data => {
                        try {
                            const typedData = data as RipgrepStreamData
                            switch (typedData.value.type) {
                                case 'end': {
                                    let filename = typedData.value.data.path.text
                                    if (filename.startsWith(`.${path.sep}`)) {
                                        filename = filename.slice(2)
                                    }
                                    if (!fileTermCounts[filename]) {
                                        fileTermCounts[filename] = { bytesSearched: 0 }
                                    }
                                    fileTermCounts[filename].bytesSearched = typedData.value.data.stats.bytes_searched
                                    break
                                }
                            }
                        } catch (error) {
                            reject(error)
                        }
                    })
                    .on('end', () => resolve())
            } catch (error) {
                reject(error)
            }
        })
        debug('fetchFileStats', JSON.stringify({ duration: performance.now() - start }))
        return fileTermCounts
    }

    private async fetchFileMatches(
        queryTerms: Term[],
        rootPath: string
    ): Promise<{
        totalFiles: number
        fileTermCounts: { [filename: string]: { [stem: string]: number } }
        termTotalFiles: { [stem: string]: number }
    }> {
        const start = performance.now()
        const termFileCountsArr: { fileCounts: { [filename: string]: number }; filesSearched: number }[] =
            await Promise.all(
                queryTerms.map(async term => {
                    const rgArgs = [
                        ...fileExtRipgrepParams,
                        '--count-matches',
                        '--stats',
                        `\\b${regexForTerms(term)}`,
                        '.',
                    ]
                    const out = await new Promise<string>((resolve, reject) => {
                        execFile(
                            this.rgPath,
                            rgArgs,
                            {
                                cwd: rootPath,
                                maxBuffer: 1024 * 1024 * 1024,
                                timeout: 1000 * 30, // timeout in 30secs
                            },
                            (error, stdout, stderr) => {
                                if (error?.code === 2) {
                                    reject(new Error(`${error.message}: ${stderr}`))
                                } else {
                                    resolve(stdout)
                                }
                            }
                        )
                    })
                    const fileCounts: { [filename: string]: number } = {}
                    const lines = out.split('\n')
                    let filesSearched = -1
                    for (const line of lines) {
                        const terms = line.split(':')
                        if (terms.length !== 2) {
                            const matches = /^(\d+) files searched$/.exec(line)
                            if (matches && matches.length === 2) {
                                try {
                                    filesSearched = parseInt(matches[1], 10)
                                } catch {
                                    console.error(`failed to parse number of files matched from string: ${matches[1]}`)
                                }
                            }
                            continue
                        }
                        try {
                            let filename = terms[0]
                            if (filename.startsWith(`.${path.sep}`)) {
                                filename = filename.slice(2)
                            }
                            const count = parseInt(terms[1], 10)
                            fileCounts[filename] = count
                        } catch {
                            console.error(`could not parse count from ${terms[1]}`)
                        }
                    }
                    return { fileCounts, filesSearched }
                })
            )

        debug('LocalKeywordContextFetcher.fetchFileMatches', JSON.stringify({ duration: performance.now() - start }))
        let totalFilesSearched = -1
        for (const { filesSearched } of termFileCountsArr) {
            if (totalFilesSearched >= 0 && totalFilesSearched !== filesSearched) {
                throw new Error('filesSearched did not match')
            }
            totalFilesSearched = filesSearched
        }

        const fileTermCounts: { [filename: string]: { [stem: string]: number } } = {}
        const termTotalFiles: { [term: string]: number } = {}
        for (let i = 0; i < queryTerms.length; i++) {
            const term = queryTerms[i]
            const fileCounts = termFileCountsArr[i].fileCounts
            termTotalFiles[term.stem] = Object.keys(fileCounts).length

            for (const [filename, count] of Object.entries(fileCounts)) {
                if (!fileTermCounts[filename]) {
                    fileTermCounts[filename] = {}
                }
                fileTermCounts[filename][term.stem] = count
            }
        }

        return {
            totalFiles: totalFilesSearched,
            termTotalFiles,
            fileTermCounts,
        }
    }

    private async fetchKeywordFiles(
        rootPath: string,
        rawQuery: string
    ): Promise<{ filename: string; score: number }[]> {
        const query = await this.userQueryToKeywordQuery(rawQuery)

        const fetchFilesStart = performance.now()
        const fileMatchesPromise = this.fetchFileMatches(query, rootPath)
        const fileStatsPromise = this.fetchFileStats(query, rootPath)
        const fileMatches = await fileMatchesPromise
        const fileStats = await fileStatsPromise
        const fetchFilesDuration = performance.now() - fetchFilesStart
        debug('LocalKeywordContextFetcher:fetchKeywordFiles', JSON.stringify({ fetchFilesDuration }))

        const { fileTermCounts, termTotalFiles, totalFiles } = fileMatches
        const idfDict = idf(termTotalFiles, totalFiles)

        const activeTextEditor = this.editor.getActiveTextEditor()
        const activeFilename = activeTextEditor
            ? path.normalize(vscode.workspace.asRelativePath(activeTextEditor.filePath))
            : undefined

        const querySizeBytes = query
            .flatMap(t => t.originals.map(orig => (orig.length + 1) * t.count))
            .reduce((a, b) => a + b, 0)
        const queryStems = query.map(({ stem }) => stem)
        const queryTf = tf(
            queryStems,
            Object.fromEntries(query.map(({ stem, count }) => [stem, count])),
            querySizeBytes
        )
        const queryVec = tfidf(queryStems, queryTf, idfDict)
        const filenamesWithScores = Object.entries(fileTermCounts)
            .flatMap(([filename, fileTermCounts]) => {
                if (activeFilename === filename) {
                    // The currently active file will always be added as context, so we can skip
                    // over it here
                    return []
                }

                if (fileStats[filename] === undefined) {
                    throw new Error(`filename ${filename} missing from fileStats`)
                }
                const tfVec = tf(queryStems, fileTermCounts, fileStats[filename].bytesSearched)
                const tfidfVec = tfidf(queryStems, tfVec, idfDict)
                const cosineScore = cosine(tfidfVec, queryVec)
                let { score, scoreComponents } = idfLogScore(queryStems, fileTermCounts, idfDict)

                const b = fileStats[filename].bytesSearched
                if (b > 10000) {
                    score *= 0.1 // downweight very large files
                }

                return [
                    {
                        filename,
                        cosineScore,
                        termCounts: fileTermCounts,
                        tfVec,
                        idfDict,
                        score,
                        scoreComponents,
                    },
                ]
            })
            .sort(({ score: score1 }, { score: score2 }) => score2 - score1)

        return uniques(filenamesWithScores)
    }
}

function idfLogScore(
    terms: string[],
    termCounts: { [term: string]: number },
    idfDict: { [term: string]: number }
): { score: number; scoreComponents: { [term: string]: number } } {
    let score = 0
    const scoreComponents: { [term: string]: number } = {}
    for (const term of terms) {
        const ct = termCounts[term] || 0
        const logScore = ct === 0 ? 0 : Math.log10(ct) + 1
        const idfLogScore = (idfDict[term] || 1) * logScore
        score += idfLogScore
        scoreComponents[term] = idfLogScore
    }
    return { score, scoreComponents }
}

function cosine(v1: number[], v2: number[]): number {
    if (v1.length !== v2.length) {
        throw new Error(`v1.length !== v2.length ${v1.length} !== ${v2.length}`)
    }
    let dotProd = 0
    let v1SqMag = 0
    let v2SqMag = 0
    for (let i = 0; i < v1.length; i++) {
        dotProd += v1[i] * v2[i]
        v1SqMag += v1[i] * v1[i]
        v2SqMag += v2[i] * v2[i]
    }
    // return dotProd / Math.sqrt(v1SqMag * v2SqMag)
    return dotProd / (Math.sqrt(v1SqMag) * Math.sqrt(v2SqMag))
}

function tfidf(terms: string[], tf: number[], idf: { [term: string]: number }): number[] {
    if (terms.length !== tf.length) {
        throw new Error(`terms.length !== tf.length ${terms.length} !== ${tf.length}`)
    }
    const tfidf = tf.slice(0)
    for (let i = 0; i < tfidf.length; i++) {
        if (idf[terms[i]] === undefined) {
            throw new Error(`term ${terms[i]} did not exist in idf dict`)
        }
        tfidf[i] *= idf[terms[i]]
    }
    return tfidf
}

function tf(terms: string[], termCounts: { [stem: string]: number }, fileSize: number): number[] {
    return terms.map(term => (termCounts[term] || 0) / fileSize)
}

function idf(termTotalFiles: { [term: string]: number }, totalFiles: number): { [term: string]: number } {
    const logTotal = Math.log(totalFiles)
    const e = Object.entries(termTotalFiles).map(([term, count]) => [term, logTotal - Math.log(count)])
    return Object.fromEntries(e)
}

function escapeRegex(s: string): string {
    return s.replace(/[$()*+./?[\\\]^{|}-]/g, '\\$&')
}

function uniques(results: { filename: string; score: number }[]): { filename: string; score: number }[] {
    const seen = new Set<string>()
    return results.filter(({ filename }) => {
        if (seen.has(filename)) {
            return false
        }
        seen.add(filename)
        return true
    })
}
