import { createContext, forwardRef, InputHTMLAttributes, useContext, useImperativeHandle, useMemo } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'

import { LazyQueryInput } from '@sourcegraph/branded'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { QueryChangeSource } from '@sourcegraph/shared/src/search'
import { ForwardReferenceComponent } from '@sourcegraph/wildcard'

import { useExperimentalFeatures } from '../../../../../stores'
import { useTheme, ThemePreference } from '../../../../../theme'

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
        'aria-labelledby': ariaLabelledby,
    } = props

    const { renderedWithinFocusContainer } = useContext(MonacoFieldContext)

    // Monaco doesn't have any native input elements, so we mock
    // ref here to avoid React warnings in console about zero usage of
    // element ref with forward ref call.
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    useImperativeHandle(reference, () => null)

    const { enhancedThemePreference } = useTheme()
    const applySuggestionsOnEnter =
        useExperimentalFeatures(features => features.applySearchQuerySuggestionOnEnter) ?? true
    const monacoOptions = useMemo(() => ({ readOnly: disabled }), [disabled])

    return (
        <LazyQueryInput
            ariaLabelledby={ariaLabelledby}
            queryState={{ query: value, changeSource: QueryChangeSource.userInput }}
            isLightTheme={enhancedThemePreference === ThemePreference.Light}
            isSourcegraphDotCom={false}
            preventNewLine={false}
            onChange={({ query }) => onChange(query)}
            patternType={patternType}
            caseSensitive={false}
            globbing={true}
            placeholder={placeholder}
            className={classNames(className, styles.monacoField, 'form-control', 'with-invalid-icon', {
                [styles.focusContainer]: !renderedWithinFocusContainer,
                [styles.monacoFieldWithoutFieldStyles]: renderedWithinFocusContainer,
            })}
            editorOptions={monacoOptions}
            autoFocus={autoFocus}
            onBlur={onBlur}
            applySuggestionsOnEnter={applySuggestionsOnEnter}
        />
    )
})
