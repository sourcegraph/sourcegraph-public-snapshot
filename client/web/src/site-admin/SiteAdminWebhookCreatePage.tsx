import React, { FC, useCallback, useEffect, useMemo, useState } from 'react'

import { mdiCog } from '@mdi/js'
import classNames from 'classnames'
import { noop } from 'lodash'
import { RouteComponentProps } from 'react-router'
import { catchError } from 'rxjs/operators'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { useMutation } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Alert, Button, Container, H2, Input, PageHeader, Select, useObservable } from '@sourcegraph/wildcard'

import { queryExternalServices as _queryExternalServices } from '../components/externalServices/backend'
import { PageTitle } from '../components/PageTitle'
import {
    CreateWebhookResult,
    CreateWebhookVariables,
    ExternalServiceKind,
    ExternalServicesResult,
} from '../graphql-operations'

import { CREATE_WEBHOOK_QUERY } from './backend'

import styles from './SiteAdminWebhookCreatePage.module.scss'

export interface SiteAdminWebhookCreatePageProps extends TelemetryProps, RouteComponentProps<{}> {
    /** For testing only. */
    queryExternalServices?: typeof _queryExternalServices
}

interface Webhook {
    name: string
    codeHostKind: ExternalServiceKind | null
    codeHostURN: string
    secret: string | null
}

export const SiteAdminWebhookCreatePage: FC<SiteAdminWebhookCreatePageProps> = ({
    telemetryService,
    history,
    queryExternalServices = _queryExternalServices,
}) => {
    const [loading, setLoading] = useState(true)

    useEffect(() => {
        telemetryService.logPageView('SiteAdminWebhookCreatePage')
    }, [telemetryService])

    const [kinds, setKinds] = useState<ExternalServiceKind[]>([])
    const [webhook, setWebhook] = useState<Webhook>({
        name: '',
        codeHostKind: null,
        codeHostURN: '',
        secret: null,
    })
    const [kindsToUrls, setKindsToUrls] = useState<Map<ExternalServiceKind, string[]>>(new Map())

    const extSvcKindsOrError: ExternalServicesResult['externalServices'] | undefined | ErrorLike = useObservable(
        useMemo(
            () =>
                queryExternalServices({
                    first: null,
                    after: null,
                }).pipe(catchError(error => [asError(error)])),
            [queryExternalServices]
        )
    )

    useMemo(() => {
        if (extSvcKindsOrError && !isErrorLike(extSvcKindsOrError)) {
            const kindToUrlMap = extSvcKindsOrError.nodes.reduce(
                (svcMap, extSvc) =>
                    svcMap.set(extSvc.kind, (svcMap.get(extSvc.kind) || []).concat([JSON.parse(extSvc.config).url])),
                new Map<ExternalServiceKind, string[]>()
            )
            setKindsToUrls(kindToUrlMap)
            // If there are no external services, then the warning is shown and webhook creation is blocked
            if (kindToUrlMap.size > 0) {
                const extSvcKinds = new Set(extSvcKindsOrError.nodes.map(node => node.kind))
                const kindsArray = Array.from(extSvcKinds)
                setKinds(kindsArray)
                const currentKind = kindsArray[0]
                // we always generate a secret once and assign it to the webhook. Bitbucket Cloud special case
                // is handled is an Input and during GraphQL query creation.
                setWebhook(webhook => ({
                    ...webhook,
                    secret: generateSecret(),
                    codeHostURN: kindToUrlMap.get(currentKind)?.[0] || '',
                    codeHostKind: currentKind,
                }))
            }
            setLoading(false)
        }
    }, [extSvcKindsOrError])

    const onCodeHostTypeChange = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(
        event => {
            const selected = event.target.value as ExternalServiceKind
            setWebhook(webhook => ({
                ...webhook,
                codeHostKind: selected,
                codeHostURN: kindsToUrls.get(selected)?.[0] || '',
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

    return (
        <Container>
            <PageTitle title="Incoming webhook" />
            <PageHeader
                path={[{ icon: mdiCog }, { to: '/site-admin/webhooks', text: 'Incoming webhooks' }, { text: 'Create' }]}
                className="mb-3"
                headingElement="h2"
            />

            {isErrorLike(extSvcKindsOrError) && (
                <Alert variant="danger" className="mt-2">
                    Error during getting external services.
                </Alert>
            )}
            {!isErrorLike(extSvcKindsOrError) && kindsToUrls.size === 0 && (
                <Alert variant="warning" className="mt-2">
                    Please add an external service to proceed with webhooks creation.
                </Alert>
            )}
            {!isErrorLike(extSvcKindsOrError) && kindsToUrls.size > 0 && (
                <div>
                    <H2>Information</H2>
                    <Form
                        onSubmit={event => {
                            event.preventDefault()
                            createWebhook({ variables: convertWebhookToCreateWebhookVariables(webhook) }).catch(noop)
                        }}
                    >
                        <div className={styles.grid}>
                            <Input
                                className={classNames(styles.first, 'flex-1 mb-0')}
                                label={<span className="small">Webhook name</span>}
                                pattern="^[a-zA-Z0-9_'\-\/\.\s]+$"
                                required={true}
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
                                onChange={onCodeHostTypeChange}
                                disabled={loading || kinds.length === 0}
                            >
                                {kinds.length > 0 &&
                                    kinds.map(kind => (
                                        <option value={kind} key={kind}>
                                            {prettyPrintExternalServiceKind(kind)}
                                        </option>
                                    ))}
                                {kinds.length === 0 && <option>Please create external service</option>}
                            </Select>
                            <Select
                                id="code-host-urn-select"
                                className={classNames(styles.second, 'flex-1 mb-0')}
                                label={<span className="small">Code host URN</span>}
                                required={true}
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
                                    webhook.codeHostKind && webhook.codeHostKind === ExternalServiceKind.BITBUCKETCLOUD
                                        ? ''
                                        : webhook.secret || ''
                                }
                                maxLength={100}
                            />
                        </div>
                        <Button
                            className="mt-2"
                            type="submit"
                            variant="primary"
                            disabled={creationLoading || webhook.name.trim() === ''}
                        >
                            Create
                        </Button>
                        {createWebhookError && (
                            <Alert variant="danger" className="mt-2">
                                Failed to create webhook: {createWebhookError.message}
                            </Alert>
                        )}
                    </Form>
                </div>
            )}
        </Container>
    )
}

function prettyPrintExternalServiceKind(kind: ExternalServiceKind): string {
    switch (kind) {
        case ExternalServiceKind.AWSCODECOMMIT:
            return 'AWS CodeCommit'
        case ExternalServiceKind.BITBUCKETCLOUD:
            return 'Bitbucket Cloud'
        case ExternalServiceKind.BITBUCKETSERVER:
            return 'Bitbucket Server'
        case ExternalServiceKind.GERRIT:
            return 'Gerrit'
        case ExternalServiceKind.GITHUB:
            return 'GitHub'
        case ExternalServiceKind.GITLAB:
            return 'GitLab'
        case ExternalServiceKind.GITOLITE:
            return 'Gitolite'
        case ExternalServiceKind.GOMODULES:
            return 'Go Modules'
        case ExternalServiceKind.JVMPACKAGES:
            return 'JVM packages'
        case ExternalServiceKind.NPMPACKAGES:
            return 'NPM packages'
        case ExternalServiceKind.OTHER:
            return 'Other'
        case ExternalServiceKind.PAGURE:
            return 'Pagure'
        case ExternalServiceKind.PERFORCE:
            return 'Perforce'
        case ExternalServiceKind.PHABRICATOR:
            return 'Phabricator'
        case ExternalServiceKind.PYTHONPACKAGES:
            return 'Python packages'
        case ExternalServiceKind.RUSTPACKAGES:
            return 'Rust packages'
        case ExternalServiceKind.RUBYPACKAGES:
            return 'Ruby packages'
        default:
            return kind
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
