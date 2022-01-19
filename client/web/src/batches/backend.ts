import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../backend/graphql'
import { RepoChangesetsStatsVariables, RepoChangesetsStatsResult } from '../graphql-operations'

/**
 * NOTE: These fields are only available from an enterprise install, but are used to
 * surface batch changes information on the repo `TreePage`, which is available to both
 * OSS and enterprise
 */
export const queryRepoChangesetsStats = ({
    name,
}: RepoChangesetsStatsVariables): Observable<NonNullable<RepoChangesetsStatsResult['repository']>> =>
    requestGraphQL<RepoChangesetsStatsResult, RepoChangesetsStatsVariables>(
        gql`
            query RepoChangesetsStats($name: String!) {
                repository(name: $name) {
                    changesetsStats {
                        open
                        merged
                    }
                }
            }
        `,
        { name }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.repository) {
                throw new Error(`Repository "${name}" not found`)
            }
            return data.repository
        })
    )
