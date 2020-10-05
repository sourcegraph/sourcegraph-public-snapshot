import React from 'react'
import { Observable } from 'rxjs'
import { filter, map } from 'rxjs/operators'
import { View } from '../../../../../shared/src/api/client/services/viewService'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { parseRepoURI } from '../../../../../shared/src/util/url'
import { requestGraphQL } from '../../../backend/graphql'
import { RepoBranchesResult, RepoBranchesVariables } from '../../../graphql-operations'
import { DirectoryViewContext, URI } from 'sourcegraph'
import { DeepReplace, isDefined } from '../../../../../shared/src/util/types'
import { gitReferenceFragments, GitReferenceNode } from '../../GitReference'
import { pluralize } from '../../../../../shared/src/util/strings'

export const repoBranches = (context: DeepReplace<DirectoryViewContext, URI, string>): Observable<View | null> => {
    const { repoName } = parseRepoURI(context.viewer.directory.uri)
    // TODO(sqs): add rev to RepoBranches query
    //
    // TODO(sqs): support commits older than 1y ago
    const branchesData = requestGraphQL<RepoBranchesResult, RepoBranchesVariables>(
        gql`
            query RepoBranches($repoName: String!, $first: Int!, $withBehindAhead: Boolean!) {
                repository(name: $repoName) {
                    defaultBranch {
                        ...GitRefFields
                    }
                    branches: gitRefs(
                        orderBy: AUTHORED_OR_COMMITTED_AT
                        type: GIT_BRANCH
                        first: $first
                        interactive: true
                    ) {
                        nodes {
                            ...GitRefFields
                        }
                        totalCount
                    }
                }
            }
            ${gitReferenceFragments}
        `,
        { repoName, first: 10, withBehindAhead: true }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.repository)
    )

    return branchesData.pipe(
        filter(isDefined),
        map(({ defaultBranch, branches }) => ({
            title: `${branches.totalCount} ${pluralize('branch', branches.totalCount, 'branches')}`,
            titleLink: `/${repoName}/-/branches`,
            content: [
                {
                    reactComponent: () => (
                        <div>
                            {defaultBranch && (
                                <GitReferenceNode
                                    node={defaultBranch}
                                    showBehindAhead={false}
                                    className="border-0 pt-0 pb-1"
                                >
                                    <small className="text-muted">Default branch</small>
                                </GitReferenceNode>
                            )}
                            {branches.nodes
                                .filter(branch => branch.id !== defaultBranch?.id)
                                .map(branch => (
                                    <GitReferenceNode key={branch.id} node={branch} className="border-0 pt-0 pb-1" />
                                ))}
                        </div>
                    ),
                },
            ],
        }))
    )
}
