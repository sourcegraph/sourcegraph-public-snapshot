import { Suspense, type PropsWithChildren, type FC } from 'react'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import type { V2SearchInputProps } from './V2SearchInput'

const V2SearchInput = lazyComponent(() => import('./V2SearchInput'), 'V2SearchInput')

export const LazyV2SearchInput: FC<PropsWithChildren<V2SearchInputProps>> = props => (
    <Suspense fallback={null}>
        <V2SearchInput {...props} />
    </Suspense>
)
