import { ApolloClient } from '@apollo/client'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { fromObservableQuery } from '@sourcegraph/shared/src/graphql/apollo'

import { ListComponentsResult } from '../../../../../graphql-operations'
import { CatalogBackend } from '../backend'

import { LIST_COMPONENTS_GQL } from './gql/ListComponents'

export class CatalogGqlBackend implements CatalogBackend {
    constructor(private apolloClient: ApolloClient<object>) {}

    public listComponents = (): Observable<ListComponentsResult['catalog']['components']> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<ListComponentsResult>({
                query: LIST_COMPONENTS_GQL,
            })
        ).pipe(map(({ data }) => data.catalog.components))
}
