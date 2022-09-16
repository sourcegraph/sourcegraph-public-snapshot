import * as React from 'react'

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
import classNames from 'classnames'
import { upperFirst } from 'lodash'

import { Icon, Tooltip } from '@sourcegraph/wildcard'

import { SymbolKind } from '../graphql-operations'

import styles from './SymbolIcon.module.scss'
/**
 * Returns the icon component for a given symbol kind
 */
const getSymbolIconComponent = (kind: SymbolKind): string => {
    switch (kind) {
        case 'FILE':
            return mdiFileDocument
        case 'MODULE':
            return mdiCodeBraces
        case 'NAMESPACE':
            return mdiWeb
        case 'PACKAGENAME':
        case 'SUBPROGSPEC':
        case 'PACKAGE':
            return mdiPackage
        case 'CLASS':
        case 'TYPE':
        case 'SERVICE':
        case 'TYPEDEF':
        case 'UNION':
        case 'SECTION':
        case 'SUBTYPE':
        case 'COMPONENT':
            return mdiSitemap
        case 'METHOD':
        case 'METHODSPEC':
            return mdiCubeOutline
        case 'PROPERTY':
            return mdiWrench
        case 'FIELD':
        case 'MEMBER':
        case 'ANONMEMBER':
        case 'RECORDFIELD':
            return mdiTextBox
        case 'CONSTRUCTOR':
            return mdiCubeOutline
        case 'ENUM':
        case 'ENUMERATOR':
            return mdiNumeric
        case 'INTERFACE':
            return mdiLink
        case 'FUN':
        case 'FUNC':
        case 'FUNCTION':
        case 'SUBROUTINE':
        case 'MACRO':
        case 'SUBPROGRAM':
        case 'PROCEDURE':
        case 'COMMAND':
        case 'SINGLETONMETHOD':
            return mdiFunction
        case 'DEFINE':
        case 'ALIAS':
        case 'VAL':
        case 'FUNCTIONVAR':
        case 'VAR':
        case 'VARIABLE':
            return mdiCube
        case 'CONSTANT':
        case 'CONST':
            return mdiPiBox
        case 'STRING':
        case 'MESSAGE':
        case 'HEREDOC':
            return mdiCodeString
        case 'NUMBER':
            return mdiPound
        case 'BOOLEAN':
            return mdiMatrix
        case 'ARRAY':
            return mdiCodeArray
        case 'OBJECT':
        case 'LITERAL':
        case 'MAP':
            return mdiDrawingBox
        case 'KEY':
        case 'LABEL':
        case 'TARGET':
        case 'SELECTOR':
        case 'ID':
        case 'TAG':
            return mdiKey
        case 'NULL':
            return mdiNull
        case 'STRUCT':
            return mdiPillar
        case 'EVENT':
            return mdiTimetable
        case 'OPERATOR':
            return mdiCodeNotEqual
        case 'TYPEALIAS':
        case 'TYPESPEC':
            return mdiCube
        case 'UNKNOWN':
        default:
            return mdiShape
    }
}

interface SymbolIconProps {
    kind: SymbolKind
    className?: string
}
function getSymbolIconClassName(kind: SymbolKind): string | undefined {
    return (styles as Record<string, string>)[`symbolIconKind${upperFirst(kind.toLowerCase())}`]
}

/**
 * Renders an Icon for a given symbol kind
 */
export const SymbolIcon: React.FunctionComponent<React.PropsWithChildren<SymbolIconProps>> = ({
    kind,
    className = '',
}) => (
    <Tooltip content={kind.toLowerCase()}>
        <Icon
            data-testid="symbol-icon"
            data-symbol-kind={kind.toLowerCase()}
            className={classNames(getSymbolIconClassName(kind), className)}
            svgPath={getSymbolIconComponent(kind)}
            aria-label={`Symbol kind ${kind.toLowerCase()}`}
        />
    </Tooltip>
)
