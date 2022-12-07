import * as React from 'react'

import { SymbolKind as SymbolKindEnum } from '../graphql-operations'
import { lazyComponent } from '../util/lazyComponent'

const SymbolTag = lazyComponent(() => import('./SymbolTag'), 'SymbolTag')
const SymbolIcon = lazyComponent(() => import('./SymbolIcon'), 'SymbolIcon')

export const SymbolKind: React.FC<{
    kind: SymbolKindEnum
    className?: string
    symbolKindTags?: boolean
}> = ({ symbolKindTags, ...props }) => {
    const Component = symbolKindTags ? SymbolTag : SymbolIcon

    return <Component {...props} />
}
