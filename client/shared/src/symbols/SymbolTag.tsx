import * as React from 'react'

import classNames from 'classnames'
import { upperFirst } from 'lodash'

import { Tooltip } from '@sourcegraph/wildcard'

import type { SymbolKind } from '../graphql-operations'

import styles from './SymbolTag.module.scss'

const getSymbolTooltip = (kind: SymbolKind): string => {
    switch (kind) {
        case 'TYPEPARAMETER': {
            return 'Type parameter'
        }
        case 'ENUMMEMBER': {
            return 'Enum member'
        }
        default: {
            return upperFirst((kind as string).toLowerCase())
        }
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
        case 'PACKAGE': {
            return styles.tagModule
        }

        case 'CLASS':
        case 'ENUM':
        case 'INTERFACE':
        case 'STRUCT': {
            return styles.tagClass
        }

        case 'METHOD':
        case 'CONSTRUCTOR':
        case 'FUNCTION': {
            return styles.tagFunction
        }

        case 'STRING':
        case 'NUMBER':
        case 'BOOLEAN':
        case 'ARRAY':
        case 'OBJECT':
        case 'NULL': {
            return styles.tagType
        }

        case 'VARIABLE':
        case 'CONSTANT':
        case 'PROPERTY':
        case 'FIELD':
        case 'KEY':
        case 'ENUMMEMBER':
        case 'TYPEPARAMETER': {
            return styles.tagVariable
        }

        case 'EVENT':
        case 'OPERATOR':
        case 'UNKNOWN':
        default: {
            return styles.tagUnknown
        }
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
