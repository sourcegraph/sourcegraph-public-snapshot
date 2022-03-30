import React, { useEffect, useState, HTMLAttributes } from 'react'

import classNames from 'classnames'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import GithubIcon from 'mdi-react/GithubIcon'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Card, CardBody, Link, PageHeader, LoadingSpinner } from '@sourcegraph/wildcard'

import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { useQueryStringParameters } from '../../members/utils'

import styles from './GitHubOrgListItem.module.scss'

export interface GitHubAppInstallation {
    id: number
    account: {
        login: string
        avatar_url: string
    }
}

type GitHubOrgListItemProps = HTMLAttributes<HTMLLIElement>

export const GitHubOrgListItem: React.FunctionComponent<GitHubOrgListItemProps> = ({ children, ...rest }) => (
    <li className={classNames('list-group-item', styles.ghOrgItem)} {...rest}>
        {children}
    </li>
)

export const ConnectGitHubAppPage: React.FunctionComponent<TelemetryProps> = () => {
    const [data, setData] = useState<GitHubAppInstallation[]>([])
    const [error, setError] = useState<boolean>(false)
    const [loading, setLoading] = useState<boolean>(true)

    const state = useQueryStringParameters().get('state') || ''

    const connectOrg = (installation_id: string) => () => {
        if (state !== '') {
            window.location.assign(
                `/setup/github/app/cloud?setup_action=install&installation_id=${installation_id}&state=${state}`
            )
        } else {
            window.location.assign('/install-github-app-success')
        }
    }

    useEffect(() => {
        fetch('/.auth/github/get-user-orgs', {
            method: 'GET',
        })
            .then(response => response.json())
            .then(response => {
                const githubAppInstallations = response as GitHubAppInstallation[]
                if (githubAppInstallations.length === 0) {
                    window.location.assign(
                        `https://github.com/apps/${window.context.githubAppCloudSlug}/installations/new?state=${state}`
                    )
                } else {
                    setData(githubAppInstallations)
                    setLoading(false)
                }
            })
            .catch(() => {
                setError(true)
                setLoading(false)
            })
    }, [state])

    return (
        <Page>
            <PageTitle>Success!</PageTitle>
            <PageHeader
                path={[
                    {
                        icon: GithubIcon,
                        text: 'Complete your GitHub connection',
                    },
                ]}
                className="mb-3"
            />
            <Card>
                <CardBody>
                    {loading ? (
                        <LoadingSpinner />
                    ) : !error ? (
                        <ul className="list-group">
                            {data.map(install => (
                                <GitHubOrgListItem onClick={connectOrg(install.id.toString())} key={install.id}>
                                    <div className="d-flex align-items-start">
                                        <div className="align-self-center">
                                            <img
                                                src={install.account.avatar_url}
                                                className={classNames(styles.ghOrgIcon, 'mr-1')}
                                                alt="Organization logo"
                                            />
                                        </div>
                                        <div className="flex-1 align-self-center">
                                            <h3 className="m-1">{install.account.login}</h3>
                                        </div>
                                        <div className="align-self-center ml-3">
                                            <ChevronRightIcon />
                                        </div>
                                    </div>
                                </GitHubOrgListItem>
                            ))}
                            <GitHubOrgListItem key={-1}>
                                <div className="d-flex align-items-start">
                                    <Link
                                        className="flex-1 align-self-center"
                                        to={`https://github.com/apps/${window.context.githubAppCloudSlug}/installations/new?state=${state}`}
                                    >
                                        Connect with a different organization
                                    </Link>
                                    <div className="align-self-center ml-3">
                                        <ChevronRightIcon />
                                    </div>
                                </div>
                            </GitHubOrgListItem>
                        </ul>
                    ) : (
                        <p>Something went wrong.</p>
                    )}
                </CardBody>
            </Card>
        </Page>
    )
}
