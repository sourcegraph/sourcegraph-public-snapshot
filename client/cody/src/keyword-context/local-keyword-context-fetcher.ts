import { execFile, spawn } from 'child_process'
import * as path from 'path'

import StreamValues from 'stream-json/streamers/StreamValues'
import * as vscode from 'vscode'
import winkUtils from 'wink-nlp-utils'

import { Editor } from '@sourcegraph/cody-shared/src/editor'
import { KeywordContextFetcher, KeywordContextFetcherResult } from '@sourcegraph/cody-shared/src/keyword-context'

const fileExtRipgrepParams = ['-Tmarkdown', '-Tyaml', '-Tjson', '-g', '!*.lock', '-g', '!*.snap']

/**
 * Term represents a single term in the keyword search.
 * - A term is uniquely defined by its stem.
 * - We keep the originals around for detecting exact matches.
 * - The prefix is the greatest common prefix of the stem and all the originals.
 * For example, if the original is "cody" and the stem is "codi", the prefix is "cod"
 * - The count is the number of times the keyword appears in the document/query.
 */
interface Term {
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

export function userQueryToKeywordQuery(query: string): Term[] {
    const longestCommonPrefix = (s: string, t: string): string => {
        let endIdx = 0
        for (let i = 0; i < s.length && i < t.length; i++) {
            if (s[i] !== t[i]) {
                break
            }
            endIdx = i + 1
        }
        return s.slice(0, endIdx)
    }

    const origWords: string[] = []
    for (const chunk of query.split(/\W+/)) {
        if (chunk.trim().length === 0) {
            continue
        }
        origWords.push(...winkUtils.string.tokenize0(chunk))
    }
    const filteredWords = winkUtils.tokens.removeWords(origWords) as string[]
    const terms: { [stem: string]: Term } = {}
    for (const word of filteredWords) {
        const stem = winkUtils.string.stem(word)
        if (terms[stem]) {
            terms[stem].originals.push(word)
            terms[stem].count++
        } else {
            terms[stem] = {
                stem,
                originals: [word],
                prefix: longestCommonPrefix(word.toLowerCase(), stem),
                count: 1,
            }
        }
    }
    return [...Object.values(terms)]
}

export class LocalKeywordContextFetcher implements KeywordContextFetcher {
    constructor(private rgPath: string, private editor: Editor) {}

    public async getContext(query: string, numResults: number): Promise<KeywordContextFetcherResult[]> {
        console.log('fetching keyword matches')
        const rootPath = this.editor.getWorkspaceRootPath()
        if (!rootPath) {
            return []
        }

        const filesnamesWithScores = await this.fetchKeywordFiles(rootPath, query)
        const top10 = filesnamesWithScores.slice(0, numResults)
        const messagePairs = await Promise.all(
            top10.map(async ({ filename }) => {
                const uri = vscode.Uri.file(path.join(rootPath, filename))
                const content = (await vscode.workspace.openTextDocument(uri)).getText()
                return { fileName: filename, content }
            })
        )
        return messagePairs.reverse().flat()
    }

    public async getSearchContext(query: string, numResults: number): Promise<KeywordContextFetcherResult[]> {
        console.log('fetching keyword context')
        const rootPath = this.editor.getWorkspaceRootPath()
        if (!rootPath) {
            return []
        }

        const stems = userQueryToKeywordQuery(query)
            .map(t => (t.prefix.length < 4 ? t.originals[0] : t.prefix))
            .join('|')

        const filesnamesWithScores = await this.fetchKeywordFiles(rootPath, query)
        const messagePairs = await Promise.all(
            filesnamesWithScores.slice(0, numResults).map(async ({ filename }) => {
                const uri = vscode.Uri.file(path.join(rootPath, filename))
                const textDocument = await vscode.workspace.openTextDocument(uri)
                const snippet = textDocument.getText()
                const keywordPattern = new RegExp(stems, 'g')
                const matches = snippet.match(keywordPattern)
                const keywordIndex = snippet.indexOf(matches ? matches[0] : query)
                // show 5 lines of code only
                const startLine = Math.max(0, textDocument.positionAt(keywordIndex).line - 2)
                const endLine = startLine + 5
                const content = textDocument.getText(new vscode.Range(startLine, 0, endLine, 0))

                return { fileName: filename, content }
            })
        )
        return messagePairs.flat()
    }

    private async fetchFileStats(
        terms: Term[],
        rootPath: string
    ): Promise<{ [filename: string]: { bytesSearched: number } }> {
        const regexQuery = `\\b${regexForTerms(...terms)}`
        const proc = spawn(this.rgPath, ['-i', ...fileExtRipgrepParams, '--json', regexQuery, './'], {
            cwd: rootPath,
            stdio: ['ignore', 'pipe', process.stderr],
        })
        const fileTermCounts: {
            [filename: string]: {
                bytesSearched: number
            }
        } = {}
        await new Promise<void>((resolve, reject) => {
            try {
                proc.stdout
                    .pipe(StreamValues.withParser())
                    .on('data', data => {
                        try {
                            switch (data.value.type) {
                                case 'end':
                                    if (!fileTermCounts[data.value.data.path.text]) {
                                        fileTermCounts[data.value.data.path.text] = {} as any
                                    }
                                    fileTermCounts[data.value.data.path.text].bytesSearched =
                                        data.value.data.stats.bytes_searched

                                    break
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
        const termFileCountsArr: { fileCounts: { [filename: string]: number }; filesSearched: number }[] =
            await Promise.all(
                queryTerms.map(async term => {
                    const out = await new Promise<string>((resolve, reject) => {
                        execFile(
                            this.rgPath,
                            [
                                '-i',
                                ...fileExtRipgrepParams,
                                '--count-matches',
                                '--stats',
                                `\\b${regexForTerms(term)}`,
                                './',
                            ],
                            {
                                cwd: rootPath,
                                maxBuffer: 1024 * 1024 * 1024,
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
                            const count = parseInt(terms[1], 10)
                            fileCounts[terms[0]] = count
                        } catch {
                            console.error(`could not parse count from ${terms[1]}`)
                        }
                    }
                    return { fileCounts, filesSearched }
                })
            )

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
        const query = userQueryToKeywordQuery(rawQuery)

        const fileMatchesPromise = this.fetchFileMatches(query, rootPath)
        const fileStatsPromise = this.fetchFileStats(query, rootPath)

        const fileMatches = await fileMatchesPromise
        const fileStats = await fileStatsPromise

        const { fileTermCounts, termTotalFiles, totalFiles } = fileMatches
        const idfDict = idf(termTotalFiles, totalFiles)

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
            .map(([filename, fileTermCounts]) => {
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

                return {
                    filename,
                    cosineScore,
                    termCounts: fileTermCounts,
                    tfVec,
                    idfDict,
                    score,
                    scoreComponents,
                }
            })
            .sort(({ score: score1 }, { score: score2 }) => score2 - score1)

        return filenamesWithScores
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
