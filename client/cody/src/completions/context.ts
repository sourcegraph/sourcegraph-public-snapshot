import * as vscode from 'vscode'

import { History } from './history'

const winkUtils = require('wink-nlp-utils')

export interface ReferenceSnippet extends JaccardMatch {
    filename: string
}

export async function getContext(
    currentEditor: vscode.TextEditor,
    history: History,
    targetText: string,
    windowSize: number,
    maxBytes: number
): Promise<ReferenceSnippet[]> {
    const files = await getFiles(currentEditor, history)
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
    let totalBytes = 0
    for (const match of matches) {
        if (totalBytes + match.text.length > maxBytes) {
            break
        }
        context.push(match)
        totalBytes += match.text.length
    }

    return context
}

interface FileContents {
    uri: vscode.Uri
    contents: string
}

async function getFiles(currentEditor: vscode.TextEditor, history: History): Promise<FileContents[]> {
    const files: FileContents[] = []
    const editorTabs = vscode.window.visibleTextEditors
    const curLang = currentEditor.document.languageId
    if (!curLang) {
        return []
    }
    for (const tab of editorTabs) {
        if (tab.document.uri === currentEditor.document.uri) {
            // omit current file
            continue
        }
        if (tab.document.languageId !== curLang) {
            // omit files of other languages
            continue
        }
        files.push({
            uri: tab.document.uri,
            contents: tab.document.getText(),
        })
    }
    const historyFiles = await Promise.all(
        history.lastN(10, curLang, [currentEditor.document.uri, ...files.map(f => f.uri)]).map(async item => {
            const contents = (await vscode.workspace.openTextDocument(item.document.uri)).getText()
            return {
                uri: item.document.uri,
                contents,
            }
        })
    )
    files.push(...historyFiles)
    return files
}

export interface JaccardMatch {
    score: number
    text: string
}

export function jaccardScore(targetWords: Map<string, number>, matchWords: Map<string, number>): number {
    const intersection = [...targetWords.keys()].reduce((acc, word) => {
        return acc + Math.min(targetWords.get(word) || 0, matchWords.get(word) || 0)
    }, 0)
    const unionSet = new Set([...targetWords.keys(), ...matchWords.keys()])

    const union = [...unionSet].reduce((acc, word) => {
        return acc + Math.max(targetWords.get(word) || 0, matchWords.get(word) || 0)
    }, 0)
    if (intersection === 0) {
        return 0
    }
    return intersection / union
}

export function bestJaccardMatch(targetText: string, matchText: string, windowSize: number): JaccardMatch | null {
    const targetWords = getWords(targetText)
    const lines = matchText.split('\n')
    const wordsForEachLine = lines.map(line => getWords(line))

    const windowWords = new Map<string, number>()
    for (let i = 0; i < Math.min(windowSize, lines.length); i++) {
        for (const wordInThisLine of wordsForEachLine[i].keys()) {
            windowWords.set(
                wordInThisLine,
                (windowWords.get(wordInThisLine) || 0) + (wordsForEachLine[i].get(wordInThisLine) || 0)
            )
        }
    }
    const bothWords = new Map<string, number>()
    for (const word of targetWords.keys()) {
        bothWords.set(word, Math.min(targetWords.get(word) || 0, windowWords.get(word) || 0))
    }

    let bestScore = jaccardScore(targetWords, windowWords)
    let bestWindow = [0, Math.min(windowSize, lines.length)]
    for (let i = 0; i < wordsForEachLine.length - windowSize; i++) {
        for (const word of wordsForEachLine[i].keys()) {
            windowWords.set(word, (windowWords.get(word) || 0) - (wordsForEachLine[i].get(word) || 0))
            if (windowWords.get(word)! < 0) {
                windowWords.set(word, 0)
            }
            bothWords.set(word, (bothWords.get(word) || 0) - (wordsForEachLine[i].get(word) || 0))
            if (bothWords.get(word)! < 0) {
                bothWords.set(word, 0)
            }
        }
        for (const word of wordsForEachLine[i + windowSize].keys()) {
            if (windowWords.get(word) === undefined) {
                windowWords.set(word, 0)
            }
            windowWords.set(word, (windowWords.get(word) || 0) + (wordsForEachLine[i + windowSize].get(word) || 0))
            if (targetWords.get(word)! > 0) {
                bothWords.set(
                    word,
                    Math.min(
                        (bothWords.get(word) || 0) + (wordsForEachLine[i + windowSize].get(word) || 0),
                        targetWords.get(word) || 0
                    )
                )
            }
        }
        const score = jaccardScore(targetWords, windowWords)
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
    let frequencyCounter = new Map<string, number>()
    const words = winkUtils.string.tokenize0(s)

    const filteredWords = winkUtils.tokens.removeWords(words)
    const stems = winkUtils.tokens.stem(filteredWords)
    for (const stem of stems) {
        frequencyCounter.set(stem, (frequencyCounter.get(stem) || 0) + 1)
    }
    return frequencyCounter
}
