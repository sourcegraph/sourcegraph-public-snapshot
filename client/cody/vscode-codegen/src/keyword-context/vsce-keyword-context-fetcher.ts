import nlp from 'compromise/two'
import * as vscode from 'vscode'

import { getContextMessageWithResponse, populateCodeContextTemplate } from '../chat/prompt'
import { Message } from '../sourcegraph-api'

import { KeywordContextFetcher } from '.'

interface TermFileCounts {
    fileCounts: { [filename: string]: number }
    filesSearched: number
}

interface FileTermCounts {
    [filename: string]: { [term: string]: number }
}

interface TermTotalFiles {
    [term: string]: number
}

export class VSCEKeywordContextFetcher implements KeywordContextFetcher {
    constructor() {}

    public async getContextMessages(query: string): Promise<Message[]> {
        console.log('fetching keyword matches with VSCE')
        const rootPath = this.getRootPath()
        if (!rootPath) return []
        const filesnamesWithScores = await this.fetchKeywordFiles(rootPath, query)
        const top10 = filesnamesWithScores.slice(0, 10)
        const messagePairs = await Promise.all(
            top10.map(async ({ filename }) => {
                const uri = vscode.Uri.file(filename)
                const text = (await vscode.workspace.openTextDocument(uri)).getText()
                // Remove root path from file path
                const file = filename.replace(`${rootPath}/`, '')
                const messageText = populateCodeContextTemplate(text, file)
                return getContextMessageWithResponse(messageText, file)
            })
        )
        return messagePairs.reverse().flat()
    }

    private async fetchKeywordFiles(rootPath: string, query: string): Promise<{ filename: string; score: number }[]> {
        // unique stemmed keywords, our representation of the user query
        const terms = nlp(query).text({ keepPunct: false, trim: true })
        const stemmedTerms = nlp(terms).match('#Noun').not('function').sort('length').out('array')
        const filteredTerms = stemmedTerms.filter((term: string) => term.length > 2)
        const filenamesWithScores = await this.fetchSymbolMatches(filteredTerms, rootPath)
        // If symbol search is empty or less than 10, look for keywords in current workspace
        if (filenamesWithScores.length < 11) {
            const filenamesWithScoresByFile = await this.fetchFileMatches(filteredTerms)
            return [...filenamesWithScores, ...filenamesWithScoresByFile]
        } else {
            return filenamesWithScores
        }
    }

    private getFileNamesWithScores(
        fileTermCounts: FileTermCounts,
        termTotalFiles: TermTotalFiles,
        totalFiles: number,
        filteredTerms
    ): { filename: string; score: number }[] {
        const idfDict = this.idf(termTotalFiles, totalFiles)
        const filenamesWithScores = Object.entries(fileTermCounts)
            .map(([filename, termCounts]) => {
                const { score, scoreComponents } = this.idfLogScore(filteredTerms, termCounts, idfDict)
                return {
                    filename,
                    score,
                    scoreComponents,
                }
            })
            .sort(({ score: score1 }, { score: score2 }) => score2 - score1)

        return filenamesWithScores
    }

    // Get file matches from symbol search through vs code API
    private async fetchSymbolMatches(
        keywords: string[],
        rootPath: string
    ): Promise<{ filename: string; score: number }[]> {
        // The total number of files searched across all keywords
        let totalFilesSearched = 0
        const wsRootPath = rootPath.replace('file://', '')
        const termFileCountsArr: TermFileCounts[] = await Promise.all(
            keywords.map(async term => {
                // Get a list of files that match the term as symbol
                const symbols = await this.nativeSymbolSearcher(term)
                // Mapping filenames to match counts
                const fileCounts: { [filename: string]: number } = {}

                symbols.forEach(symbol => {
                    const symbolPath = symbol.location.uri.fsPath
                    // filter files that are not in current workspace
                    if (symbolPath.includes(wsRootPath)) {
                        fileCounts[symbolPath] = (fileCounts[symbolPath] || 0) + 1
                    }
                })
                // Set the length of fileCounts as files searched
                const filesSearched = Object.entries(fileCounts).length || 0
                totalFilesSearched += filesSearched
                return { fileCounts, filesSearched }
            })
        )
        const termsCount = this.termCounter(keywords, termFileCountsArr)

        const filenamesWithScores = this.getFileNamesWithScores(
            termsCount.fileTermCounts,
            termsCount.termTotalFiles,
            totalFilesSearched,
            keywords
        )

        return filenamesWithScores
    }

    private async fetchFileMatches(keywords: string[]): Promise<{ filename: string; score: number }[]> {
        const matchedFiles: { filename: string; score: number }[] = []
        const excludePattern = this.generateExcludePattern()
        const foundFiles = await vscode.workspace.findFiles(`**/{${keywords.join(',')}}`, excludePattern, 10000)
        const keywordRegex = new RegExp(keywords.join('|'), 'g')
        for (const file of foundFiles) {
            let score = 0
            const content = await vscode.workspace.fs.readFile(file)
            const fileContent = content.toString()
            if (!file.path.includes('node_modules') && keywordRegex.test(fileContent)) {
                const matches = fileContent.match(keywordRegex)
                const filSet = new Set(matches)
                score += [...filSet].length
            }

            matchedFiles.push({
                filename: file.path,
                score,
            })
        }

        return matchedFiles.sort(({ score: score1 }, { score: score2 }) => score2 - score1)
    }

    private termCounter(
        keywords: string[],
        termFileCountsArr: TermFileCounts[]
    ): {
        termTotalFiles: TermTotalFiles
        fileTermCounts: FileTermCounts
    } {
        // Map filenames to an object mapping terms to match counts for that filename
        const fileTermCounts: FileTermCounts = {}
        // Map terms to the number of files they matched
        const termTotalFiles: TermTotalFiles = {}
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
            termTotalFiles,
            fileTermCounts,
        }
    }

    public getRootPath(): string | null {
        const docUri = vscode.window.activeTextEditor?.document.uri
        const wsFolderUri = docUri ? vscode.workspace.getWorkspaceFolder(docUri)?.uri.fsPath : null
        const wsRootUri = vscode.workspace.workspaceFolders?.[0]?.uri?.fsPath || null
        return wsFolderUri === undefined ? wsRootUri : wsFolderUri
    }

    private async nativeSymbolSearcher(term: string): Promise<vscode.SymbolInformation[]> {
        const command = 'vscode.executeWorkspaceSymbolProvider'
        return await vscode.commands.executeCommand<vscode.SymbolInformation[]>(command, term)
    }

    private idfLogScore(
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

    private idf(termTotalFiles: { [term: string]: number }, totalFiles: number): { [term: string]: number } {
        const logTotal = Math.log(totalFiles)
        const e = Object.entries(termTotalFiles).map(([term, count]) => [term, logTotal - Math.log(count)])
        return Object.fromEntries(e)
    }

    // produce string as global pattern for list of files to exclude
    // eg. {**/*node_modules,**/*.json}
    public generateExcludePattern(): string {
        const patterns = this.excludePattern
            .map(exclude => (exclude.startsWith('.') ? `**/*${exclude}` : `**/${exclude}`))
            .join(',')
        return `{${patterns}}`
    }

    private excludePattern = [
        'node_modules',
        'build',
        'dist',
        'out',
        'coverage',
        'node_modules',
        '.logs',
        '.git',
        '.vscode',
        '.css',
        '.json',
        '.txt',
        '.rst',
        '.bazel',
        '.tmp',
        '.log',
        '.zip',
        '.gz',
        '.bz2',
        '.7z',
        '.tgz',
        '.exe',
        '.dmg',
        '.pkg',
        '.deb',
        '.rpm',
        '.jpg',
        '.jpeg',
        '.png',
        '.gif',
        '.bmp',
        '.svg',
        '.mp4',
        '.avi',
        '.mov',
        '.wmv',
        '.m4v',
        '.flv',
        '.mkv',
        '.sql',
        '.golden',
        '.d.*',
        '.*ignore',
        '*.config.*',
        '.csv',
        '.nix',
        'LICENSE',
        '.yaml',
    ]
}
