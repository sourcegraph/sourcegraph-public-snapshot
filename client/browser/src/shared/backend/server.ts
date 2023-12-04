import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import type { PlatformContext } from '@sourcegraph/shared/src/platform/context'

import type { SiteProductVersionResult } from '../../graphql-operations'

export const fetchSite = (
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<SiteProductVersionResult['site']> =>
    requestGraphQL<SiteProductVersionResult>({
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
