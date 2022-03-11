import GithubIcon from 'mdi-react/GithubIcon'
import React from 'react'

import { Card, CardBody, PageHeader } from '@sourcegraph/wildcard'

import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'

export const InstallGithubAppSuccessPage: React.FunctionComponent<{}> = () => (
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
