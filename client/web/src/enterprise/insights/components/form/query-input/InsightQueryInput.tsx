import classNames from 'classnames'
import React, { forwardRef } from 'react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import type { MonacoFieldProps } from '../monaco-field'
import * as Monaco from '../monaco-field'

import styles from './InsightQueryInput.module.scss'

export interface InsightQueryInputProps extends MonacoFieldProps {
    patternType: SearchPatternType
}

export const InsightQueryInput = forwardRef<HTMLInputElement, InsightQueryInputProps>((props, reference) => {
    const { children, patternType } = props

    return (
        <div className={styles.root}>
            {children ? (
                <Monaco.Root className={classNames(props.className, styles.inputWrapper)}>
                    <Monaco.Field {...props} ref={reference} className={props.className} />

                    {children}
                </Monaco.Root>
            ) : (
                <Monaco.Field {...props} ref={reference} className={props.className} />
            )}

            <Monaco.PreviewLink query={props.value} patternType={patternType} className={styles.previewButton} />
        </div>
    )
})
