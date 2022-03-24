import React, { useEffect, useState, HTMLAttributes } from 'react'

import classNames from 'classnames'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import GithubIcon from 'mdi-react/GithubIcon'

import { Card, CardBody, Link, PageHeader, LoadingSpinner } from '@sourcegraph/wildcard'

import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { UserAvatar } from '../../../user/UserAvatar'

import styles from './GHOrgListItem.module.scss'

const connectOrg = (installation_id: string) => () => {
    const queryString = window.location.search
    const urlParameters = new URLSearchParams(queryString)
    const state = urlParameters.get('state')
    if (state !== null) {
        window.location.assign(
            `/setup/github/app/cloud?setup_action=install&installation_id=${installation_id}&state=${state}`
        )
    } else {
        window.location.assign('/install-github-app-success')
    }
}

export interface GitHubAppInstallation {
    id: number
    account: {
        login: string
        avatar_url: string
    }
}

type GHOrgListItemProps = HTMLAttributes<HTMLLIElement>

export const GHOrgListItem: React.FunctionComponent<GHOrgListItemProps> = ({ children, ...rest }) => (
    <li className={classNames('list-group-item', styles.ghOrgItem)} {...rest}>
        {children}
    </li>
)

export const ConnectGitHubAppPage: React.FunctionComponent<{}> = () => {
    const [data, setData] = useState<GitHubAppInstallation[]>([])
    const [error, setError] = useState<boolean>(false)
    const [loading, setLoading] = useState<boolean>(true)

    const queryString = window.location.search
    const urlParameters = new URLSearchParams(queryString)
    let state = urlParameters.get('state')
    if (state === null) {
        state = ''
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
                                <GHOrgListItem onClick={connectOrg(install.id.toString())} key={install.id}>
                                    <div className="d-flex align-items-start">
                                        <div className="align-self-center">
                                            <UserAvatar
                                                className="icon-inline mb-0 mr-1"
                                                user={{ avatarURL: install.account.avatar_url, displayName: '' }}
                                            />
                                        </div>
                                        <div className="flex-1 align-self-center">
                                            <h3 className="m-0">{install.account.login}</h3>
                                        </div>
                                        <div className="align-self-center ml-3">
                                            <ChevronRightIcon />
                                        </div>
                                    </div>
                                </GHOrgListItem>
                            ))}
                            <GHOrgListItem key={-1}>
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
                            </GHOrgListItem>
                        </ul>
                    ) : (
                        <p>Something went wrong.</p>
                    )}
                </CardBody>
            </Card>
        </Page>
    )
}
