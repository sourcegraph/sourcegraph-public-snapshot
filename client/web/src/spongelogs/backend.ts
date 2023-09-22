import { QueryResult } from '@apollo/client'

import { gql, useQuery } from '@sourcegraph/http-client'

import type { SpongeLogResult, SpongeLogVariables } from '../graphql-operations'

export function useSpongeLog(uuid: string): QueryResult<SpongeLogResult, SpongeLogVariables> {
    return useQuery<SpongeLogResult, SpongeLogVariables>(
        gql`
            query SpongeLog($uuid: String!) {
                spongeLog(uuid: $uuid) {
                    id
                    log
                    interpreter
                }
            }
        `,
        {
            variables: { uuid },
        }
    )
}
