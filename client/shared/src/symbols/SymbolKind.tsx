import * as React from 'react'

import { SymbolKind as SymbolKindEnum } from '../graphql-operations'

import { SymbolIcon } from './SymbolIcon'
import { SymbolTag } from './SymbolTag'

export const SymbolKind: React.FC<{
    kind: SymbolKindEnum
    className?: string
    symbolKindTags?: boolean
}> = ({ symbolKindTags, ...props }) => {
    const Component = symbolKindTags ? SymbolTag : SymbolIcon

    return <Component {...props} />
}
