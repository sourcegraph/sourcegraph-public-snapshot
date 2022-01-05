import { GolangIndexer } from './golang'
import { Indexer } from './lsif'

const all: Indexer[] = [new GolangIndexer()]
// const all: Indexer[] = []
export function syntaxHighlight(filePath: string, text: string): Promise<string> | undefined {
    const indexer = indexerForFilePath(filePath)
    // console.log({ indexer, filePath, matches: new GolangIndexer().matchesFilePath(filePath) })
    return indexer?.highlight({ text, path: filePath })
}

export function indexerForFilePath(filePath: string): Indexer | undefined {
    for (const indexer of all) {
        if (indexer.matchesFilePath(filePath)) {
            return indexer
        }
    }
    return undefined
}
