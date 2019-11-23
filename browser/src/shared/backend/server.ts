import { Observable } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContext } from '../../../../shared/src/platform/context'

export const fetchCurrentUser = (
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<GQL.IUser | undefined> =>
    requestGraphQL<GQL.IQuery>({
        request: gql`query CurrentUser() {
            currentUser {
                id
                displayName
                username
                avatarURL
                url
                settingsURL
                emails {
                    email
                }
                siteAdmin
            }
        }`,
        variables: {},
        mightContainPrivateInfo: false,
    }).pipe(
        map(dataOrThrowErrors),
        map(
            ({ currentUser }) => currentUser || undefined,
            catchError((err, caught) => caught)
        )
    )

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
        map(
            ({ site }) => site,
            catchError((err, caught) => caught)
        )
    )
