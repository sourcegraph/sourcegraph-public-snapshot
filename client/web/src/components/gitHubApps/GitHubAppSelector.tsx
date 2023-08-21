import React, { useState, useMemo, useEffect } from 'react'

import { useLocation, useNavigate } from 'react-router-dom'

import { useQuery } from '@sourcegraph/http-client'
import { Alert, Label, Link, LoadingSpinner, Select, Text } from '@sourcegraph/wildcard'

import type { GitHubAppsWithInstallsResult, GitHubAppsWithInstallsVariables } from '../../graphql-operations'
import type { GitHubAppDetails } from '../externalServices/backend'

import { GITHUB_APPS_WITH_INSTALLATIONS_QUERY } from './backend'

interface GitHubApp extends GitHubAppDetails {
    id?: string
    name?: string
}

interface Props {
    disabled?: boolean
    gitHubApp?: GitHubApp
}

const parseQueryParams = (search: string, gitHubApp?: GitHubAppDetails): GitHubApp => {
    if (gitHubApp) {
        return gitHubApp
    }

    const urlParams = new URLSearchParams(search)
    return {
        baseURL: urlParams.get('url') ?? '',
        appID: Number(urlParams.get('appID') ?? -1),
        installationID: Number(urlParams.get('installationID') ?? -1),
    }
}

export const GitHubAppSelector: React.FC<Props> = ({ disabled = false, gitHubApp }) => {
    const navigate = useNavigate()
    const { search } = useLocation()

    const { data, loading, error } = useQuery<GitHubAppsWithInstallsResult, GitHubAppsWithInstallsVariables>(
        GITHUB_APPS_WITH_INSTALLATIONS_QUERY,
        {}
    )
    const apps = useMemo(() => data?.gitHubApps?.nodes ?? [], [data])

    const { appID, baseURL, installationID, id, name } = parseQueryParams(search, gitHubApp)

    const [selectedAppID, setSelectedAppID] = useState<string>(`${baseURL}|${appID}`)
    const [selectedInstallID, setSelectedInstallID] = useState<number>(installationID)

    const selectedApp = useMemo(() => {
        const [baseURL, appID] = selectedAppID.split('|')
        const app = apps?.find(a => a.baseURL === baseURL && a.appID === Number(appID))
        if (app?.installations?.length === 1) {
            setSelectedInstallID(app.installations[0].id)
        }
        return app
    }, [apps, selectedAppID, setSelectedInstallID])

    useEffect(() => {
        if (disabled) {
            return
        }

        const params = new URLSearchParams(search)
        if (selectedApp) {
            params.set('appID', String(selectedApp.appID))
            params.set('url', selectedApp.baseURL)
        } else {
            params.delete('appID')
            params.delete('url')
        }
        if (selectedApp && selectedInstallID >= 0) {
            params.set('installationID', String(selectedInstallID))
            const selectedInstallation = selectedApp.installations.find(item => item.id === selectedInstallID)
            if (selectedInstallation) {
                params.set('org', selectedInstallation.account.login)
            }
        } else {
            params.delete('installationID')
            params.delete('org')
        }
        navigate(`?${params.toString()}`)
    }, [search, selectedApp, selectedInstallID, disabled, navigate])

    const onAppChange = (event: React.ChangeEvent<HTMLSelectElement>): void => {
        setSelectedAppID(event.target.value.trim())
        setSelectedInstallID(-1)
    }

    const onInstallationChange = (event: React.ChangeEvent<HTMLSelectElement>): void => {
        setSelectedInstallID(Number(event.target.value))
    }

    if (loading) {
        return <LoadingSpinner />
    }
    if (error) {
        return <Alert variant="danger">{error.message}</Alert>
    }

    return (
        <div className="d-flex flex-column">
            <Label className="mt-2">
                <Text className="mb-2" id="github-app-label">
                    GitHub App
                </Text>
                {disabled && id ? (
                    <Link to={`/site-admin/github-apps/${encodeURIComponent(id)}`}>{name}</Link>
                ) : (
                    <Select
                        aria-labelledby="github-app-label"
                        value={selectedAppID}
                        onChange={onAppChange}
                        className="mb-0"
                        disabled={disabled}
                        isCustomStyle={true}
                    >
                        <option value="">Choose a GitHub App</option>
                        {apps?.map(app => (
                            <option key={app.id} value={`${app.baseURL}|${app.appID}`}>
                                {app.name}
                            </option>
                        ))}
                    </Select>
                )}
            </Label>
            <Label className="mt-2">
                <Text className="mb-2" id="installation-id-label">
                    Installation
                </Text>
                <Select
                    aria-labelledby="installation-id-label"
                    value={selectedInstallID}
                    onChange={onInstallationChange}
                    className="mb-0"
                    disabled={disabled || selectedApp?.installations.length === 1}
                    isCustomStyle={true}
                >
                    <option value={-1}>Choose a GitHub App</option>
                    {selectedApp?.installations?.map(installation => (
                        <option key={installation.id} value={installation.id}>
                            {installation.account.login}
                        </option>
                    ))}
                </Select>
            </Label>
        </div>
    )
}
