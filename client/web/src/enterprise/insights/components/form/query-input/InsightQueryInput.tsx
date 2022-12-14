import { forwardRef } from 'react'

import classNames from 'classnames'
import LinkExternalIcon from 'mdi-react/OpenInNewIcon'

import { QueryState } from '@sourcegraph/search'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import type { MonacoFieldProps } from '../monaco-field'
import * as Monaco from '../monaco-field'

import { generateRepoFiltersQuery } from './utils/generate-repo-filters-query'

import styles from './InsightQueryInput.module.scss'

export interface InsightQueryInputProps extends Omit<MonacoFieldProps, 'queryState' | 'onChange'> {
    value: string
    patternType: SearchPatternType
    repositories?: string
    onChange: (value: string) => void
}

export const InsightQueryInput = forwardRef<HTMLInputElement, InsightQueryInputProps>((props, reference) => {
    const { children, patternType, repositories = '', value, onChange, ...otherProps } = props
    const previewQuery = `${generateRepoFiltersQuery(repositories)} ${props.value}`.trim()

    const handleOnChange = (queryState: QueryState): void => {
        if (queryState.query !== value) {
            onChange(queryState.query)
        }
    }

    return (
        <div className={styles.root}>
            {children ? (
                <Monaco.Root className={classNames(props.className, styles.inputWrapper)}>
                    <Monaco.Field
                        {...otherProps}
                        ref={reference}
                        patternType={patternType}
                        queryState={{ query: value }}
                        className={props.className}
                        onChange={handleOnChange}
                    />

                    {children}
                </Monaco.Root>
            ) : (
                <Monaco.Field
                    {...otherProps}
                    ref={reference}
                    patternType={patternType}
                    queryState={{ query: value }}
                    className={classNames(styles.inputWrapper, props.className)}
                    onChange={handleOnChange}
                />
            )}

            <Monaco.PreviewLink query={previewQuery} patternType={patternType} className={styles.previewButton}>
                Preview results <LinkExternalIcon size={18} />
            </Monaco.PreviewLink>
        </div>
    )
})
