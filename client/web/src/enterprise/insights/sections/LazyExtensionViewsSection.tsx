import React, { Suspense } from 'react'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import type { ExtensionViewsSectionProps } from './ExtensionViewsSection'

const ExtensionViewsSection = lazyComponent(() => import('./ExtensionViewsSection'), 'ExtensionViewsSection')

/**
 * A lazily-loaded {@link ExtensionViewsSection}.
 */
export const LazyExtensionViewsSection: React.FunctionComponent<ExtensionViewsSectionProps> = props => (
    <Suspense fallback={null}>
        <ExtensionViewsSection {...props} />
    </Suspense>
)
