// TODO: Reuse code from main app

import {
    mdiCodeArray,
    mdiCodeBraces,
    mdiCodeNotEqual,
    mdiCodeString,
    mdiCube,
    mdiCubeOutline,
    mdiDrawingBox,
    mdiFileDocument,
    mdiFunction,
    mdiKey,
    mdiLink,
    mdiMatrix,
    mdiNull,
    mdiNumeric,
    mdiPackage,
    mdiPiBox,
    mdiPillar,
    mdiPound,
    mdiShape,
    mdiSitemap,
    mdiTextBox,
    mdiTimetable,
    mdiWeb,
    mdiWrench,
} from '@mdi/js'

import { SymbolKind } from '$lib/graphql-types'

const symbolKindIconMap: Record<SymbolKind, string> = {
    [SymbolKind.FILE]: mdiFileDocument,
    [SymbolKind.MODULE]: mdiCodeBraces,
    [SymbolKind.NAMESPACE]: mdiWeb,
    [SymbolKind.PACKAGE]: mdiPackage,
    [SymbolKind.CLASS]: mdiSitemap,
    [SymbolKind.METHOD]: mdiCubeOutline,
    [SymbolKind.PROPERTY]: mdiWrench,
    [SymbolKind.FIELD]: mdiTextBox,
    [SymbolKind.CONSTRUCTOR]: mdiCubeOutline,
    [SymbolKind.ENUM]: mdiNumeric,
    [SymbolKind.INTERFACE]: mdiLink,
    [SymbolKind.FUNCTION]: mdiFunction,
    [SymbolKind.VARIABLE]: mdiCube,
    [SymbolKind.CONSTANT]: mdiPiBox,
    [SymbolKind.STRING]: mdiCodeString,
    [SymbolKind.NUMBER]: mdiPound,
    [SymbolKind.BOOLEAN]: mdiMatrix,
    [SymbolKind.ARRAY]: mdiCodeArray,
    [SymbolKind.OBJECT]: mdiDrawingBox,
    [SymbolKind.KEY]: mdiKey,
    [SymbolKind.NULL]: mdiNull,
    [SymbolKind.ENUMMEMBER]: mdiNumeric,
    [SymbolKind.STRUCT]: mdiPillar,
    [SymbolKind.EVENT]: mdiTimetable,
    [SymbolKind.OPERATOR]: mdiCodeNotEqual,
    [SymbolKind.TYPEPARAMETER]: mdiCube,
    [SymbolKind.UNKNOWN]: mdiShape,
}

/**
 * Returns the icon path for a given symbol kind
 */
export function getSymbolIconPath(kind: SymbolKind | string): string {
    if (symbolKindIconMap[kind as SymbolKind]) {
        return symbolKindIconMap[kind as SymbolKind]
    }
    return mdiShape
}

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
