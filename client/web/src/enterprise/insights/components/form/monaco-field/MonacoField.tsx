import { createContext, forwardRef, InputHTMLAttributes, useContext, useImperativeHandle, useMemo } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'
import * as Monaco from 'monaco-editor'

import { QueryChangeSource } from '@sourcegraph/search'
import { LazyMonacoQueryInput, DEFAULT_MONACO_OPTIONS } from '@sourcegraph/search-ui'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { ForwardReferenceComponent } from '@sourcegraph/wildcard'

import { useExperimentalFeatures } from '../../../../../stores'
import { ThemePreference } from '../../../../../stores/themeState'
import { useTheme } from '../../../../../theme'

import styles from './MonacoField.module.scss'

interface Context {
    renderedWithinFocusContainer: boolean
}

const MonacoFieldContext = createContext<Context>({ renderedWithinFocusContainer: false })

const MONACO_CONTAINER_MARK = { renderedWithinFocusContainer: true }

export const MonacoFocusContainer = forwardRef((props, reference) => {
    const { as: Component = 'div', className, children, ...otherProps } = props

    return (
        <MonacoFieldContext.Provider value={MONACO_CONTAINER_MARK}>
            <Component
                {...otherProps}
                className={classNames(
                    'form-control',
                    'with-invalid-icon',
                    styles.container,
                    styles.focusContainer,
                    className
                )}
            >
                {children}
            </Component>
        </MonacoFieldContext.Provider>
    )
}) as ForwardReferenceComponent<'div'>

const MONACO_OPTIONS: Monaco.editor.IStandaloneEditorConstructionOptions = {
    ...DEFAULT_MONACO_OPTIONS,
    wordWrap: 'on',
    fixedOverflowWidgets: false,
    lineHeight: 21,
    scrollbar: {
        vertical: 'auto',
        horizontal: 'hidden',
    },
}

export interface MonacoFieldProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'value' | 'onChange' | 'onBlur'> {
    value: string
    patternType?: SearchPatternType
    onBlur?: () => void
    onChange?: (value: string) => void
}

export const MonacoField = forwardRef<HTMLInputElement, MonacoFieldProps>((props, reference) => {
    const {
        value,
        className,
        onChange = noop,
        onBlur = noop,
        disabled,
        autoFocus,
        placeholder,
        patternType = SearchPatternType.regexp,
    } = props

    const { renderedWithinFocusContainer } = useContext(MonacoFieldContext)

    // Monaco doesn't have any native input elements, so we mock
    // ref here to avoid React warnings in console about zero usage of
    // element ref with forward ref call.
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    useImperativeHandle(reference, () => null)

    const { enhancedThemePreference } = useTheme()
    const editorComponent = useExperimentalFeatures(features => features.editor ?? 'monaco')
    const monacoOptions = useMemo(() => ({ ...MONACO_OPTIONS, readOnly: disabled }), [disabled])

    return (
        <LazyMonacoQueryInput
            editorComponent={editorComponent}
            queryState={{ query: value, changeSource: QueryChangeSource.userInput }}
            isLightTheme={enhancedThemePreference === ThemePreference.Light}
            isSourcegraphDotCom={false}
            preventNewLine={false}
            onChange={({ query }) => onChange(query)}
            patternType={patternType}
            caseSensitive={false}
            globbing={true}
            height="auto"
            placeholder={placeholder}
            className={classNames(className, styles.monacoField, 'form-control', 'with-invalid-icon', {
                [styles.focusContainer]: !renderedWithinFocusContainer,
                [styles.monacoFieldWithoutFieldStyles]: renderedWithinFocusContainer,
            })}
            editorOptions={monacoOptions}
            editorClassName={classNames(styles.editor, { [styles.editorWithPlaceholder]: !value })}
            autoFocus={autoFocus}
            onBlur={onBlur}
        />
    )
})
