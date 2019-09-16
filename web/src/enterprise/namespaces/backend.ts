import { gql, dataOrThrowErrors } from '../../../../shared/src/graphql/graphql'
import { queryGraphQL } from '../../backend/graphql'
import { Observable } from 'rxjs'
import { Namespace } from '../../../../shared/src/graphql/schema'
import { map } from 'rxjs/operators'

export const queryNamespaces = (): Observable<Namespace[]> =>
    queryGraphQL(
        gql`
            query ViewerNamespaces {
                # TODO expose combined namespaces field
                users {
                    nodes {
                        __typename
                        id
                        namespaceName
                        url
                    }
                }
                organizations {
                    nodes {
                        __typename
                        id
                        namespaceName
                        url
                    }
                }
            }
        `
    ).pipe(
        map(dataOrThrowErrors),
        map(data => [...data.users.nodes, ...data.organizations.nodes])
    )
