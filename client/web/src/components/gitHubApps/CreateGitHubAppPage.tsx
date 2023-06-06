import React, { FC, useState, useCallback, useRef, useEffect } from 'react'

import { noop } from 'lodash'
import { Link } from 'react-router-dom'

import { Alert, Container, Button, Input, Label, Text, PageHeader, ButtonLink } from '@sourcegraph/wildcard'

import { GitHubAppDomain } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { PageTitle } from '../PageTitle'

interface StateResponse {
    state: string
    webhookUUID?: string
    appID?: string
}

interface FormOptions {
    state: string
    name: string
    webhookURL?: string
}

export interface CreateGitHubAppPageProps {
    defaultEvents: string[]
    defaultPermissions: Record<string, string>
    pageTitle?: string
    headerDescription?: React.ReactNode
    appDomain: GitHubAppDomain
    defaultAppName?: string
    /*
     * If omitted, the user will be asked to specify a URL from the form. If provided, it
     * will be left off the form.
     */
    baseURL?: string
}

/**
 * Page for creating and connecting a new GitHub App.
 */
export const CreateGitHubAppPage: FC<CreateGitHubAppPageProps> = ({
    defaultEvents,
    defaultPermissions,
    pageTitle = 'Create GitHub App',
    headerDescription,
    appDomain,
    defaultAppName = 'Sourcegraph',
    baseURL,
}) => {
    const ref = useRef<HTMLFormElement>(null)
    const formInput = useRef<HTMLInputElement>(null)
    const [name, setName] = useState<string>(defaultAppName)
    const [nameError, setNameError] = useState<string>()
    const [url, setUrl] = useState<string>(baseURL || 'https://github.com')
    const [org, setOrg] = useState<string>('')
    const [error, setError] = useState<string>()

    useEffect(() => eventLogger.logPageView('SiteAdminCreateGiHubApp'), [])

    const originURL = window.location.origin
    const getManifest = useCallback(
        (name: string, webhookURL?: string): string =>
            JSON.stringify({
                name: name.trim(),
                url: originURL,
                hook_attributes: webhookURL ? { url: webhookURL } : undefined,
                redirect_url: new URL('/.auth/githubapp/redirect', originURL).href,
                setup_url: new URL('/.auth/githubapp/setup', originURL).href,
                callback_urls: [new URL('/.auth/github/callback', originURL).href],
                setup_on_update: true,
                public: false,
                default_permissions: defaultPermissions,
                default_events: defaultEvents,
            }),
        [originURL, defaultEvents, defaultPermissions]
    )

    const createActionUrl = useCallback(
        (state: string): string => {
            let prefix = 'settings/apps/new'
            const trimmedOrg = org.trim()
            if (trimmedOrg !== '') {
                prefix = `organizations/${encodeURIComponent(trimmedOrg)}/settings/apps/new`
            }

            const trimmedUrl = url.trim()
            const originURL = trimmedUrl.endsWith('/') ? trimmedUrl : `${trimmedUrl}/`
            return new URL(`${prefix}?state=${state}`, originURL).href
        },
        [org, url]
    )

    const submitForm = useCallback(
        ({ state, webhookURL, name }: FormOptions) => {
            if (state && ref.current && formInput.current) {
                const actionUrl = createActionUrl(state)
                ref.current.action = actionUrl
                formInput.current.value = getManifest(name, webhookURL)
                ref.current.submit()
            }
        },
        [createActionUrl, getManifest]
    )

    const createState = useCallback(async () => {
        setError(undefined)
        try {
            const response = await fetch(
                `/.auth/githubapp/new-app-state?appName=${name}&webhookURN=${url}&domain=${appDomain}`
            )
            const state: StateResponse = await response.json()
            let webhookURL: string | undefined
            if (state.webhookUUID?.length) {
                webhookURL = new URL(`/.api/webhooks/${state.webhookUUID}`, originURL).href
            }
            submitForm({ state: state.state, webhookURL, name })
        } catch (error_) {
            if (error_ instanceof Error) {
                setError(error_.message)
            } else {
                // eslint-disable-next-line no-console
                console.error(error_)
                setError('Unknown error occurred.')
            }
        }
    }, [submitForm, name, appDomain, url, originURL])

    const handleNameChange = useCallback((event: React.ChangeEvent<HTMLInputElement>) => {
        setName(event.target.value)
        const trimmedValue = event.target.value.trim()
        if (trimmedValue.length < 3) {
            setNameError('Name must be 3 characters or more.')
        } else if (trimmedValue.length > 34) {
            // A requirement from GitHub's side
            setNameError('Name must be 34 characters or less.')
        } else {
            setNameError(undefined)
        }
    }, [])

    const handleUrlChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => setUrl(event.target.value.trim()),
        []
    )

    const handleOrgChange = useCallback((event: React.ChangeEvent<HTMLInputElement>) => setOrg(event.target.value), [])

    return (
        <>
            <PageTitle title={pageTitle} />
            <PageHeader
                path={[{ text: pageTitle }]}
                description={
                    headerDescription || (
                        <>
                            Register a GitHub App to better manage GitHub code host connections.
                            <Link to="/help/admin/external_service/github#using-a-github-app" className="ml-1">
                                See how GitHub App configuration works.
                            </Link>
                        </>
                    )
                }
            />
            <Container className="mt-3">
                {error && <Alert variant="danger">Error creating github app: {error}</Alert>}
                <Text>
                    Provide the details for a new GitHub App with the form below. Once you click "Create GitHub App",
                    you will be routed to {baseURL || 'GitHub'} to create the App and choose which repositories to grant
                    it access to. Once created on {baseURL || 'GitHub'}, you'll be redirected back here to finish
                    installing it on Sourcegraph.
                </Text>

                <Label className="w-100">
                    <Text alignment="left" className="mb-2">
                        GitHub App Name
                    </Text>
                    <Input
                        type="text"
                        onChange={handleNameChange}
                        value={name}
                        error={nameError}
                        status={nameError ? 'error' : undefined}
                        placeholder="Sourcegraph"
                        message="The display name of your GitHub App. It must be unique across the GitHub instance."
                    />
                </Label>
                {baseURL ? null : (
                    <Label className="w-100 mt-2">
                        <Text alignment="left" className="mb-2">
                            GitHub URL
                        </Text>
                        <Input
                            type="text"
                            onChange={handleUrlChange}
                            value={url}
                            placeholder="https://github.com"
                            message="The base URL of the GitHub instance, e.g., https://github.com, https://github.company.com."
                        />
                    </Label>
                )}
                <Label className="w-100 mt-2">
                    <Text alignment="left" className="mb-2">
                        Organization name <span className="text-muted">(optional)</span>
                    </Text>
                    <Input
                        type="text"
                        onChange={handleOrgChange}
                        value={org}
                        message={
                            <>
                                By default, the GitHub App will be registered on your personal account. To register the
                                App on a GitHub organization instead, specify the organization name. Only{' '}
                                <Link
                                    to="https://docs.github.com/en/organizations/managing-peoples-access-to-your-organization-with-roles/roles-in-an-organization#organization-owners"
                                    target="_blank"
                                    rel="noopener"
                                >
                                    organization owners
                                </Link>{' '}
                                can register GitHub Apps.
                            </>
                        }
                    />
                </Label>
                <div className="mt-3">
                    <Button variant="primary" onClick={createState}>
                        Create Github App
                    </Button>
                    <ButtonLink className="ml-3" to="/site-admin/github-apps" variant="secondary">
                        Cancel
                    </ButtonLink>
                </div>
                {/* eslint-disable-next-line react/forbid-elements */}
                <form ref={ref} method="post">
                    {/* eslint-disable-next-line react/forbid-elements */}
                    <input ref={formInput} name="manifest" onChange={noop} hidden={true} />
                </form>
            </Container>
        </>
    )
}
