import React from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { View } from '../../../../../shared/src/api/client/services/viewService'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { parseRepoURI } from '../../../../../shared/src/util/url'
import { requestGraphQL } from '../../../backend/graphql'
import { TreeDependenciesResult, TreeDependenciesVariables } from '../../../graphql-operations'
import { DirectoryViewContext, URI } from 'sourcegraph'
import { DeepReplace } from '../../../../../shared/src/util/types'
import { pluralize } from '../../../../../shared/src/util/strings'
import { RepoLink } from '../../../../../shared/src/components/RepoLink'

export const treeDependencies = (context: DeepReplace<DirectoryViewContext, URI, string>): Observable<View | null> => {
    const { repoName, filePath } = parseRepoURI(context.viewer.directory.uri)
    // TODO(sqs): add rev to TreeDependencies query
    //
    // TODO(sqs): support commits older than 1y ago
    const dependencies = requestGraphQL<TreeDependenciesResult, TreeDependenciesVariables>(
        gql`
            query TreeDependencies($repoName: String!, $path: String!, $first: Int!) {
                repository(name: $repoName) {
                    defaultBranch {
                        target {
                            commit {
                                tree(path: $path) {
                                    lsif {
                                        dependencies(first: $first) {
                                            nodes {
                                                lsifName
                                                lsifVersion
                                                lsifManager
                                            }
                                            totalCount
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        `,
        { repoName, path: filePath || '', first: 25 }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.repository?.defaultBranch?.target.commit?.tree?.lsif?.dependencies)
    )

    return dependencies.pipe(
        map(dependencies =>
            dependencies
                ? {
                      title: `${dependencies.totalCount} ${pluralize(
                          'dependency',
                          dependencies.totalCount,
                          'dependencies'
                      )}`,
                      content: [
                          {
                              reactComponent: () => (
                                  <div>
                                      <ul className="list-unstyled">
                                          {dependencies.nodes
                                              .filter(dependency => dependency.lsifName !== repoName)
                                              .map(dependency => (
                                                  <li
                                                      key={`${dependency.lsifManager}:${dependency.lsifName}@${dependency.lsifVersion}`}
                                                      className="pb-1"
                                                  >
                                                      <RepoLink
                                                          repoName={dependency.lsifName}
                                                          to={`/${dependency.lsifName}`}
                                                      />
                                                      {dependency.lsifVersion && (
                                                          <span className="text-muted">
                                                              {' '}
                                                              @ {dependency.lsifVersion}
                                                          </span>
                                                      )}
                                                  </li>
                                              ))}
                                      </ul>
                                  </div>
                              ),
                          },
                      ],
                  }
                : null
        )
    )
}
