import React, { useCallback, useEffect, useState } from 'react'

import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Alert, Form, Input, LoadingSpinner } from '@sourcegraph/wildcard'

import { BrandLogo } from '../../../components/branding/BrandLogo'
import { useFeatureFlag } from '../../../featureFlags/useFeatureFlag'
import { useURLSyncedString } from '../../../hooks/useUrlSyncedString'
import { eventLogger } from '../../../tracking/eventLogger'

import { translateToQuery } from './translateToQuery'

import searchPageStyles from '../../../storm/pages/SearchPage/SearchPageContent.module.scss'
import styles from './CodySearchPage.module.scss'

export const CodySearchPage: React.FunctionComponent<{ authenticatedUser: AuthenticatedUser | null }> = ({
    authenticatedUser,
}) => {
    useEffect(() => {
        eventLogger.logPageView('CodySearch')
    }, [])

    const navigate = useNavigate()

    const [codyEnabled] = useFeatureFlag('cody-experimental', false)

    /** The value entered by the user in the query input */
    // const [input, setInput] = useState('')
    const [input, setInput] = useURLSyncedString('cody-search', '')

    const [inputError, setInputError] = useState<string | null>(null)

    const onInputChange = (newInput: string): void => {
        setInput(newInput)
        setInputError(null)
    }

    const [loading, setLoading] = useState(false)

    const onSubmit = useCallback(() => {
        eventLogger.log('web:codySearch:submit', { input })
        setLoading(true)
        translateToQuery(input, authenticatedUser).then(
            query => {
                setLoading(false)

                if (query) {
                    eventLogger.log('web:codySearch:submitSucceeded', { input, translatedQuery: query })
                    navigate({
                        pathname: '/search',
                        search: buildSearchURLQuery(query, SearchPatternType.regexp, false),
                    })
                } else {
                    eventLogger.log('web:codySearch:submitFailed', { input, reason: 'untranslatable' })
                    setInputError('Cody does not understand this query. Try rephrasing it.')
                }
            },
            error => {
                eventLogger.log('web:codySearch:submitFailed', { input, reason: 'unreachable', error: error?.message })
                setLoading(false)
                setInputError(`Unable to reach Cody. Error: ${error?.message}`)
            }
        )
    }, [navigate, input, authenticatedUser])

    const isLightTheme = useIsLightTheme()

    return (
        <div className={classNames('d-flex flex-column align-items-center px-3', searchPageStyles.searchPage)}>
            <BrandLogo className={searchPageStyles.logo} isLightTheme={isLightTheme} variant="logo" />
            {codyEnabled ? (
                <SearchInput
                    value={input}
                    onChange={onInputChange}
                    onSubmit={onSubmit}
                    loading={loading}
                    error={inputError}
                    className={classNames('mt-5 w-100', styles.inputContainer)}
                />
            ) : (
                <Alert variant="info" className="mt-5">
                    Cody is not enabled on this Sourcegraph instance.
                </Alert>
            )}
        </div>
    )
}

const SearchInput: React.FunctionComponent<{
    value: string
    loading: boolean
    error: string | null
    onChange: (value: string) => void
    onSubmit: () => void
    className?: string
}> = ({ value, loading, error, onChange, onSubmit: parentOnSubmit, className }) => {
    const onInput = useCallback<React.FormEventHandler<HTMLInputElement>>(
        event => {
            onChange(event.currentTarget.value)
        },
        [onChange]
    )

    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        event => {
            event.preventDefault()
            parentOnSubmit()
        },
        [parentOnSubmit]
    )

    return (
        <Form onSubmit={onSubmit} className={className}>
            <Input inputClassName={styles.input} value={value} onInput={onInput} disabled={loading} autoFocus={true} />
            {error ? (
                <Alert variant="danger" className="mt-2 w-100">
                    {error}
                </Alert>
            ) : loading ? (
                <LoadingSpinner className="mt-2 d-block mx-auto" />
            ) : null}
        </Form>
    )
}
