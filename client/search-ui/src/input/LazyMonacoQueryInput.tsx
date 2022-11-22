import React, { Suspense } from 'react'

import classNames from 'classnames'

import { SettingsExperimentalFeatures } from '@sourcegraph/shared/src/schema/settings.schema'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { Input } from '@sourcegraph/wildcard'

import { CodeMirrorQueryInputFacadeProps } from './CodeMirrorQueryInput'
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
    React.PropsWithChildren<
        Pick<MonacoQueryInputProps, 'queryState' | 'autoFocus' | 'onChange' | 'className' | 'placeholder'>
    >
> = ({ queryState, autoFocus, onChange, className, placeholder }) => {
    const onInputChange = React.useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            onChange({ query: event.target.value })
        },
        [onChange]
    )
    return (
        <Input
            autoFocus={autoFocus}
            inputClassName={classNames('text-code', styles.lazyMonacoQueryInputIntermediateInput, className)}
            className="w-100"
            value={queryState.query}
            onChange={onInputChange}
            spellCheck={false}
            placeholder={placeholder}
        />
    )
}

type QueryInputProps = CodeMirrorQueryInputFacadeProps & MonacoQueryInputProps

export interface LazyMonacoQueryInputProps extends QueryInputProps {
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
    const isCodeMirror = editorComponent === 'codemirror6'
    const QueryInput = isCodeMirror ? CodemirrorQueryInput : MonacoQueryInput

    return (
        <Suspense fallback={<PlainQueryInput {...props} placeholder={isCodeMirror ? props.placeholder : undefined} />}>
            <QueryInput {...props} />
        </Suspense>
    )
}
