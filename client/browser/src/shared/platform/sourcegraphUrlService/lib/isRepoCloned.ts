import { of, from } from 'rxjs'
import { map, catchError } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'

import { background } from '../../../../browser-extension/web-extension-api/runtime'
import { logger } from '../../../code-hosts/shared/util/logger'

const QUERY = gql`
    query ResolveRawRepoName($repoName: String!) {
        repository(name: $repoName) {
            mirrorInfo {
                cloned
            }
        }
    }
`
export const isRepoCloned = (sourcegraphURL: string, repoName: string): Promise<boolean> =>
    from(
        background.requestGraphQL<GQL.IQuery>({
            request: QUERY,
            variables: { repoName },
            sourcegraphURL,
        })
    )
        .pipe(
            map(dataOrThrowErrors),
            map(({ repository }) => !!repository?.mirrorInfo?.cloned),
            catchError(error => {
                logger.error(error)
                return of(false)
            })
        )
        .toPromise()
