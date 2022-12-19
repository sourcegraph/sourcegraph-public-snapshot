import { forwardRef, InputHTMLAttributes, PropsWithChildren } from 'react'

import classNames from 'classnames'
import LinkExternalIcon from 'mdi-react/OpenInNewIcon'

import { QueryChangeSource, QueryState } from '@sourcegraph/search'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import type { MonacoFieldProps } from '../monaco-field'
import * as Monaco from '../monaco-field'

import { generateRepoFiltersQuery } from './utils/generate-repo-filters-query'

import styles from './InsightQueryInput.module.scss'

type NativeInputProps = Omit<InputHTMLAttributes<HTMLInputElement>, 'onChange' | 'onBlur'>
type MonacoPublicProps = Omit<MonacoFieldProps, 'queryState' | 'onChange' | 'aria-invalid'>

export interface InsightQueryInputProps extends MonacoPublicProps, NativeInputProps {
    value: string
    patternType: SearchPatternType
    repositories?: string
    onChange: (value: string) => void
}

export const InsightQueryInput = forwardRef<HTMLInputElement, PropsWithChildren<InsightQueryInputProps>>(
    (props, reference) => {
        const {
            value,
            patternType,
            repositories = '',
            'aria-invalid': ariaInvalid,
            onChange,
            children,
            ...otherProps
        } = props
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
                            queryState={{ query: value, changeSource: QueryChangeSource.userInput }}
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
                        queryState={{ query: value, changeSource: QueryChangeSource.userInput }}
                        className={classNames(styles.inputWrapper, props.className)}
                        onChange={handleOnChange}
                    />
                )}

                <Monaco.PreviewLink query={previewQuery} patternType={patternType} className={styles.previewButton}>
                    Preview results <LinkExternalIcon size={18} />
                </Monaco.PreviewLink>
            </div>
        )
    }
)
