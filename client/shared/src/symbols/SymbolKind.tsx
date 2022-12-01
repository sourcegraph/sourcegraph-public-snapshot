import * as React from 'react'

import { lazyComponent } from 'src/util/lazyComponent'
import { SymbolIcon } from 'src/symbols/SymbolIcon'

import { SymbolKind as SymbolKindEnum } from '@sourcegraph/web/out/src/graphql-operations'

const SymbolTag = lazyComponent(() => import('src/symbols/SymbolTag'), 'SymbolTag')

export const SymbolKind: React.FC<{
    kind: SymbolKindEnum
    className?: string
    asTag?: boolean
}> = ({ asTag, ...props }) => {
    const Component = asTag ? SymbolTag : SymbolIcon

    return <Component {...props} />
}
