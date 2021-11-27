import { ApolloClient } from '@apollo/client'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { fromObservableQuery } from '@sourcegraph/shared/src/graphql/apollo'

import { GetFooResult } from '../../../../../graphql-operations'
import { CatalogBackend } from '../backend'

import { GET_FOO_GQL } from './gql/GetFoo'

export class CatalogGqlBackend implements CatalogBackend {
    constructor(private apolloClient: ApolloClient<object>) {}

    public getFoo = (): Observable<string[]> =>
        fromObservableQuery(
            this.apolloClient.watchQuery<GetFooResult>({
                query: GET_FOO_GQL,
            })
        ).pipe(map(({ data }) => data.catalog.foo))
}
