import * as React from 'react'

import classNames from 'classnames'
import { upperFirst } from 'lodash'

import { Tooltip } from '@sourcegraph/wildcard'

import { SymbolKind } from '../graphql-operations'

import styles from './SymbolTag.module.scss'

const getSymbolTooltip = (kind: SymbolKind): string => {
    switch (kind) {
        case 'ACTIVEBINDINGFUNC':
            return 'Active binding func'
        case 'ALTSTEP':
            return 'Altstep definition'
        case 'ANNOTATION':
            return 'Annotation declarations'
        case 'ANONMEMBER':
            return 'Struct anonymous member'
        case 'ARTIFACTID':
            return 'Artifact identifiers'
        case 'CHUNKLABEL':
            return 'Chunck labels'
        case 'ENUMCONSTANT':
            return 'Enum constant'
        case 'HEADING1':
            return 'H1 headings'
        case 'HEADING2':
            return 'H2 headings'
        case 'HEADING3':
            return 'H3 headings'
        case 'IFCLASS':
            return 'Interface class'
        case 'METHODSPEC':
            return 'Interface method specification'
        default:
            return upperFirst((kind as string).toLowerCase())
    }
}

export const getSymbolInitial = (kind: SymbolKind): string => (kind as string)[0].toUpperCase()

interface SymbolTagProps {
    kind: SymbolKind
    className?: string
}

function getSymbolClassName(kind: SymbolKind): string {
    switch (kind) {
        case 'FILE':
        case 'MODULE':
        case 'NAMESPACE':
        case 'PACKAGE':
            return styles.tagModule

        case 'CLASS':
        case 'ENUM':
        case 'INTERFACE':
        case 'STRUCT':
            return styles.tagClass

        case 'METHOD':
        case 'CONSTRUCTOR':
        case 'FUNCTION':
        case 'FUN':
        case 'FUNC':
            return styles.tagFunction

        case 'STRING':
        case 'NUMBER':
        case 'BOOLEAN':
        case 'ARRAY':
        case 'OBJECT':
        case 'NULL':
            return styles.tagType

        case 'VARIABLE':
        case 'CONSTANT':
        case 'PROPERTY':
        case 'FIELD':
        case 'KEY':
        case 'ENUMCONSTANT':
            return styles.tagVariable

        case 'EVENT':
        case 'OPERATOR':
        case 'UNKNOWN':
        default:
            return styles.tagUnknown
    }
}

export const SymbolTag: React.FunctionComponent<React.PropsWithChildren<SymbolTagProps>> = ({ kind, className }) => (
    <Tooltip content={getSymbolTooltip(kind)}>
        <span
            className={classNames(getSymbolClassName(kind), className, styles.tag)}
            aria-label={`Symbol kind ${kind.toLowerCase()}`}
        >
            {getSymbolInitial(kind)}
        </span>
    </Tooltip>
)
