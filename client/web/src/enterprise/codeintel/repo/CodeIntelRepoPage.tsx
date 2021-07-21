import * as H from 'history'
import BrainIcon from 'mdi-react/BrainIcon'
import React from 'react'

import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { PageHeader } from '@sourcegraph/wildcard'

import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { RepositoryFields } from '../../../graphql-operations'

interface CodeIntelRepoPageProps extends ThemeProps {
    history: H.History
    location: H.Location
    repo: RepositoryFields
}

export const CodeIntelRepoPage: React.FunctionComponent<CodeIntelRepoPageProps> = ({ repo }) => {
    const repoDisplayName = displayRepoName(repo.name)

    return (
        <Page>
            <PageTitle title="Code Intelligence" />
            <PageHeader
                path={[{ icon: BrainIcon, text: 'Code Intelligence' }]}
                headingElement="h1"
                description="Index your repository for more precise code navigation."
            />
            <p>Add content here for {repoDisplayName}.</p>
        </Page>
    )
}
