import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'

export const fetchSite = (requestGraphQL: PlatformContext['requestGraphQL']): Observable<GQL.ISite> =>
    requestGraphQL<GQL.IQuery>({
        request: gql`
            query SiteProductVersion {
                site {
                    productVersion
                    buildVersion
                    hasCodeIntelligence
                }
            }
        `,
        variables: {},
        mightContainPrivateInfo: false,
    }).pipe(
        map(dataOrThrowErrors),
        map(({ site }) => site)
    )
