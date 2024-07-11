import { SymbolKind } from '$lib/graphql-types'

/**
 * Returns a friendly symbol name for the given symbol kind.
 *
 * @param kind The symbol kind.
 * @returns The friendly symbol name.
 */
export function humanReadableSymbolKind(kind: SymbolKind | string): string {
    switch (kind) {
        case SymbolKind.TYPEPARAMETER:
            return 'Type parameter'
        case SymbolKind.ENUMMEMBER:
            return 'Enum member'
        default:
            return kind.charAt(0).toUpperCase() + kind.slice(1).toLowerCase()
    }
}
