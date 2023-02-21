import React, { Suspense } from 'react'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import type { CodeMirrorQueryInputWrapperProps } from './CodeMirrorQueryInputWrapper'

const CodeMirrorQueryInput = lazyComponent(() => import('./CodeMirrorQueryInputWrapper'), 'CodeMirrorQueryInputWrapper')

export const LazyCodeMirrorQueryInput: React.FunctionComponent<
    React.PropsWithChildren<CodeMirrorQueryInputWrapperProps>
> = props => (
    <Suspense fallback={null}>
        <CodeMirrorQueryInput {...props} />
    </Suspense>
)
