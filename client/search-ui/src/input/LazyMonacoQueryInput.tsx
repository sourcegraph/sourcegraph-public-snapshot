import React, { Suspense } from 'react'

import classNames from 'classnames'

import { SettingsExperimentalFeatures } from '@sourcegraph/shared/src/schema/settings.schema'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { MonacoQueryInputProps } from './MonacoQueryInput'

import styles from './LazyMonacoQueryInput.module.scss'

/**
 * Minimal interface for external interaction with the editor.
 */
export interface IEditor {
    focus(): void
}

const MonacoQueryInput = lazyComponent(() => import('./MonacoQueryInput'), 'MonacoQueryInput')
const CodemirrorQueryInput = lazyComponent(() => import('./CodeMirrorQueryInput'), 'CodeMirrorMonacoFacade')

/**
 * A plain query input displayed during lazy-loading of the MonacoQueryInput.
 * It has no suggestions, but still allows to type in and submit queries.
 */
export const PlainQueryInput: React.FunctionComponent<
    React.PropsWithChildren<Pick<MonacoQueryInputProps, 'queryState' | 'autoFocus' | 'onChange' | 'className'>>
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

export interface LazyMonacoQueryInputProps extends MonacoQueryInputProps {
    /**
     * Determines which editor implementation to use.
     */
    editorComponent: SettingsExperimentalFeatures['editor']
}

/**
 * A lazily-loaded {@link MonacoQueryInput}, displaying a read-only query field as a fallback during loading.
 */
export const LazyMonacoQueryInput: React.FunctionComponent<React.PropsWithChildren<LazyMonacoQueryInputProps>> = ({
    editorComponent,
    ...props
}) => {
    const QueryInput = editorComponent === 'codemirror6' ? CodemirrorQueryInput : MonacoQueryInput

    return (
        <Suspense fallback={<PlainQueryInput {...props} />}>
            <QueryInput {...props} />
        </Suspense>
    )
}
