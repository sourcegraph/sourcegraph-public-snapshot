import classNames from 'classnames'
import React, { useEffect } from 'react'
import { Redirect } from 'react-router'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useQuery } from '@sourcegraph/http-client'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'
import { Container, LoadingSpinner } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../../components/PageTitle'
import { TreeOrComponentPageResult, TreeOrComponentPageVariables } from '../../../../graphql-operations'
import { isNotTreeError, TreePage, useTreePageBreadcrumb } from '../../../../repo/tree/TreePage'
import treePageStyles from '../../../../repo/tree/TreePage.module.scss'
import { basename } from '../../../../util/path'

import { TREE_OR_COMPONENT_PAGE } from './gql'
import { TreeOrComponent } from './TreeOrComponent'
import styles from './TreeOrComponentPage.module.scss'

interface Props extends React.ComponentPropsWithoutRef<typeof TreePage> {}

export const TreeOrComponentPage: React.FunctionComponent<Props> = ({
    repo,
    revision,
    commitID,
    filePath,
    ...props
}) => {
    useEffect(() => props.telemetryService.logViewEvent(filePath === '' ? 'Repository' : 'Tree'), [
        filePath,
        props.telemetryService,
    ])
    useTreePageBreadcrumb({
        repo,
        revision,
        filePath,
        telemetryService: props.telemetryService,
        useBreadcrumb: props.useBreadcrumb,
    })

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
        <Container className={classNames(treePageStyles.container, styles.container)}>
            <PageTitle title={pageTitle} />
            {loading && !data ? (
                <LoadingSpinner className="m-3 icon-inline" />
            ) : error && !data ? (
                <ErrorAlert error={error} />
            ) : !data || !data.node || data.node.__typename !== 'Repository' ? (
                <ErrorAlert error="Not a repository" />
            ) : (
                <TreeOrComponent {...props} data={data.node} />
            )}
        </Container>
    )
}
