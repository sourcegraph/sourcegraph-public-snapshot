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

import type { SymbolKind } from '../graphql-operations'

export const getSymbolIconSVGPath = (kind: SymbolKind): string => {
    switch (kind) {
        case 'FILE': {
            return mdiFileDocument
        }
        case 'MODULE': {
            return mdiCodeBraces
        }
        case 'NAMESPACE': {
            return mdiWeb
        }
        case 'PACKAGE': {
            return mdiPackage
        }
        case 'CLASS': {
            return mdiSitemap
        }
        case 'METHOD': {
            return mdiCubeOutline
        }
        case 'PROPERTY': {
            return mdiWrench
        }
        case 'FIELD': {
            return mdiTextBox
        }
        case 'CONSTRUCTOR': {
            return mdiCubeOutline
        }
        case 'ENUM': {
            return mdiNumeric
        }
        case 'INTERFACE': {
            return mdiLink
        }
        case 'FUNCTION': {
            return mdiFunction
        }
        case 'VARIABLE': {
            return mdiCube
        }
        case 'CONSTANT': {
            return mdiPiBox
        }
        case 'STRING': {
            return mdiCodeString
        }
        case 'NUMBER': {
            return mdiPound
        }
        case 'BOOLEAN': {
            return mdiMatrix
        }
        case 'ARRAY': {
            return mdiCodeArray
        }
        case 'OBJECT': {
            return mdiDrawingBox
        }
        case 'KEY': {
            return mdiKey
        }
        case 'NULL': {
            return mdiNull
        }
        case 'ENUMMEMBER': {
            return mdiNumeric
        }
        case 'STRUCT': {
            return mdiPillar
        }
        case 'EVENT': {
            return mdiTimetable
        }
        case 'OPERATOR': {
            return mdiCodeNotEqual
        }
        case 'TYPEPARAMETER': {
            return mdiCube
        }
        case 'UNKNOWN':
        default: {
            return mdiShape
        }
    }
}
