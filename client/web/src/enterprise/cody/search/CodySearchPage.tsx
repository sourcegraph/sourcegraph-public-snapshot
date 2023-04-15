import React, { useCallback, useEffect, useState } from 'react'

import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Alert, Form, Input, LoadingSpinner, Text, Badge, useSessionStorage } from '@sourcegraph/wildcard'

import { CodyIcon } from '../../../cody/CodyIcon'
import { BrandLogo } from '../../../components/branding/BrandLogo'
import { useFeatureFlag } from '../../../featureFlags/useFeatureFlag'
import { useURLSyncedString } from '../../../hooks/useUrlSyncedString'
import { eventLogger } from '../../../tracking/eventLogger'

import { translateToQuery } from './translateToQuery'

import searchPageStyles from '../../../storm/pages/SearchPage/SearchPageContent.module.scss'
import styles from './CodySearchPage.module.scss'

interface CodeSearchPageProps {
    authenticatedUser: AuthenticatedUser | null
    telemetryService: TelemetryService
}

export const CodySearchPage: React.FunctionComponent<CodeSearchPageProps> = ({ authenticatedUser }) => {
    useEffect(() => {
        eventLogger.logPageView('CodySearch')
    }, [])

    const navigate = useNavigate()

    const [codyEnabled] = useFeatureFlag('cody-experimental', true)

    /** The value entered by the user in the query input */
    // const [input, setInput] = useState('')
    const [input, setInput] = useURLSyncedString('cody-search', '')
    const codySearchStorage = useSessionStorage<string>('cody-search-input', '')
    const setCodySearchInput = codySearchStorage[1]

    const [inputError, setInputError] = useState<string | null>(null)

    const onInputChange = (newInput: string): void => {
        setInput(newInput)
        setInputError(null)
    }

    const [loading, setLoading] = useState(false)

    const onSubmit = useCallback(() => {
        const sanitizedInput = input.trim()

        if (!sanitizedInput) {
            return
        }

        eventLogger.log('web:codySearch:submit', { input: sanitizedInput }, { input: sanitizedInput })
        setLoading(true)
        translateToQuery(sanitizedInput, authenticatedUser).then(
            query => {
                setLoading(false)

                if (query) {
                    eventLogger.log(
                        'web:codySearch:submitSucceeded',
                        { input: sanitizedInput, translatedQuery: query },
                        { input: sanitizedInput, translatedQuery: query }
                    )
                    setCodySearchInput(JSON.stringify({ input: sanitizedInput, translatedQuery: query }))
                    navigate({
                        pathname: '/search',
                        search: buildSearchURLQuery(query, SearchPatternType.regexp, false) + '&ref=cody-search',
                    })
                } else {
                    eventLogger.log(
                        'web:codySearch:submitFailed',
                        { input: sanitizedInput, reason: 'untranslatable' },
                        { input: sanitizedInput, reason: 'untranslatable' }
                    )
                    setInputError('Cody does not understand this query. Try rephrasing it.')
                }
            },
            error => {
                eventLogger.log(
                    'web:codySearch:submitFailed',
                    {
                        input: sanitizedInput,
                        reason: 'unreachable',
                        error: error?.message,
                    },
                    {
                        input: sanitizedInput,
                        reason: 'unreachable',
                        error: error?.message,
                    }
                )
                setLoading(false)
                setInputError(`Unable to reach Cody. Error: ${error?.message}`)
            }
        )
    }, [navigate, input, authenticatedUser, setCodySearchInput])

    const isLightTheme = useIsLightTheme()

    return (
        <div className={classNames('d-flex flex-column align-items-center px-3', searchPageStyles.searchPage)}>
            <BrandLogo className={searchPageStyles.logo} isLightTheme={isLightTheme} variant="logo" />
            <div className="text-muted mt-3 mr-sm-2 pr-2 text-center">Searching millions of public repositories</div>
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
            <Input
                inputClassName={styles.input}
                value={value}
                onInput={onInput}
                disabled={loading}
                autoFocus={true}
                placeholder="Search for code or files in natural language..."
            />
            <div className="align-items-center d-flex mt-4 justify-content-center">
                <Text className="text-muted mb-0 mr-2" size="small">
                    Powered by Cody <CodyIcon />
                </Text>
                <Badge variant="warning">Experimental</Badge>
            </div>
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
