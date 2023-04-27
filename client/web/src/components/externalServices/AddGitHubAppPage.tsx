import { FC, useState, useCallback, useEffect, useRef } from 'react'

import { useLocation } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Input, Label, Alert, H2, Text, Container } from '@sourcegraph/wildcard'

import { PageTitle } from '../PageTitle'

export interface AddGitHubPageProps extends TelemetryProps {}

/**
 * Page for choosing a service kind and variant to add, among the available options.
 */
export const AddGitHubAppPage: FC<AddGitHubPageProps> = () => {
    const { search } = useLocation()
    const ref = useRef<HTMLFormElement>(null)
    const [name, setName] = useState<string>('Sourcegraph')
    const [url, setUrl] = useState<string>('https://github.com/organizations/example')
    const [state, setState] = useState<string | null>(null)

    let alert = null
    const id = new URLSearchParams(search).get('id')
    if (id) {
        // TODO: do something about this
        alert = <Alert variant="info">GitHub App installed successfully.</Alert>
    }

    const manifest = JSON.stringify({
        name: 'milan-test-app-manifest',
        url: 'https://sourcegraph.test:3443',
        hook_attributes: {
            url: 'https://sourcegraph.test/.auth/github/events',
        },
        redirect_url: 'https://sourcegraph.test:3443/.auth/githubapp/redirect',
        setup_url: 'https://sourcegraph.test:3443/.auth/githubapp/setup',
        callback_urls: ['https://sourcegraph.test:3443/.auth/github/callback'],
        public: true,
        default_permissions: {
            issues: 'write',
            checks: 'write',
        },
        default_events: ['issues', 'issue_comment', 'check_suite', 'check_run'],
    })

    let error: any = null

    const createState = useCallback(async () => {
        error = null
        try {
            const response = await fetch('/.auth/githubapp/state')
            const body = await response.text()
            setState(body)
        } catch (e) {
            error = e
        }
    }, [setState])

    useEffect(() => {
        if (state != null && ref.current != null) {
            ref.current.submit()
        }
    }, [error, state, ref])

    const handleNameChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            setName(event.target.value.trim())
        },
        [setName]
    )

    const handleUrlChange = useCallback(
        (event: React.ChangeEvent<HTMLInputElement>) => {
            setUrl(event.target.value.trim())
        },
        [setUrl]
    )

    const actionUrl = new URL(`settings/apps/new?state=${state}`, url.endsWith('/') ? url : `${url}/`).href

    return (
        <>
            <PageTitle title="Add GitHubApp" />
            <H2>Add GitHub App</H2>
            <Container>
                {alert}
                <div className="form-group">
                    <Label className="w-100">
                        <Text alignment="left" className="mb-2">
                            Name
                        </Text>
                        <Input type="text" onChange={handleNameChange} value={name} />
                    </Label>
                </div>
                <div className="form-group">
                    <Label className="w-100">
                        <Text alignment="left" className="mb-2">
                            GitHub URL
                        </Text>
                        <Input type="text" onChange={handleUrlChange} value={url} />
                    </Label>
                </div>
                <div className="form-group">
                    <Button variant="primary" onClick={createState}>
                        Create Github App
                    </Button>
                </div>
                <form ref={ref} action={actionUrl} method="post">
                    <input type="text" name="manifest" value={manifest} onChange={() => {}} hidden />
                </form>
            </Container>
        </>
    )
}
