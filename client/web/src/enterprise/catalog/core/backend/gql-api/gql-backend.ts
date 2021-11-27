import { ApolloClient } from '@apollo/client'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { fromObservableQuery } from '@sourcegraph/shared/src/graphql/apollo'

import { CatalogComponentsResult, CatalogComponentsVariables } from '../../../../../graphql-operations'
import { CatalogBackend } from '../backend'

import { CATALOG_COMPONENTS_GQL } from './gql/CatalogComponents'

export class CatalogGqlBackend implements CatalogBackend {
    constructor(private apolloClient: ApolloClient<object>) {}

    public listComponents = (
        variables: CatalogComponentsVariables
    ): Observable<CatalogComponentsResult['catalog']['components']> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<CatalogComponentsResult>({
                query: CATALOG_COMPONENTS_GQL,
                variables,
            })
        ).pipe(map(({ data }) => data.catalog.components))
}
