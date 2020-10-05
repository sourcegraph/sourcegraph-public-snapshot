import React from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { View } from '../../../../../shared/src/api/client/services/viewService'
import { Markdown } from '../../../../../shared/src/components/Markdown'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { parseRepoURI } from '../../../../../shared/src/util/url'
import { requestGraphQL } from '../../../backend/graphql'
import { TreeReadmeResult, TreeReadmeVariables } from '../../../graphql-operations'
import H from 'history'
import { DirectoryViewContext, URI } from 'sourcegraph'
import { DeepReplace } from '../../../../../shared/src/util/types'

export const treeReadme = (
    context: DeepReplace<DirectoryViewContext, URI, string>,
    history: H.History
): Observable<View | null> => {
    const u = parseRepoURI(context.viewer.directory.uri)
    // TODO(sqs): add rev to TreeReadme query
    //
    // TODO(sqs): support readmes other than README.md (eg README, README.txt, ReadMe, etc.)
    const readme = requestGraphQL<TreeReadmeResult, TreeReadmeVariables>(
        gql`
            query TreeReadme($repoName: String!, $path: String!) {
                repository(name: $repoName) {
                    defaultBranch {
                        target {
                            commit {
                                blob(path: $path) {
                                    richHTML
                                }
                            }
                        }
                    }
                }
            }
        `,
        { repoName: u.repoName, path: `${u.filePath ? `${u.filePath}/` : ''}README.md` }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.repository?.defaultBranch?.target.commit?.blob?.richHTML)
    )

    return readme.pipe(
        map(readme =>
            readme
                ? {
                      title: null,
                      content: [
                          {
                              reactComponent: () => (
                                  <Markdown
                                      className="view-content__markdown mb-1 pr-3"
                                      dangerousInnerHTML={readme}
                                      history={history}
                                  />
                              ),
                          },
                      ],
                  }
                : null
        )
    )
}
