import * as vscode from 'vscode'
import winkUtils from 'wink-nlp-utils'

import { History } from './history'

export interface ReferenceSnippet extends JaccardMatch {
    filename: string
}

export async function getContext(
    currentEditor: vscode.TextEditor,
    history: History,
    targetText: string,
    windowSize: number,
    maxChars: number
): Promise<ReferenceSnippet[]> {
    const files = await getRelevantFiles(currentEditor, history)

    const matches: ReferenceSnippet[] = []
    for (const { uri, contents } of files) {
        const match = bestJaccardMatch(targetText, contents, windowSize)
        if (!match) {
            continue
        }
        matches.push({ filename: uri.fsPath, ...match })
    }

    matches.sort((a, b) => b.score - a.score)

    const context: ReferenceSnippet[] = []
    let totalChars = 0
    for (const match of matches) {
        if (totalChars + match.text.length > maxChars) {
            break
        }
        context.push(match)
        totalChars += match.text.length
    }

    return context
}

interface FileContents {
    uri: vscode.Uri
    contents: string
}

/**
 * Loads all relevant files for for a given text editor. Relevant files are defined as:
 *
 * - All currently open tabs matching the same language
 * - The last 10 files that were edited matching the same language
 *
 * For every file, we will load up to 10.000 lines to avoid OOMing when working with very large
 * files.
 */
async function getRelevantFiles(currentEditor: vscode.TextEditor, history: History): Promise<FileContents[]> {
    const files: FileContents[] = []

    const curLang = currentEditor.document.languageId
    if (!curLang) {
        return []
    }

    function addDocument(document: vscode.TextDocument): void {
        if (document.uri === currentEditor.document.uri) {
            // omit current file
            return
        }
        if (document.languageId !== curLang) {
            // TODO(beyang): handle JavaScript <-> TypeScript and verify this works for C header files
            // omit files of other languages
            return
        }

        // TODO(philipp-spiess): Find out if we have a better approach to truncate very large files.
        const endLine = Math.min(document.lineCount, 10_000)
        const range = new vscode.Range(0, 0, endLine, 0)

        files.push({
            uri: document.uri,
            contents: document.getText(range),
        })
    }

    const documents = vscode.workspace.textDocuments
    for (const document of documents) {
        if (document.fileName.endsWith('.git')) {
            // The VS Code API returns fils with the .git suffix for every open file
            continue
        }
        addDocument(document)
    }

    await Promise.all(
        history.lastN(10, curLang, [currentEditor.document.uri, ...files.map(f => f.uri)]).map(async item => {
            try {
                const document = await vscode.workspace.openTextDocument(item.document.uri)
                addDocument(document)
            } catch (error) {
                console.error(error)
            }
        })
    )
    return files
}

export interface JaccardMatch {
    score: number
    text: string
}

export function jaccardDistance(left: number, right: number, intersection: number): number {
    const union = left + right - intersection
    if (union < 0) {
        throw new Error("intersection can't be greater than the sum of left and right")
    }
    if (union === 0) {
        return 0
    }
    return intersection / union
}

/**
 * Finds the window from matchText with the lowest Jaccard distance from targetText.
 * The Jaccard distance is the ratio of intersection over union, using a bag-of-words-with-count as
 * the representation for text snippet.
 *
 * @param targetText is the text that serves as the target we are trying to find a match for
 * @param matchText is the text we are sliding our window through to find the best match
 * @param windowSize is the size of the match window in number of lines
 * @returns
 */
export function bestJaccardMatch(targetText: string, matchText: string, windowSize: number): JaccardMatch | null {
    const wordCount = (words: Map<string, number>): number => {
        let count = 0
        for (const v of words.values()) {
            count += v
        }
        return count
    }
    // subract the subtrahend bag of words from minuend and return the net change in word count
    const subtract = (minuend: Map<string, number>, subtrahend: Map<string, number>): number => {
        let decrease = 0 // will be non-positive
        for (const [word, count] of subtrahend) {
            const currentCount = minuend.get(word) || 0
            const newCount = Math.max(0, currentCount - count)
            minuend.set(word, newCount)
            decrease += newCount - currentCount
        }
        return decrease
    }
    // add incorporates a new line into window and intersection, updating each, and returns the
    // net increase in size for each
    const add = (
        target: Map<string, number>,
        window: Map<string, number>,
        intersection: Map<string, number>,
        newLine: Map<string, number>
    ): { windowIncrease: number; intersectionIncrease: number } => {
        let windowIncrease = 0
        let intersectionIncrease = 0
        for (const [word, count] of newLine) {
            windowIncrease += count
            window.set(word, (window.get(word) || 0) + count)

            const targetCount = target.get(word) || 0
            if (targetCount > 0) {
                const intersectionCount = intersection.get(word) || 0
                const newIntersectionCount = Math.min(count + intersectionCount, targetCount)
                intersection.set(word, newIntersectionCount)
                intersectionIncrease += newIntersectionCount - intersectionCount
            }
        }
        return { windowIncrease, intersectionIncrease }
    }

    // get the bag-of-words-count dictionary for the target text
    const targetWords = getWords(targetText)
    const targetCount = wordCount(targetWords)

    // split the matchText into lines
    const lines = matchText.split('\n')
    const wordsForEachLine = lines.map(line => getWords(line))

    // initialize the bag of words for the topmost window
    const windowWords = new Map<string, number>()
    for (let i = 0; i < Math.min(windowSize, lines.length); i++) {
        for (const [wordInThisLine, wordInThisLineCount] of wordsForEachLine[i].entries()) {
            windowWords.set(wordInThisLine, (windowWords.get(wordInThisLine) || 0) + wordInThisLineCount)
        }
    }

    let windowCount = wordCount(windowWords)
    // initialize the bag of words for the intersection of the match window and targetText
    const bothWords = new Map<string, number>()
    for (const [word, wordCount] of targetWords.entries()) {
        bothWords.set(word, Math.min(wordCount, windowWords.get(word) || 0))
    }
    let bothCount = wordCount(bothWords)

    // slide our window through matchText, keeping track of the best score and window so far
    let bestScore = jaccardDistance(targetCount, windowCount, bothCount)
    let bestWindow = [0, Math.min(windowSize, lines.length)]
    for (let i = 0; i < wordsForEachLine.length - windowSize; i++) {
        // subtract the words from the line we are scrolling past
        windowCount += subtract(windowWords, wordsForEachLine[i])
        bothCount += subtract(bothWords, wordsForEachLine[i])

        // add the words from the new line our window just slid over
        const { windowIncrease, intersectionIncrease } = add(
            targetWords,
            windowWords,
            bothWords,
            wordsForEachLine[i + windowSize]
        )
        windowCount += windowIncrease
        bothCount += intersectionIncrease

        // compute the jaccard distance between our target text and window
        const score = jaccardDistance(targetCount, windowCount, bothCount)
        if (score > bestScore) {
            bestScore = score
            bestWindow = [i + 1, i + windowSize + 1]
        }
    }

    return {
        score: bestScore,
        text: lines.slice(bestWindow[0], bestWindow[1]).join('\n'),
    }
}

export function getWords(s: string): Map<string, number> {
    const frequencyCounter = new Map<string, number>()
    const words = winkUtils.string.tokenize0(s)

    const filteredWords = winkUtils.tokens.removeWords(words)
    const stems = winkUtils.tokens.stem(filteredWords)
    for (const stem of stems) {
        frequencyCounter.set(stem, (frequencyCounter.get(stem) || 0) + 1)
    }
    return frequencyCounter
}
