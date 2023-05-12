import { FC, useState, useCallback, useRef } from 'react'

import { Link, useLocation, useNavigate } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Alert, Container, Button, Input, Label, Text, PageHeader } from '@sourcegraph/wildcard'

import { PageTitle } from '../PageTitle'

export interface AddGitHubPageProps extends TelemetryProps {}

interface stateResponse {
    state: string
    webhookUUID: string
}

/**
 * Page for choosing a service kind and variant to add, among the available options.
 */
export const AddGitHubAppPage: FC<AddGitHubPageProps> = () => {
    const { search } = useLocation()
    const navigate = useNavigate()
    const ref = useRef<HTMLFormElement>(null)
    const formInput = useRef<HTMLInputElement>(null)
    const [name, setName] = useState<string>('')
    const [url, setUrl] = useState<string>('')
    const [org, setOrg] = useState<string>('')
    const [error, setError] = useState<any>(null)

    let alert = null
    const id = new URLSearchParams(search).get('id')
    if (id) {
        // TODO: handle this on a different page
        alert = <Alert variant="info">GitHub App installed successfully.</Alert>
    }

    const baseUrl = window.location.origin
    const getManifest = useCallback(
        (name: string, webhookURL: string): string =>
            JSON.stringify({
                name: name.trim(),
                url: baseUrl,
                hook_attributes: {
                    url: webhookURL,
                },
                redirect_url: new URL('/.auth/githubapp/redirect', baseUrl).href,
                setup_url: new URL('/.auth/githubapp/setup', baseUrl).href,
                callback_urls: [new URL('/.auth/github/callback', baseUrl).href],
                setup_on_update: true,
                public: false,
                default_permissions: {
                    contents: 'read',
                    emails: 'read',
                    members: 'read',
                    metadata: 'read',
                },
                default_events: [
                    'repository',
                    'public',
                    'member',
                    'membership',
                    'organization',
                    'team',
                    'team_add',
                    'meta',
                    'push',
                ],
            }),
        [baseUrl]
    )

    const createActionUrl = useCallback(
        (state: string): string => {
            let prefix = 'settings/apps/new'
            const trimmedOrg = org.trim()
            if (trimmedOrg !== '') {
                prefix = `organizations/${encodeURIComponent(trimmedOrg)}/settings/apps/new`
            }

            const trimmedUrl = url.trim()
            const baseUrl = trimmedUrl.endsWith('/') ? trimmedUrl : `${trimmedUrl}/`
            return new URL(`${prefix}?state=${state}`, baseUrl).href
        },
        [org, url]
    )

    const submitForm = useCallback(
        (state: string, webhookURL: string, name: string) => {
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
        setError(null)
        try {
            const response = await fetch(`/.auth/githubapp/new-app-state?appName=${name}&webhookURN=${url}`)
            const state: stateResponse = await response.json()
            const webhookURL = new URL(`/.api/webhooks/${state.webhookUUID}`, baseUrl).href
            submitForm(state.state, webhookURL, name)
        } catch (_error) {
            setError(_error)
        }
    }, [submitForm, name, url, baseUrl])

    const handleNameChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            setName(event.target.value)
        },
        [setName]
    )

    const handleUrlChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            setUrl(event.target.value)
        },
        [setUrl]
    )

    const handleOrgChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            setOrg(event.target.value)
        },
        [setOrg]
    )

    return (
        <>
            <PageTitle title="Create GitHub App" />
            <PageHeader
                path={[{ text: 'Create GitHub App' }]}
                description={
                    <>
                        Create and connect a GitHub App.
                        {/* TODO: add proper link here */}
                        <Link to="" className="ml-1">
                            See how GitHub App configuration works.
                        </Link>
                    </>
                }
            />
            <Container className="mt-3">
                {alert}
                {error && <Alert variant="danger">Error creating github app: {error}</Alert>}
                <Text>
                    Create a new GitHub App by completing the form below. Once you click "Create GitHub App", you will
                    be redirected to GitHub where you will create your GitHub App, choose which organizations to connect
                    to your GitHub App, and choose which repositories to install.
                </Text>
                <Text>Once completing install in GitHub, you'll be redirected back here.</Text>

                <Label className="w-100">
                    <Text alignment="left" className="mb-2">
                        GitHub App Name
                    </Text>
                    <Input
                        type="text"
                        onChange={handleNameChange}
                        value={name}
                        placeholder="Sourcegraph"
                        message="This is the display name of your GitHub App in GitHub."
                    />
                </Label>
                <Label className="w-100 mt-2">
                    <Text alignment="left" className="mb-2">
                        GitHub URL
                    </Text>
                    <Input
                        type="text"
                        onChange={handleUrlChange}
                        value={url}
                        placeholder="https://github.com"
                        message="This is the URL of the GitHub account or organization that you would like to create your GitHub App in."
                    />
                </Label>
                <Label className="w-100 mt-2">
                    <Text alignment="left" className="mb-2">
                        Organization name <span className="text-muted">(optional)</span>
                    </Text>
                    <Input
                        type="text"
                        onChange={handleOrgChange}
                        value={org}
                        message="If creating a GitHub App for your GitHub Organization, this should match your GitHub Organization name."
                    />
                </Label>
                <div className="mt-3">
                    <Button
                        variant="primary"
                        onClick={createState}
                        disabled={name.trim().length < 3 || url.trim().length < 10}
                    >
                        Create Github App
                    </Button>
                    <Button className="ml-3" onClick={() => navigate(-1)} variant="secondary">
                        Cancel
                    </Button>
                </div>
                {/* eslint-disable-next-line react/forbid-elements */}
                <form ref={ref} method="post">
                    {/* eslint-disable-next-line react/forbid-elements */}
                    <input ref={formInput} name="manifest" onChange={() => {}} hidden={true} />
                </form>
            </Container>
        </>
    )
}
