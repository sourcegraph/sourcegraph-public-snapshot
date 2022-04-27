import { forwardRef } from 'react'

import classNames from 'classnames'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import type { MonacoFieldProps } from '../monaco-field'
import * as Monaco from '../monaco-field'

import { generateRepoFiltersQuery } from './utils/generate-repo-filters-query'

import styles from './InsightQueryInput.module.scss'

export interface InsightQueryInputProps extends MonacoFieldProps {
    patternType: SearchPatternType
    repositories?: string
}

export const InsightQueryInput = forwardRef<HTMLInputElement, InsightQueryInputProps>((props, reference) => {
    const { children, patternType, repositories = '', ...otherProps } = props
    const previewQuery = `${generateRepoFiltersQuery(repositories)} ${props.value}`.trim()

    return (
        <div className={styles.root}>
            {children ? (
                <Monaco.Root className={classNames(props.className, styles.inputWrapper)}>
                    <Monaco.Field
                        {...otherProps}
                        patternType={patternType}
                        ref={reference}
                        className={props.className}
                    />

                    {children}
                </Monaco.Root>
            ) : (
                <Monaco.Field {...props} ref={reference} className={props.className} />
            )}

            <Monaco.PreviewLink query={previewQuery} patternType={patternType} className={styles.previewButton} />
        </div>
    )
})
