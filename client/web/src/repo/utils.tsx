import { groupBy } from 'lodash'

import { RepositoryType, SymbolNodeFields, type GitCommitFields } from '../graphql-operations'
import { parseBrowserRepoURL } from '../util/url'

import { CodeHostType } from './constants'

export const isPerforceChangelistMappingEnabled = (): boolean =>
    window.context.experimentalFeatures.perforceChangelistMapping === 'enabled'

export const isPerforceDepotSource = (sourceType: string): boolean => sourceType === RepositoryType.PERFORCE_DEPOT

export const getRefType = (sourceType: RepositoryType | string): string =>
    isPerforceDepotSource(sourceType) ? 'changelist' : 'commit'

export const getCanonicalURL = (sourceType: RepositoryType | string, node: GitCommitFields): string =>
    isPerforceChangelistMappingEnabled() && isPerforceDepotSource(sourceType) && node.perforceChangelist
        ? node.perforceChangelist.canonicalURL
        : node.canonicalURL

export const getInitialSearchTerm = (repo: string): string => {
    const r = repo.split('/')
    return r.at(-1)?.trim() ?? ''
}

export const stringToCodeHostType = (codeHostType: string): CodeHostType => {
    switch (codeHostType) {
        case 'github': {
            return CodeHostType.GITHUB
        }
        case 'gitlab': {
            return CodeHostType.GITLAB
        }
        case 'bitbucketCloud': {
            return CodeHostType.BITBUCKETCLOUD
        }
        case 'gitolite': {
            return CodeHostType.GITOLITE
        }
        case 'awsCodeCommit': {
            return CodeHostType.AWSCODECOMMIT
        }
        case 'azureDevOps': {
            return CodeHostType.AZUREDEVOPS
        }
        default: {
            return CodeHostType.OTHER
        }
    }
}

/**
 * When searching symbols, results may contain only child symbols without their parents
 * (e.g. when searching for "bar", a class named "Foo" with a method named "bar" will
 * return "bar" as a result, and "bar" will say that "Foo" is its parent).
 * The placeholder symbols exist to show the hierarchy of the results, but these placeholders
 * are not interactive (cannot be clicked to navigate) and don't have any other information.
 */
export interface SymbolPlaceholder {
    __typename: 'SymbolPlaceholder'
    name: string
}

export type SymbolWithChildren = (SymbolNodeFields | SymbolPlaceholder) & { children: SymbolWithChildren[] }

/** Organize symbols hierarchically and sort them by line number. */
export const hierarchyOf = (symbols: SymbolNodeFields[]): SymbolWithChildren[] => {
    // Group all symbols by their parent symbol. Symbols with no children will not be included here.
    const containerGroups = groupBy(symbols, sym => sym.containerName)

    // Sort inner-most symbols first to fold this structure bottom-up.
    // This way, by the time we reach any parent symbol, all its children will have been collected.
    symbols.sort(longerPathFirst)

    // Collect children for each of the initial symbols and organize them hierarchically.
    const topLevelSymbols: Record<string, { name: string; children: SymbolWithChildren[] }> = {}

    // As far as I understand there should be at most 1 symbol with an empty `containerName`.
    // In that case it's missing from `containerGroups`, even though it parents all other symbols.
    let topMost: SymbolWithChildren | undefined

    symbols.forEach(sym => {
        if (!sym.containerName) {
            // We've got the top-most symbol
            topMost = {
                ...sym,
                children: [],
            }
            return
        }
        const parentName = symbolName(sym.containerName)

        /** Create parent container in `topLevelSymbols` if not exist and add children to it. */
        function addChildren(symbol: SymbolNodeFields, children: SymbolWithChildren[]) {
            if (!(parentName in topLevelSymbols)) {
                topLevelSymbols[parentName] = {
                    name: parentName,
                    children: [],
                }
            }
            topLevelSymbols[parentName].children.push({
                ...symbol,
                children,
            })
        }

        const children = containerGroups[fullName(sym)]
        if (!children || !children.length) {
            addChildren(sym, [])
            return
        }
        const thisSymbol = topLevelSymbols[sym.name]
        addChildren(sym, thisSymbol.children.sort(lineNumbersAsc))
        // By the end of the loop, only symbols with no parent containers should remain in `topLevelSymbols`.
        delete topLevelSymbols[sym.name]
    })

    // The top-level symbols are SymbolPlaceholders by definition -- they do not have a parent container.
    const result: SymbolWithChildren[] = []
    for (const key in topLevelSymbols) {
        result.push({
            __typename: 'SymbolPlaceholder',
            name: key,
            children: topLevelSymbols[key].children.sort(lineNumbersAsc),
        })
    }

    if (topMost) {
        topMost.children.push(...result)
        return [topMost]
    }
    return result
}

/**
 * Symbol's full name consists of its `containerName` and `name`. "Full names" are not necessarily unique
 * within a symbol tree, as some language features (e.g. method overloading) permit repeated identifiers.
 */
const fullName = (symbol: SymbolNodeFields | SymbolPlaceholder): string => {
    return `${symbol.__typename === 'Symbol' && symbol.containerName ? symbol.containerName + '.' : ''}${symbol.name}`
}

/** Extract the right-most identifier from the full name, e.g. `package.Class.[[fieldA]]` */
const symbolName = (full: string): string => {
    return full.split('.').at(-1) || full
}

/**
 * Order symbols in the order of ascending line numbers -- the way they appear in the file.
 *
 * @example
 * const symbols: SymbolWithChildren[] = [...]
 * symbols.sort(lineNumbersAsc)
 */
function lineNumbersAsc(s1: SymbolWithChildren, s2: SymbolWithChildren): number {
    const url1 = parseBrowserRepoURL((s1 as SymbolNodeFields).url),
        url2 = parseBrowserRepoURL((s2 as SymbolNodeFields).url)
    if (!url1.position || !url2.position) {
        return 0
    }
    return url1.position.line - url2.position.line
}

/**
 * Order symbols in the order of descending path length, i.e.
 * `['package.Foo.fieldA', 'package.Bar.fieldB', 'package.Foo', 'package.Bar', 'package']`.
 *
 * @example
 * const symbols: SymbolNodeFields[] = [...]
 * symbols.sort(longerPathFirst)
 */
function longerPathFirst(s1: SymbolNodeFields, s2: SymbolNodeFields): number {
    const c1 = s1.containerName,
        c2 = s2.containerName

    // Is path to s1 shorter than path to s2?
    switch (true) {
        case !c1 && !c2:
            return 0
        case !c1: // no, s1 has no container => s1 comes *after* s2
            return 1
        case !c2: // yes, s2 has no container => s1 comes *before* s2
            return -1
        default:
            return c2!.split('.').length - c1!.split('.').length // longer path first
    }
}
