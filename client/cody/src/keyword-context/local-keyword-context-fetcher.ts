import { execFile, spawn } from 'child_process'
import * as path from 'path'

import { removeStopwords } from 'stopword'
import StreamValues from 'stream-json/streamers/StreamValues'
import * as vscode from 'vscode'

import { KeywordContextFetcher, KeywordContextFetcherResult } from '.'

const fileExtRipgrepParams = ['-Tmarkdown', '-Tyaml', '-Tjson', '-g', '!*.lock']

export class LocalKeywordContextFetcher implements KeywordContextFetcher {
    constructor(private rgPath: string) {}

    public async getContext(query: string, numResults: number): Promise<KeywordContextFetcherResult[]> {
        console.log('fetching keyword matches')
        const rootPath = getRootPath()
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

    private async fetchFileStats(
        keywords: string[],
        rootPath: string
    ): Promise<{ [filename: string]: { bytesSearched: number } }> {
        const regexQuery = `\\b(?:${keywords.join('|')})`
        const proc = spawn(this.rgPath, [...fileExtRipgrepParams, '--json', regexQuery, './'], {
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
        keywords: string[],
        rootPath: string
    ): Promise<{
        totalFiles: number
        fileTermCounts: { [filename: string]: { [term: string]: number } }
        termTotalFiles: { [term: string]: number }
    }> {
        const termFileCountsArr: { fileCounts: { [filename: string]: number }; filesSearched: number }[] =
            await Promise.all(
                keywords.map(async term => {
                    const out = await new Promise<string>((resolve, reject) => {
                        execFile(
                            this.rgPath,
                            [...fileExtRipgrepParams, '--count-matches', '--stats', `\\b${term}`, './'],
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

        const fileTermCounts: { [filename: string]: { [term: string]: number } } = {}
        const termTotalFiles: { [term: string]: number } = {}
        for (let i = 0; i < keywords.length; i++) {
            const term = keywords[i]
            const fileCounts = termFileCountsArr[i].fileCounts
            termTotalFiles[term] = Object.keys(fileCounts).length

            for (const [filename, count] of Object.entries(fileCounts)) {
                if (!fileTermCounts[filename]) {
                    fileTermCounts[filename] = {}
                }
                fileTermCounts[filename][term] = count
            }
        }

        return {
            totalFiles: totalFilesSearched,
            termTotalFiles,
            fileTermCounts,
        }
    }

    private async fetchKeywordFiles(rootPath: string, query: string): Promise<{ filename: string; score: number }[]> {
        const terms = query.split(/\W+/)
        // TODO: Stemming using the `natural` package was introducing failing licensing checks. Find a replacement stemming package.
        const stemmedTerms = terms.map(term => escapeRegex(term))
        // unique stemmed keywords, our representation of the user query
        const filteredTerms = Array.from(new Set(removeStopwords(stemmedTerms).filter(term => term.length >= 3)))

        const fileMatchesPromise = this.fetchFileMatches(filteredTerms, rootPath)
        const fileStatsPromise = this.fetchFileStats(filteredTerms, rootPath)

        const fileMatches = await fileMatchesPromise
        const fileStats = await fileStatsPromise

        const { fileTermCounts, termTotalFiles, totalFiles } = fileMatches
        const idfDict = idf(termTotalFiles, totalFiles)

        const queryTf = tf(
            filteredTerms,
            Object.fromEntries(filteredTerms.map(term => [term, 1])),
            filteredTerms.map(t => t.length).reduce((a, b) => a + b + 1, 0)
        )
        const queryVec = tfidf(filteredTerms, queryTf, idfDict)
        const filenamesWithScores = Object.entries(fileTermCounts)
            .map(([filename, termCounts]) => {
                if (fileStats[filename] === undefined) {
                    throw new Error(`filename ${filename} missing from fileStats`)
                }
                const tfVec = tf(filteredTerms, termCounts, fileStats[filename].bytesSearched)
                const tfidfVec = tfidf(filteredTerms, tfVec, idfDict)
                const cosineScore = cosine(tfidfVec, queryVec)
                let { score, scoreComponents } = idfLogScore(filteredTerms, termCounts, idfDict)

                const b = fileStats[filename].bytesSearched
                if (b > 10000) {
                    score *= 0.1 // downweight very large files
                }

                return {
                    filename,
                    cosineScore,
                    termCounts,
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

function getRootPath(): string | null {
    const uri = vscode.window.activeTextEditor?.document.uri
    if (uri) {
        const wsFolder = vscode.workspace.getWorkspaceFolder(uri)
        if (wsFolder) {
            return wsFolder.uri.fsPath
        }
    }

    if (vscode.workspace.workspaceFolders && vscode.workspace.workspaceFolders.length >= 1) {
        return vscode.workspace.workspaceFolders[0].uri.fsPath
    }
    return null
}

async function fetchFileStats(
    keywords: string[],
    rootPath: string
): Promise<{ [filename: string]: { bytesSearched: number } }> {
    const regexQuery = `\\b(?:${keywords.join('|')})`
    const proc = spawn('rg', [...fileExtRipgrepParams, '--json', regexQuery, './'], {
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

async function fetchFileMatches(
    keywords: string[],
    rootPath: string
): Promise<{
    totalFiles: number
    fileTermCounts: { [filename: string]: { [term: string]: number } }
    termTotalFiles: { [term: string]: number }
}> {
    const termFileCountsArr: { fileCounts: { [filename: string]: number }; filesSearched: number }[] =
        await Promise.all(
            keywords.map(async term => {
                const out = await new Promise<string>((resolve, reject) => {
                    execFile(
                        'rg',
                        [...fileExtRipgrepParams, '--count-matches', '--stats', `\\b${term}`, './'],
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

    const fileTermCounts: { [filename: string]: { [term: string]: number } } = {}
    const termTotalFiles: { [term: string]: number } = {}
    for (let i = 0; i < keywords.length; i++) {
        const term = keywords[i]
        const fileCounts = termFileCountsArr[i].fileCounts
        termTotalFiles[term] = Object.keys(fileCounts).length

        for (const [filename, count] of Object.entries(fileCounts)) {
            if (!fileTermCounts[filename]) {
                fileTermCounts[filename] = {}
            }
            fileTermCounts[filename][term] = count
        }
    }

    return {
        totalFiles: totalFilesSearched,
        termTotalFiles,
        fileTermCounts,
    }
}

export async function fetchKeywordFiles(
    rootPath: string,
    query: string
): Promise<{ filename: string; score: number }[]> {
    const terms = query.split(/\W+/)
    // TODO: Stemming using the `natural` package was introducing failing licensing checks. Find a replacement stemming package.
    const stemmedTerms = terms.map(term => escapeRegex(term))
    // unique stemmed keywords, our representation of the user query
    const filteredTerms = Array.from(new Set(removeStopwords(stemmedTerms).filter(term => term.length >= 3)))

    const fileMatchesPromise = fetchFileMatches(filteredTerms, rootPath)
    const fileStatsPromise = fetchFileStats(filteredTerms, rootPath)

    const fileMatches = await fileMatchesPromise
    const fileStats = await fileStatsPromise

    const { fileTermCounts, termTotalFiles, totalFiles } = fileMatches
    const idfDict = idf(termTotalFiles, totalFiles)

    const queryTf = tf(
        filteredTerms,
        Object.fromEntries(filteredTerms.map(term => [term, 1])),
        filteredTerms.map(t => t.length).reduce((a, b) => a + b + 1, 0)
    )
    const queryVec = tfidf(filteredTerms, queryTf, idfDict)
    const filenamesWithScores = Object.entries(fileTermCounts)
        .map(([filename, termCounts]) => {
            if (fileStats[filename] === undefined) {
                throw new Error(`filename ${filename} missing from fileStats`)
            }
            const tfVec = tf(filteredTerms, termCounts, fileStats[filename].bytesSearched)
            const tfidfVec = tfidf(filteredTerms, tfVec, idfDict)
            const cosineScore = cosine(tfidfVec, queryVec)
            let { score, scoreComponents } = idfLogScore(filteredTerms, termCounts, idfDict)

            const b = fileStats[filename].bytesSearched
            if (b > 10000) {
                score *= 0.1 // downweight very large files
            }

            return {
                filename,
                cosineScore,
                termCounts,
                tfVec,
                idfDict,
                score,
                scoreComponents,
            }
        })
        .sort(({ score: score1 }, { score: score2 }) => score2 - score1)

    return filenamesWithScores
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

function tf(terms: string[], termCounts: { [term: string]: number }, fileSize: number): number[] {
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
