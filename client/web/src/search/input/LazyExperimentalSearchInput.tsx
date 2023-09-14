import { Suspense, type PropsWithChildren, type FC } from 'react'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import type { ExperimentalSearchInputProps } from './ExperimentalSearchInput'

const ExperimentalSearchInput = lazyComponent(() => import('./ExperimentalSearchInput'), 'ExperimentalSearchInput')

export const LazyExperimentalSearchInput: FC<PropsWithChildren<ExperimentalSearchInputProps>> = props => (
    <Suspense fallback={null}>
        <ExperimentalSearchInput {...props} />
    </Suspense>
)
