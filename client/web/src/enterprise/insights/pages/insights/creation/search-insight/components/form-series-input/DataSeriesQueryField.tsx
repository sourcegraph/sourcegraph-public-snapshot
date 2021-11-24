import classNames from 'classnames'
import { noop } from 'lodash'
import * as Monaco from 'monaco-editor'
import React, { InputHTMLAttributes, useMemo } from 'react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { QueryChangeSource } from '../../../../../../../../search/helpers'
import { LazyMonacoQueryInput } from '../../../../../../../../search/input/LazyMonacoQueryInput'
import { DEFAULT_MONACO_OPTIONS } from '../../../../../../../../search/input/MonacoQueryInput'
import { ThemePreference, useThemeState } from '../../../../../../../../theme'

import styles from './DataSeriesQueryField.module.scss'

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

interface DataSeriesQueryFieldProps
    extends Omit<InputHTMLAttributes<HTMLInputElement>, 'value' | 'onChange' | 'onBlur'> {
    value: string
    onBlur: () => void
    onChange: (value: string) => void
}

export const DataSeriesQueryField: React.FunctionComponent<DataSeriesQueryFieldProps> = props => {
    const { value, className, onChange, onBlur = noop, disabled, autoFocus } = props
    const { enhancedThemePreference } = useThemeState()

    const monacoOptions = useMemo(() => ({ ...MONACO_OPTIONS, readOnly: disabled }), [disabled])

    return (
        <LazyMonacoQueryInput
            queryState={{ query: value, changeSource: QueryChangeSource.userInput }}
            isLightTheme={enhancedThemePreference === ThemePreference.Light}
            isSourcegraphDotCom={false}
            preventNewLine={false}
            onChange={({ query }) => onChange(query)}
            patternType={SearchPatternType.regexp}
            caseSensitive={false}
            globbing={true}
            height="auto"
            onSubmit={noop}
            className={classNames(className, styles.field)}
            editorOptions={monacoOptions}
            autoFocus={autoFocus}
            onBlur={onBlur}
        />
    )
}
