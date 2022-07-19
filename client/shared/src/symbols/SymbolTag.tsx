import * as React from 'react'

import classNames from 'classnames'

import { SymbolKind } from '../graphql-operations'

import styles from './SymbolTag.module.scss'

const getSymbolTag = (kind: SymbolKind): string => {
    switch (kind) {
        case 'FILE':
            return 'file'
        case 'MODULE':
            return 'module'
        case 'NAMESPACE':
            return 'namespace'
        case 'PACKAGE':
            return 'package'
        case 'CLASS':
            return 'class'
        case 'METHOD':
            return 'method'
        case 'PROPERTY':
            return 'property'
        case 'FIELD':
            return 'field'
        case 'CONSTRUCTOR':
            return 'constructor'
        case 'ENUM':
            return 'enum'
        case 'INTERFACE':
            return 'interface'
        case 'FUNCTION':
            return 'function'
        case 'VARIABLE':
            return 'var'
        case 'CONSTANT':
            return 'const'
        case 'STRING':
            return 'string'
        case 'NUMBER':
            return 'number'
        case 'BOOLEAN':
            return 'bool'
        case 'ARRAY':
            return 'array'
        case 'OBJECT':
            return 'object'
        case 'KEY':
            return 'key'
        case 'NULL':
            return 'null'
        case 'ENUMMEMBER':
            return 'enum member'
        case 'STRUCT':
            return 'struct'
        case 'EVENT':
            return 'event'
        case 'OPERATOR':
            return 'operator'
        case 'TYPEPARAMETER':
            return 'type param'
        case 'UNKNOWN':
        default:
            return 'unknown'
    }
}

interface SymbolTagProps {
    kind: SymbolKind
    className?: string
}

function getSymbolIconClassName(kind: SymbolKind): string | undefined {
    return (styles as Record<string, string>)[`${kind.toLowerCase()}Tag`]
}

export const SymbolTag: React.FunctionComponent<React.PropsWithChildren<SymbolTagProps>> = ({ kind, className }) => (
    <span
        className={classNames(getSymbolIconClassName(kind), className, styles.tag)}
        aria-label={`Symbol kind ${kind.toLowerCase()}`}
    >
        {getSymbolTag(kind)}
    </span>
)
