import * as React from 'react'

import { SymbolKind as SymbolKindEnum } from '../graphql-operations'
import { lazyComponent } from '../util/lazyComponent'

const SymbolTag = lazyComponent(() => import('./SymbolTag'), 'SymbolTag')
const SymbolIcon = lazyComponent(() => import('./SymbolIcon'), 'SymbolIcon')

export const SymbolKind: React.FC<{
    kind: SymbolKindEnum
    className?: string
    enableSymbolTags?: boolean
}> = ({ enableSymbolTags, ...props }) => {
    const Component = enableSymbolTags ? SymbolTag : SymbolIcon

    return <Component {...props} />
}
