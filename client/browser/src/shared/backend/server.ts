import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import * as GQL from '@sourcegraph/shared/src/schema'

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
