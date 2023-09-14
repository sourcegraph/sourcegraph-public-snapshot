import * as React from 'react'

import type { SymbolKind as SymbolKindEnum } from '../graphql-operations'

import { SymbolIcon } from './SymbolIcon'
import { SymbolTag } from './SymbolTag'

export const SymbolKind: React.FC<{
    kind: SymbolKindEnum
    className?: string
    symbolKindTags?: boolean
}> = ({ symbolKindTags, className, kind }) => {
    const Component = symbolKindTags ? SymbolTag : SymbolIcon
    return <Component kind={kind} className={className} />
}
