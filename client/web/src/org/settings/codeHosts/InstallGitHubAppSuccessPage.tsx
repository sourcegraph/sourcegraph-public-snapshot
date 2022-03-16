import React from 'react'

import GithubIcon from 'mdi-react/GithubIcon'

import { Card, CardBody, PageHeader } from '@sourcegraph/wildcard'

import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'

export const InstallGitHubAppSuccessPage: React.FunctionComponent<{}> = () => (
    <Page>
        <PageTitle>Success!</PageTitle>
        <PageHeader
            path={[
                {
                    icon: GithubIcon,
                    text: 'Success!',
                },
            ]}
            className="mb-3"
        />
        <Card>
            <CardBody>
                <p>The Sourcegraph GitHub App has been successfully installed!</p>
            </CardBody>
        </Card>
    </Page>
)
