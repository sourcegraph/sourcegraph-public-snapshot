import React, { useCallback, useEffect, useState } from 'react'

// eslint-disable-next-line no-restricted-imports
import { logger } from '@sourcegraph/common/src/util/logger'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import { ErrorAlert, Text, H3, LoadingSpinner, PageHeader, Input, Container } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../components/LoaderButton'
import { PageTitle } from '../../../components/PageTitle'
import type { Scalars } from '../../../graphql-operations'

import { SET_USER_CODE_COMPLETIONS_QUOTA, SET_USER_COMPLETIONS_QUOTA, USER_REQUEST_QUOTAS } from './backend'

interface Props {
    user: {
        id: Scalars['ID']
    }
}

export const UserQuotaProfilePage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    user: { id: userID },
}) => {
    const { data, loading, error } = useQuery(USER_REQUEST_QUOTAS, { variables: { userID } })
    const [quota, setQuota] = useState<string>('')
    const [codeCompletionsQuota, setCodeCompletionsQuota] = useState<string>('')

    const [
        setUserCompletionsQuota,
        {
            data: setCompletionsQuotaResponse,
            loading: setUserCompletionsQuotaLoading,
            error: setUserCompletionsQuotaError,
        },
    ] = useMutation(SET_USER_COMPLETIONS_QUOTA)

    const [
        setUserCodeCompletionsQuota,
        {
            data: setCodeCompletionsQuotaResponse,
            loading: setUserCodeCompletionsQuotaLoading,
            error: setUserCodeCompletionsQuotaError,
        },
    ] = useMutation(SET_USER_CODE_COMPLETIONS_QUOTA)

    useEffect(() => {
        if (data?.node?.__typename === 'User' && data.node.completionsQuotaOverride !== null) {
            setQuota(data.node.completionsQuotaOverride)
        } else {
            // No overridden limit.
            setQuota('')
        }
        if (data?.node?.__typename === 'User' && data.node.codeCompletionsQuotaOverride !== null) {
            setCodeCompletionsQuota(data.node.codeCompletionsQuotaOverride)
        } else {
            // No overridden limit.
            setCodeCompletionsQuota('')
        }
    }, [data])

    useEffect(() => {
        if (setCompletionsQuotaResponse) {
            if (setCompletionsQuotaResponse.completionsQuotaOverride !== null) {
                setQuota(setCompletionsQuotaResponse.completionsQuotaOverride)
            } else {
                // No overridden limit.
                setQuota('')
            }
        }
    }, [setCompletionsQuotaResponse])

    useEffect(() => {
        if (setCodeCompletionsQuotaResponse) {
            if (setCodeCompletionsQuotaResponse.codeCompletionsQuotaOverride !== null) {
                setCodeCompletionsQuota(setCodeCompletionsQuotaResponse.codeCompletionsQuotaOverride)
            } else {
                // No overridden limit.
                setCodeCompletionsQuota('')
            }
        }
    }, [setCodeCompletionsQuotaResponse])

    const storeCompletionsQuota = useCallback(() => {
        setUserCompletionsQuota({ variables: { userID, quota: quota === '' ? null : parseInt(quota, 10) } }).catch(
            error => {
                logger.error(error)
            }
        )
    }, [quota, userID, setUserCompletionsQuota])

    const storeCodeCompletionsQuota = useCallback(() => {
        setUserCodeCompletionsQuota({
            variables: { userID, quota: codeCompletionsQuota === '' ? null : parseInt(codeCompletionsQuota, 10) },
        }).catch(error => {
            logger.error(error)
        })
    }, [codeCompletionsQuota, userID, setUserCodeCompletionsQuota])

    if (loading) {
        return <LoadingSpinner />
    }

    if (error) {
        return <ErrorAlert error={error} />
    }

    return (
        <>
            <PageTitle title="User quotas" />
            <PageHeader
                path={[{ text: 'Quotas' }]}
                headingElement="h2"
                description={
                    <>
                        Configure custom quotas for the user. Custom quotas can be used to allow increased load for a
                        specific user, or to reduce the impact a user can have on the system performance.
                    </>
                }
                className="mb-3"
            />
            <Container className="mb-3">
                <H3>Completions</H3>
                <Text>Number of requests per day allowed against the completions APIs.</Text>
                <div className="d-flex justify-content-between align-items-end mb-5">
                    <Input
                        id="completions-quota"
                        name="completions-quota"
                        type="number"
                        value={quota}
                        onChange={event => setQuota(event.currentTarget.value)}
                        spellCheck={false}
                        min={1}
                        disabled={setUserCompletionsQuotaLoading}
                        placeholder={`Global limit: ${
                            data?.site.perUserCompletionsQuota === null
                                ? 'infinite'
                                : data?.site.perUserCompletionsQuota
                        }`}
                        label="Custom completions quota"
                        className="flex-grow-1 mb-0"
                    />
                    <LoaderButton
                        loading={setUserCompletionsQuotaLoading}
                        label="Save"
                        onClick={storeCompletionsQuota}
                        disabled={setUserCompletionsQuotaLoading}
                        variant="primary"
                        className="ml-2"
                    />
                </div>
                {setUserCompletionsQuotaError && <ErrorAlert error={setUserCompletionsQuotaError} className="mb-0" />}
                <Text>Number of requests per day allowed against the code completions APIs.</Text>
                <div className="d-flex justify-content-between align-items-end">
                    <Input
                        id="code-completions-quota"
                        name="code-completions-quota"
                        type="number"
                        value={codeCompletionsQuota}
                        onChange={event => setCodeCompletionsQuota(event.currentTarget.value)}
                        spellCheck={false}
                        min={1}
                        disabled={setUserCodeCompletionsQuotaLoading}
                        placeholder={`Global limit: ${
                            data?.site.perUserCodeCompletionsQuota === null
                                ? 'infinite'
                                : data?.site.perUserCodeCompletionsQuota
                        }`}
                        label="Custom code completions quota"
                        className="flex-grow-1 mb-0"
                    />
                    <LoaderButton
                        loading={setUserCodeCompletionsQuotaLoading}
                        label="Save"
                        onClick={storeCodeCompletionsQuota}
                        disabled={setUserCodeCompletionsQuotaLoading}
                        variant="primary"
                        className="ml-2"
                    />
                </div>
                {setUserCodeCompletionsQuotaError && (
                    <ErrorAlert error={setUserCodeCompletionsQuotaError} className="mb-0" />
                )}
            </Container>
        </>
    )
}
