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
        case 'PACKAGE':
            return mdiPackage
        case 'CLASS':
            return mdiSitemap
        case 'METHOD':
            return mdiCubeOutline
        case 'PROPERTY':
            return mdiWrench
        case 'FIELD':
            return mdiTextBox
        case 'CONSTRUCTOR':
            return mdiCubeOutline
        case 'ENUM':
            return mdiNumeric
        case 'INTERFACE':
            return mdiLink
        case 'FUNCTION':
            return mdiFunction
        case 'VARIABLE':
            return mdiCube
        case 'CONSTANT':
            return mdiPiBox
        case 'STRING':
            return mdiCodeString
        case 'NUMBER':
            return mdiPound
        case 'BOOLEAN':
            return mdiMatrix
        case 'ARRAY':
            return mdiCodeArray
        case 'OBJECT':
            return mdiDrawingBox
        case 'KEY':
            return mdiKey
        case 'NULL':
            return mdiNull
        case 'ENUMMEMBER':
            return mdiNumeric
        case 'STRUCT':
            return mdiPillar
        case 'EVENT':
            return mdiTimetable
        case 'OPERATOR':
            return mdiCodeNotEqual
        case 'TYPEPARAMETER':
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
