import React, { FC, useCallback, useMemo, useState } from 'react'

import classNames from 'classnames'
import { parse as parseJSONC } from 'jsonc-parser'
import { noop } from 'lodash'
import { RouteComponentProps } from 'react-router'

import { useMutation, useQuery } from '@sourcegraph/http-client'
import { Alert, Button, ButtonLink, H2, Input, Select, ErrorAlert, Form } from '@sourcegraph/wildcard'

import { EXTERNAL_SERVICES } from '../components/externalServices/backend'
import { defaultExternalServices } from '../components/externalServices/externalServices'
import { ConnectionLoading } from '../components/FilteredConnection/ui'
import {
    CreateWebhookResult,
    CreateWebhookVariables,
    ExternalServiceKind,
    ExternalServicesResult,
    ExternalServicesVariables,
    UpdateWebhookResult,
    UpdateWebhookVariables,
    WebhookFields,
} from '../graphql-operations'

import { CREATE_WEBHOOK_QUERY, UPDATE_WEBHOOK_QUERY } from './backend'

import styles from './WebhookCreateUpdatePage.module.scss'

interface WebhookCreateUpdatePageProps extends Pick<RouteComponentProps, 'history'> {
    // existingWebhook is present when this page is used as an update page.
    existingWebhook?: WebhookFields
}

export interface Webhook {
    name: string
    codeHostKind: ExternalServiceKind | null
    codeHostURN: string
    secret: string | null
}

export const WebhookCreateUpdatePage: FC<WebhookCreateUpdatePageProps> = ({ history, existingWebhook }) => {
    const update = existingWebhook !== undefined
    const initialWebhook = update
        ? {
              name: existingWebhook.name,
              codeHostKind: existingWebhook.codeHostKind,
              codeHostURN: existingWebhook.codeHostURN,
              secret: existingWebhook.secret,
          }
        : {
              name: '',
              codeHostKind: null,
              codeHostURN: '',
              secret: null,
          }

    const [webhook, setWebhook] = useState<Webhook>(initialWebhook)
    const [kindsToUrls, setKindsToUrls] = useState<Map<ExternalServiceKind, string[]>>(new Map())

    const { loading, data, error } = useQuery<ExternalServicesResult, ExternalServicesVariables>(EXTERNAL_SERVICES, {
        variables: {
            first: null,
            after: null,
        },
    })
    useMemo(() => {
        if (data?.externalServices && data?.externalServices?.__typename === 'ExternalServiceConnection') {
            const kindToUrlMap = new Map<ExternalServiceKind, string[]>()

            for (const extSvc of data.externalServices.nodes) {
                if (!supportedExternalServiceKind(extSvc.kind)) {
                    continue
                }
                const conf = parseJSONC(extSvc.config)
                if (conf.url) {
                    kindToUrlMap.set(extSvc.kind, (kindToUrlMap.get(extSvc.kind) || []).concat([conf.url]))
                }
            }

            // If there are no external services, then the warning is shown and webhook creation is blocked.
            // At this point we can only have external services with existing URL which existence is enforced
            // by the code host configuration schema.
            if (kindToUrlMap.size !== 0) {
                setKindsToUrls(kindToUrlMap)

                // only fill the initial values for webhook creation
                if (!update) {
                    const [currentKind] = kindToUrlMap.keys()
                    const [currentUrls] = kindToUrlMap.values()
                    // we always generate a secret once and assign it to the webhook. Bitbucket Cloud special case
                    // is handled is an Input and during GraphQL query creation.
                    setWebhook(webhook => ({
                        ...webhook,
                        secret: generateSecret(),
                        codeHostURN: currentUrls[0],
                        codeHostKind: currentKind,
                    }))
                }
            }
        }
    }, [data, update])

    const onCodeHostTypeChange = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(
        event => {
            const selected = event.target.value as ExternalServiceKind
            const selectedUrns = kindsToUrls.get(selected)
            // This cannot happen, because the form is not rendered when there are no created external services
            // which support webhooks (and effectively have URLs in their code host configurations).
            if (!selectedUrns) {
                throw new Error(
                    `${defaultExternalServices[selected].title} code host connection has no URL. Please check related code host configuration.`
                )
            }
            const selectedUrn = selectedUrns[0]
            setWebhook(webhook => ({
                ...webhook,
                codeHostKind: selected,
                codeHostURN: selectedUrn,
            }))
        },
        [kindsToUrls]
    )
    const onCodeHostUrnChange = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(event => {
        setWebhook(webhook => ({ ...webhook, codeHostURN: event.target.value }))
    }, [])
    const onNameChange = useCallback((name: string): void => {
        setWebhook(webhook => ({ ...webhook, name }))
    }, [])
    const onSecretChange = useCallback((secret: string): void => {
        setWebhook(webhook => ({ ...webhook, secret: secret.length === 0 ? null : secret }))
    }, [])
    const [createWebhook, { error: createWebhookError, loading: creationLoading }] = useMutation<
        CreateWebhookResult,
        CreateWebhookVariables
    >(CREATE_WEBHOOK_QUERY, { onCompleted: data => history.push(`/site-admin/webhooks/${data.createWebhook.id}`) })

    const [updateWebhook, { error: updateWebhookError, loading: updateLoading }] = useMutation<
        UpdateWebhookResult,
        UpdateWebhookVariables
    >(UPDATE_WEBHOOK_QUERY, {
        variables: buildUpdateWebhookVariables(webhook, existingWebhook?.id),
        onCompleted: data => history.push(`/site-admin/webhooks/${data.updateWebhook.id}`),
    })

    return (
        <>
            {error && <ErrorAlert error={error} />}
            {loading && <ConnectionLoading />}
            {!loading &&
                !error &&
                (kindsToUrls.size === 0 ? (
                    <Alert variant="warning" className="mt-2">
                        Please add a code host connection in order to create a webhook.
                    </Alert>
                ) : (
                    <div>
                        <H2>Information</H2>
                        <Form
                            onSubmit={event => {
                                event.preventDefault()
                                createWebhook({ variables: convertWebhookToCreateWebhookVariables(webhook) }).catch(
                                    // noop here is used because creation error is handled directly when useMutation is called
                                    noop
                                )
                            }}
                        >
                            <div className={styles.grid}>
                                <Input
                                    className={classNames(styles.first, 'flex-1 mb-0')}
                                    label={<span className="small">Webhook name</span>}
                                    pattern="^[a-zA-Z0-9_'\-\/\.\s]+$"
                                    required={true}
                                    defaultValue={update ? webhook.name : ''}
                                    onChange={event => {
                                        onNameChange(event.target.value)
                                    }}
                                    maxLength={100}
                                />
                                <Select
                                    id="code-host-type-select"
                                    className={classNames(styles.first, 'flex-1 mb-0')}
                                    label={<span className="small">Code host type</span>}
                                    required={true}
                                    defaultValue={webhook.codeHostKind?.toString()}
                                    onChange={onCodeHostTypeChange}
                                    disabled={loading}
                                >
                                    {Array.from(kindsToUrls.keys()).map(kind => (
                                        <option value={kind} key={kind}>
                                            {defaultExternalServices[kind].title}
                                        </option>
                                    ))}
                                </Select>
                                <Select
                                    id="code-host-urn-select"
                                    className={classNames(styles.second, 'flex-1 mb-0')}
                                    label={<span className="small">Code host URN</span>}
                                    required={true}
                                    defaultValue={webhook.codeHostURN}
                                    onChange={onCodeHostUrnChange}
                                    disabled={loading || !webhook.codeHostKind}
                                >
                                    {webhook.codeHostKind &&
                                        kindsToUrls.get(webhook.codeHostKind) &&
                                        kindsToUrls.get(webhook.codeHostKind)?.map(urn => (
                                            <option value={urn} key={urn}>
                                                {urn}
                                            </option>
                                        ))}
                                </Select>
                                <Input
                                    className={classNames(styles.first, 'flex-1 mb-0')}
                                    message={
                                        webhook.codeHostKind &&
                                        webhook.codeHostKind === ExternalServiceKind.BITBUCKETCLOUD ? (
                                            <small>Bitbucket Cloud doesn't support secrets.</small>
                                        ) : (
                                            <small>Randomly generated. Alter as required.</small>
                                        )
                                    }
                                    label={<span className="small">Secret</span>}
                                    disabled={
                                        webhook.codeHostKind !== null &&
                                        webhook.codeHostKind === ExternalServiceKind.BITBUCKETCLOUD
                                    }
                                    pattern="^[a-zA-Z0-9]+$"
                                    onChange={event => {
                                        onSecretChange(event.target.value)
                                    }}
                                    value={
                                        webhook.codeHostKind &&
                                        webhook.codeHostKind === ExternalServiceKind.BITBUCKETCLOUD
                                            ? ''
                                            : webhook.secret || ''
                                    }
                                    maxLength={100}
                                />
                            </div>
                            {update ? (
                                <div className="d-flex flex-shrink-0 mt-2">
                                    <div>
                                        <Button
                                            onClick={event => {
                                                event.preventDefault()
                                                updateWebhook().catch(
                                                    // noop here is used because update error is handled directly when useMutation is called
                                                    noop
                                                )
                                            }}
                                            variant="primary"
                                            disabled={updateLoading || webhook.name.trim() === ''}
                                        >
                                            Update
                                        </Button>
                                    </div>
                                    <div className="ml-1">
                                        <ButtonLink
                                            to={`/site-admin/webhooks/${existingWebhook.id}`}
                                            variant="secondary"
                                        >
                                            Cancel
                                        </ButtonLink>
                                    </div>
                                </div>
                            ) : (
                                <Button
                                    className="mt-2"
                                    type="submit"
                                    variant="primary"
                                    disabled={creationLoading || webhook.name.trim() === ''}
                                >
                                    Create
                                </Button>
                            )}
                            {(createWebhookError || updateWebhookError) && (
                                <ErrorAlert
                                    className="mt-2"
                                    prefix={`Error during ${createWebhookError ? 'creating' : 'updating'} of webhook`}
                                    error={createWebhookError || updateWebhookError}
                                />
                            )}
                        </Form>
                    </div>
                ))}
        </>
    )
}

function supportedExternalServiceKind(kind: ExternalServiceKind): boolean {
    switch (kind) {
        case ExternalServiceKind.BITBUCKETSERVER:
            return true
        case ExternalServiceKind.BITBUCKETCLOUD:
            return true
        case ExternalServiceKind.GITHUB:
            return true
        case ExternalServiceKind.GITLAB:
            return true
        default:
            return false
    }
}

function buildUpdateWebhookVariables(webhook: Webhook, id?: string): UpdateWebhookVariables {
    const secret =
        webhook.codeHostKind !== null && webhook.codeHostKind === ExternalServiceKind.BITBUCKETCLOUD
            ? null
            : webhook.secret

    return {
        // should not happen when update is called
        id: id || '',
        name: webhook.name,
        codeHostKind: webhook.codeHostKind || ExternalServiceKind.OTHER,
        codeHostURN: webhook.codeHostURN,
        secret,
    }
}

function convertWebhookToCreateWebhookVariables(webhook: Webhook): CreateWebhookVariables {
    const secret =
        webhook.codeHostKind !== null && webhook.codeHostKind === ExternalServiceKind.BITBUCKETCLOUD
            ? null
            : webhook.secret
    return {
        name: webhook.name,
        codeHostKind: webhook.codeHostKind || ExternalServiceKind.OTHER,
        codeHostURN: webhook.codeHostURN,
        secret,
    }
}

function generateSecret(): string {
    let text = ''
    const possible = 'ABCDEF0123456789'
    for (let index = 0; index < 12; index++) {
        text += possible.charAt(Math.floor(Math.random() * possible.length))
    }
    return text
}
