import React, { useState } from 'react'

import classNames from 'classnames'
import GithubIcon from 'mdi-react/GithubIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import { Media } from 'reactstrap'

import { SourcegraphIcon, Card, CardBody, Link } from '@sourcegraph/wildcard'

import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'

import { GitHubAppInstallation } from './ConnectGitHubAppPage'

import styles from './AppLogo.module.scss'

export const InstallGitHubAppSuccessPage: React.FunctionComponent<{}> = () => {
    const [data, setData] = useState<GitHubAppInstallation | null>()

    const queryString = window.location.search
    const urlParameters = new URLSearchParams(queryString)
    const installationID = urlParameters.get('installation_id')

    if (installationID !== null) {
        fetch(`/.auth/github/get-github-app-installation?installation_id=${installationID}`, {
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

    return (
        <Page>
            <PageTitle>Success!</PageTitle>
            <br />
            <Card>
                <CardBody>
                    <div className="d-flex justify-content-center align-items-center">
                        <SourcegraphIcon className={classNames(styles.appLogo)} />
                        <PlusIcon />
                        {data ? (
                            <Media src={data?.account.avatar_url} className={classNames(styles.appLogo)} />
                        ) : (
                            <GithubIcon className={classNames(styles.appLogo)} />
                        )}
                    </div>
                    <h2 className="text-center">Sourcegraph Cloud for GitHub installed on {data?.account.login}</h2>
                    <br />
                    <p className="mr-3 ml-3 text-center">
                        <b>One more thing:</b> to finish setup, let the requestor know that the Sourcegraph Cloud for
                        GitHub App has been installed, and they can complete the connection with GitHub for your
                        organization.
                    </p>
                </CardBody>
            </Card>
            <p className="text-center mt-3 text-muted">
                New to Sourcegraph? <Link to="/sign-up">Sign up now</Link> to start searching across your team's code!
            </p>
        </Page>
    )
}
