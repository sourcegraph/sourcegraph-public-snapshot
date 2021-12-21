import React, { useEffect } from 'react'
import { Redirect } from 'react-router'

import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { Container, LoadingSpinner } from '@sourcegraph/wildcard'

import { isNotTreeError, TreePage, useTreePageBreadcrumb } from '../../../../repo/tree/TreePage'
import treePageStyles from '../../../../repo/tree/TreePage.module.scss'
import { PageTitle } from '../../../../components/PageTitle'
import { basename } from '../../../../util/path'
import { TREE_OR_COMPONENT_PAGE } from './gql'

import { TreeOrComponentPageResult, TreeOrComponentPageVariables } from '../../../../graphql-operations'

import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'

import { ErrorAlert } from '../../../../components/alerts'

interface Props extends React.ComponentPropsWithoutRef<typeof TreePage> {}

export const TreeOrComponentPage: React.FunctionComponent<Props> = ({
    repo,
    revision,
    commitID,
    filePath,
    useBreadcrumb,
    telemetryService,
    ...props
}) => {
    useEffect(() => telemetryService.logViewEvent(filePath === '' ? 'Repository' : 'Tree'), [
        filePath,
        telemetryService,
    ])
    useTreePageBreadcrumb({ repo, revision, filePath, telemetryService, useBreadcrumb })

    const { data, error, loading } = useQuery<TreeOrComponentPageResult, TreeOrComponentPageVariables>(
        TREE_OR_COMPONENT_PAGE,
        {
            variables: { repo: repo.id, commitID, inputRevspec: revision, path: filePath },
            fetchPolicy: 'cache-first',
        }
    )

    const pageTitle = `${filePath ? `${basename(filePath)} - ` : ''}${displayRepoName(repo.name)}`

    if (error && isNotTreeError(error)) {
        return <Redirect to={toPrettyBlobURL({ repoName: repo.name, revision, commitID, filePath })} />
    }
    return (
        <Container className={treePageStyles.container}>
            <PageTitle title={pageTitle} />
            {error && <ErrorAlert error={error} />}
            {loading && <LoadingSpinner />}
            {JSON.stringify(data)}
        </Container>
    )
}
