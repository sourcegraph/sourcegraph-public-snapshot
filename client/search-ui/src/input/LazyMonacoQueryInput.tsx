import classNames from 'classnames'
import React, { Suspense } from 'react'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import styles from './LazyMonacoQueryInput.module.scss'
import { MonacoQueryInputProps } from './MonacoQueryInput'

const MonacoQueryInput = lazyComponent(() => import('./MonacoQueryInput'), 'MonacoQueryInput')

/**
 * A plain query input displayed during lazy-loading of the MonacoQueryInput.
 * It has no suggestions, but still allows to type in and submit queries.
 */
export const PlainQueryInput: React.FunctionComponent<
    Pick<MonacoQueryInputProps, 'queryState' | 'autoFocus' | 'onChange' | 'className'>
> = ({ queryState, autoFocus, onChange, className }) => {
    const onInputChange = React.useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            onChange({ query: event.target.value })
        },
        [onChange]
    )
    return (
        <input
            type="text"
            autoFocus={autoFocus}
            className={classNames('form-control text-code', styles.lazyMonacoQueryInputIntermediateInput, className)}
            value={queryState.query}
            onChange={onInputChange}
            spellCheck={false}
        />
    )
}

/**
 * Dev feature flag to use a plain (non-Monaco) query input, which can load faster and is therefore
 * sometimes nice when you are reloading the page frequently during local development. Run
 * `localStorage.usePlainQueryInput=true;location.reload()` in the JavaScript console to enable,
 * `delete localStorage.usePlainQueryInput;location.reload()` to disable.
 */
const USE_PLAIN_QUERY_INPUT = true // localStorage.getItem('usePlainQueryInput') !== null

/**
 * A lazily-loaded {@link MonacoQueryInput}, displaying a read-only query field as a fallback during loading.
 */
export const LazyMonacoQueryInput: React.FunctionComponent<MonacoQueryInputProps> = props =>
    USE_PLAIN_QUERY_INPUT ? (
        <PlainQueryInput {...props} />
    ) : (
        <Suspense fallback={<PlainQueryInput {...props} />}>
            <MonacoQueryInput {...props} />
        </Suspense>
    )
