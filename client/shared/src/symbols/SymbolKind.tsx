import * as React from 'react'

import { SymbolKind as SymbolKindEnum } from '../graphql-operations'
import { lazyComponent } from '../util/lazyComponent'

import { SymbolIcon } from './SymbolIcon'

const SymbolTag = lazyComponent(() => import('./SymbolTag'), 'SymbolTag')

export const SymbolKind: React.FC<{
    kind: SymbolKindEnum
    className?: string
    asTag?: boolean
}> = ({ asTag, ...props }) => {
    const Component = asTag ? SymbolTag : SymbolIcon

    return <Component {...props} />
}
