import React, { useEffect, useState } from 'react'

import { mdiPlus, mdiGithub } from '@mdi/js'
import classNames from 'classnames'
import { useLocation } from 'react-router'

import { SourcegraphIcon, Card, CardBody, Link, H2, Text, Icon } from '@sourcegraph/wildcard'

import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'

import styles from './InstallGitHubAppSuccessPage.module.scss'

interface GitHubAppInstallation {
    id: number
    account: {
        login: string
        avatar_url: string
    }
}

export const InstallGitHubAppSuccessPage: React.FunctionComponent<React.PropsWithChildren<{}>> = () => {
    const [data, setData] = useState<GitHubAppInstallation | null>()

    const { search } = useLocation()

    const installationID = React.useMemo(() => new URLSearchParams(search).get('installation_id'), [search])

    useEffect(() => {
        if (installationID !== null) {
            fetch(`/.auth/github/get-github-app-installation?installation_id=${encodeURIComponent(installationID)}`, {
                method: 'GET',
            })
                .then(response => response.json())
                .then(response => {
                    const githubAppInstallation = response as GitHubAppInstallation

                    setData(githubAppInstallation)
                })
                .catch(() => setData(null))
        } else {
            setData(null)
        }
    }, [installationID])

    return (
        <Page>
            <PageTitle>Success!</PageTitle>
            <br />
            <Card>
                <CardBody>
                    <div className="d-flex justify-content-center align-items-center">
                        <SourcegraphIcon className={classNames(styles.appLogo)} />
                        <Icon svgPath={mdiPlus} inline={false} aria-hidden={true} />
                        {data ? (
                            <img
                                alt="Organization logo"
                                src={data?.account.avatar_url}
                                className={classNames('media', styles.appLogo)}
                            />
                        ) : (
                            <Icon
                                className={classNames(styles.appLogo)}
                                svgPath={mdiGithub}
                                inline={false}
                                aria-hidden={true}
                            />
                        )}
                    </div>
                    <H2 className="text-center">Sourcegraph.com for GitHub installed on {data?.account.login}</H2>
                    <br />
                    <Text alignment="center" className="mr-3 ml-3">
                        <b>One more thing:</b> to finish setup, let the requestor know that the Sourcegraph Cloud for
                        GitHub App has been installed, and they can complete the connection with GitHub for your
                        organization.
                    </Text>
                </CardBody>
            </Card>
            <Text alignment="center" className="mt-3 text-muted">
                New to Sourcegraph? <Link to="/sign-up">Sign up now</Link> to start searching across your team's code!
            </Text>
        </Page>
    )
}
