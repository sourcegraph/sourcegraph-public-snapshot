import * as GQL from '../../../../../../shared/src/graphql/schema'
import { queryGraphQL } from '../../../../backend/graphql'
import { gql, dataOrThrowErrors } from '../../../../../../shared/src/graphql/graphql'
import { map } from 'rxjs/operators'
import { Observable } from 'rxjs'

export function queryActionJobs(
    args: GQL.IActionJobsOnQueryArguments
): Observable<Pick<GQL.IActionJobConnection, 'totalCount'>> {
    return queryGraphQL(
        gql`
            query ActionJobs($state: ActionJobState!, $first: Int) {
                actionJobs(state: $state, first: $first) {
                    totalCount
                }
            }
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.actionJobs)
    )
}
