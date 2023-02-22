import * as React from 'react'

import { SymbolKind as SymbolKindEnum } from '../graphql-operations'

import { SymbolIcon } from './SymbolIcon'
import { SymbolTag } from './SymbolTag'

export const SymbolKind: React.FC<{
    kind: SymbolKindEnum
    className?: string
    symbolKindTags?: boolean
    // A tabIndex overwrite is used only for the symbol tree, where the Tree
    // component is expected to take over all of the keyboard navigation.
    tabIndex?: number
}> = ({ symbolKindTags, tabIndex, className, kind }) => {
    const Component = symbolKindTags ? SymbolTag : SymbolIcon
    return <Component kind={kind} className={className} tabIndex={tabIndex} />
}
