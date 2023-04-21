// SCIP index as object for https://sourcegraph.com/github.com/codecov/sourcegraph-codecov@92d2f701f935b7ce3c3504ab893f808643e6eb24/-/blob/src/insights.ts
import scipIndex from './index.json'
import { Descriptor_Suffix, ISymbol, parseSymbol, SCIPDocument } from './SymbolParser'

const cache = new Map<string, Map<string, ISymbol>>()

function getSymbolsMap(document: SCIPDocument): Map<string, ISymbol> {
    const existing = cache.get(document.relative_path)
    if (existing) {
        return existing
    }

    const symbolsMap = new Map<string, ISymbol>()
    for (const { symbol } of document.occurrences) {
        const parsedSymbol = parseSymbol(symbol)
        if (parsedSymbol instanceof Error) {
            continue
        }
        symbolsMap.set(symbol, parsedSymbol)
    }

    cache.set(document.relative_path, symbolsMap)
    return symbolsMap
}

export function getNamespaceSymbols(document: SCIPDocument): string[] {
    const symbolsMap = getSymbolsMap(document)
    const result = new Set()

    for (const { symbol } of document.occurrences) {
        const parsedSymbol = symbolsMap.get(symbol)

        if (parsedSymbol?.descriptors.every(d => d.suffix === Descriptor_Suffix.Namespace)) {
            result.add(symbol)
        }
    }

    return Array.from(result) as string[]
}

/**
 * Returns scip document as object, see [Document.toObject](https://sourcegraph.com/github.com/sourcegraph/scip@d62dfc4d962f4ac975429e0fbb0ebdda25b46503/-/blob/bindings/typescript/scip.ts?L614-634).
 */
export function getDocument(path: string): SCIPDocument | undefined {
    // TODO: replace with API call
    return scipIndex.documents.find(d => d.relative_path === path) as SCIPDocument | undefined
}
