import type { InputHTMLAttributes } from 'react'

import classNames from 'classnames'
import LinkExternalIcon from 'mdi-react/OpenInNewIcon'

import type { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { QueryChangeSource, type QueryState } from '@sourcegraph/shared/src/search'

import type { FieldProps } from '../field'
import { Field, FocusContainer, PreviewLink } from '../field'

import { generateRepoFiltersQuery, getRepoQueryPreview } from './utils/generate-repo-filters-query'

import styles from './InsightQueryInput.module.scss'

type NativeInputProps = Omit<InputHTMLAttributes<HTMLInputElement>, 'onChange' | 'onBlur'>
type FieldPublicProps = Omit<FieldProps, 'queryState' | 'onChange' | 'aria-invalid'>

export interface InsightQueryInputProps extends FieldPublicProps, NativeInputProps {
    value: string
    repoQuery: string | null
    repositories: string[]
    patternType: SearchPatternType
    onChange: (value: string) => void
}

export const InsightQueryInput: React.FunctionComponent<InsightQueryInputProps> = ({
    value,
    patternType,
    repoQuery,
    repositories = [],
    'aria-invalid': ariaInvalid,
    onChange,
    children,
    className,
    ...otherProps
}) => {
    const repoQueryPreview =
        repoQuery !== null ? getRepoQueryPreview(repoQuery) : generateRepoFiltersQuery(repositories)
    const previewQuery = `${repoQueryPreview} ${value}`.trim()

    const handleOnChange = (queryState: QueryState): void => {
        if (queryState.query !== value) {
            onChange(queryState.query)
        }
    }

    return (
        <div className={styles.root}>
            {children ? (
                <FocusContainer className={classNames(className, styles.inputWrapper)}>
                    <Field
                        {...otherProps}
                        patternType={patternType}
                        queryState={{ query: value, changeSource: QueryChangeSource.userInput }}
                        className={className}
                        onChange={handleOnChange}
                    />

                    {children}
                </FocusContainer>
            ) : (
                <Field
                    {...otherProps}
                    patternType={patternType}
                    queryState={{ query: value, changeSource: QueryChangeSource.userInput }}
                    className={classNames(styles.inputWrapper, className)}
                    onChange={handleOnChange}
                />
            )}

            <PreviewLink query={previewQuery} patternType={patternType} className={styles.previewButton}>
                Preview results <LinkExternalIcon size={18} />
            </PreviewLink>
        </div>
    )
}
