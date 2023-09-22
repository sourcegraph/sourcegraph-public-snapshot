import { forwardRef, type InputHTMLAttributes, type PropsWithChildren } from 'react'

import classNames from 'classnames'
import LinkExternalIcon from 'mdi-react/OpenInNewIcon'

import type { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { Input } from '@sourcegraph/wildcard'

import * as Monaco from '../monaco-field'

import { generateRepoFiltersQuery, getRepoQueryPreview } from './utils/generate-repo-filters-query'

import styles from './InsightQueryInput.module.scss'

type NativeInputProps = Omit<InputHTMLAttributes<HTMLInputElement>, 'onChange' | 'onBlur'>

export interface InsightQueryInputProps extends NativeInputProps {
    value: string
    repoQuery: string | null
    repositories: string[]
    patternType: SearchPatternType
    onChange: (value: string) => void
}

export const InsightQueryInput = forwardRef<HTMLInputElement, PropsWithChildren<InsightQueryInputProps>>(
    ({ value, patternType, repoQuery, repositories = [], onChange, children, className, ...otherProps }, reference) => {
        const repoQueryPreview =
            repoQuery !== null ? getRepoQueryPreview(repoQuery) : generateRepoFiltersQuery(repositories)
        const previewQuery = `${repoQueryPreview} ${value}`.trim()

        return (
            <div className={styles.root}>
                <div className={classNames(children ? className : undefined, styles.inputWrapper)}>
                    <Input
                        {...otherProps}
                        ref={reference}
                        value={value}
                        className={children ? undefined : className}
                        onChange={event => onChange(event.currentTarget.value)}
                    />
                    {children}
                </div>
                <Monaco.PreviewLink query={previewQuery} patternType={patternType} className={styles.previewButton}>
                    Preview results <LinkExternalIcon size={18} />
                </Monaco.PreviewLink>
            </div>
        )
    }
)
InsightQueryInput.displayName = 'InsightQueryInput'
