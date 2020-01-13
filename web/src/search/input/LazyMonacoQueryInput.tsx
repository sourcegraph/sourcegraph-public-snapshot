import React, { Suspense } from 'react'
import { MonacoQueryInputProps } from './MonacoQueryInput'
import { lazyComponent } from '../../util/lazyComponent'

const MonacoQueryInput = lazyComponent(() => import('./MonacoQueryInput'), 'MonacoQueryInput')

const ReadonlyQueryInput: React.FunctionComponent<MonacoQueryInputProps> = ({ queryState }) =>
    <div className="query-input2">
        <input type="text" readOnly={true} className="form-control query-input2__input e2e-query-input" value={queryState.query}/>
    </div>

/**
 * A lazily-loaded {@link MonacoQueryInput}, displaying a read-only query field as a fallback during loading.
 */
export const LazyMonacoQueryInput: React.FunctionComponent<MonacoQueryInputProps> = props => (
    <Suspense fallback={<ReadonlyQueryInput {...props}/>}>
        <MonacoQueryInput {...props}></MonacoQueryInput>
    </Suspense>
)
