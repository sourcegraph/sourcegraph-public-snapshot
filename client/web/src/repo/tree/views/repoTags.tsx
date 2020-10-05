import React from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { View } from '../../../../../shared/src/api/client/services/viewService'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { parseRepoURI } from '../../../../../shared/src/util/url'
import { requestGraphQL } from '../../../backend/graphql'
import { RepoTagsResult, RepoTagsVariables } from '../../../graphql-operations'
import { DirectoryViewContext, URI } from 'sourcegraph'
import { DeepReplace } from '../../../../../shared/src/util/types'
import { gitReferenceFragments, GitReferenceNode } from '../../GitReference'
import { pluralize } from '../../../../../shared/src/util/strings'

export const repoTags = (context: DeepReplace<DirectoryViewContext, URI, string>): Observable<View | null> => {
    const { repoName } = parseRepoURI(context.viewer.directory.uri)
    // TODO(sqs): add rev to RepoTags query
    //
    // TODO(sqs): support commits older than 1y ago
    const tags = requestGraphQL<RepoTagsResult, RepoTagsVariables>(
        gql`
            query RepoTags($repoName: String!, $first: Int!, $withBehindAhead: Boolean!) {
                repository(name: $repoName) {
                    tags: gitRefs(orderBy: AUTHORED_OR_COMMITTED_AT, type: GIT_TAG, first: $first) {
                        nodes {
                            ...GitRefFields
                        }
                        totalCount
                    }
                }
            }
            ${gitReferenceFragments}
        `,
        { repoName, first: 10, withBehindAhead: false }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.repository?.tags)
    )

    return tags.pipe(
        map(tags =>
            tags
                ? {
                      title: `${tags.totalCount} ${pluralize('tag', tags.totalCount)}`,
                      titleLink: `/${repoName}/-/tags`,
                      content: [
                          {
                              reactComponent: () => (
                                  <div>
                                      {tags.nodes.map(tag => (
                                          <GitReferenceNode key={tag.id} node={tag} className="border-0 pt-0 pb-1" />
                                      ))}
                                  </div>
                              ),
                          },
                      ],
                  }
                : null
        )
    )
}
