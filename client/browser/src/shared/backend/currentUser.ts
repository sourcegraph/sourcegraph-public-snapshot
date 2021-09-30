import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { once } from 'lodash'
import { Observable } from 'rxjs'
import { catchError, map, tap } from 'rxjs/operators'
import { CurrentUserResult, CurrentUserVariables } from '../../graphql-operations'

const currentUserQuery = gql`
    query CurrentUser {
        currentUser {
            id
        }
    }
`
export const getCurrentUserID = once(
    (requestGraphQL: PlatformContext['requestGraphQL']): Observable<string | null> => {
        return requestGraphQL<CurrentUserResult, CurrentUserVariables>({
            request: currentUserQuery,
            variables: {},
            mightContainPrivateInfo: false,
        }).pipe(
            map(result => result.data?.currentUser?.id ?? null),
            tap(value => console.log({ value })),
            catchError(error => {
                console.error(error)
                return [null]
            })
        )
    }
)
