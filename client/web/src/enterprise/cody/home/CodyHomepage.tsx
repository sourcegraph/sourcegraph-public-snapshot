import React, { useCallback, useState } from 'react'

import classNames from 'classnames'

import { QueryState } from '@sourcegraph/shared/src/search'
import { Input } from '@sourcegraph/wildcard'

import { BrandLogo } from '../../../components/branding/BrandLogo'

import searchPageStyles from '../../../storm/pages/SearchPage/SearchPageContent.module.scss'
import styles from './CodyHomepage.module.scss'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'

interface Props  {}

export const CodyHomepage: React.FunctionComponent<Props> = () => {
    /** The value entered by the user in the query input */
    const [queryState, setQueryState] = useState<QueryState>({
        query: '',
    })

    const onQueryChange = useCallback((value: string) => {
        setQueryState(prev => ({ ...prev, query: value }))
    }, [])

    const isLightTheme = useIsLightTheme()

    return (
        <div className={classNames('d-flex flex-column align-items-center px-3', searchPageStyles.searchPage)}>
            <BrandLogo className={searchPageStyles.logo} isLightTheme={isLightTheme} variant="logo" />
            <SearchInput
                value={queryState.query}
                onChange={onQueryChange}
                className={classNames('mt-5 w-100', styles.inputContainer)}
            />
        </div>
    )
}

const SearchInput: React.FunctionComponent<{
    value: string
    onChange: (value: string) => void
    className?: string
}> = ({ value, onChange, className }) => {
    const onInput = useCallback<React.FormEventHandler<HTMLInputElement>>(
        event => {
            onChange(event.currentTarget.value)
        },
        [onChange]
    )
    return <Input className={className} inputClassName={styles.input} value={value} onInput={onInput} />
}
