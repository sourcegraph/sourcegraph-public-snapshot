import classNames from 'classnames'
import { noop } from 'lodash'
import * as Monaco from 'monaco-editor'
import React, { forwardRef, InputHTMLAttributes, useImperativeHandle, useMemo } from 'react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { QueryChangeSource } from '../../../../../search/helpers'
import { LazyMonacoQueryInput } from '../../../../../search/input/LazyMonacoQueryInput'
import { DEFAULT_MONACO_OPTIONS } from '../../../../../search/input/MonacoQueryInput'
import { ThemePreference } from '../../../../../stores/themeState'
import { useTheme } from '../../../../../theme'

import styles from './MonacoField.module.scss'

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
    patternType?: SearchPatternType
    value: string
    onBlur: () => void
    onChange: (value: string) => void
}

export const MonacoField = forwardRef<HTMLInputElement, MonacoFieldProps>((props, reference) => {
    const {
        value,
        className,
        onChange,
        onBlur = noop,
        disabled,
        autoFocus,
        placeholder,
        patternType = SearchPatternType.regexp,
    } = props

    // Monaco doesn't have any native input elements, so we mock
    // ref here to avoid React warnings in console about zero usage of
    // element ref with forward ref call.
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    useImperativeHandle(reference, () => null)

    const { enhancedThemePreference } = useTheme()
    const monacoOptions = useMemo(() => ({ ...MONACO_OPTIONS, readOnly: disabled }), [disabled])

    return (
        <LazyMonacoQueryInput
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
            onSubmit={noop}
            className={classNames(className, styles.field, { [styles.fieldWithPlaceholder]: !value })}
            editorOptions={monacoOptions}
            autoFocus={autoFocus}
            onBlur={onBlur}
        />
    )
})
