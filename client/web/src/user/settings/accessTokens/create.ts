import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { EVENT_LOGGER } from '@sourcegraph/shared/src/telemetry/web/eventLogger'

import { requestGraphQL } from '../../../backend/graphql'
import type { CreateAccessTokenResult, CreateAccessTokenVariables, Scalars } from '../../../graphql-operations'

interface createAccessTokenProps extends TelemetryV2Props {
    user: Scalars['ID']
    scopes: string[]
    note: string
    durationSeconds: number | null
}

export function createAccessToken(
    props: createAccessTokenProps
): Observable<CreateAccessTokenResult['createAccessToken']> {
    const { user, scopes, note, durationSeconds, telemetryRecorder } = props
    return requestGraphQL<CreateAccessTokenResult, CreateAccessTokenVariables>(
        gql`
            mutation CreateAccessToken($user: ID!, $scopes: [String!]!, $note: String!, $durationSeconds: Int) {
                createAccessToken(user: $user, scopes: $scopes, note: $note, durationSeconds: $durationSeconds) {
                    id
                    token
                }
            }
        `,
        { user, scopes, note, durationSeconds }
    ).pipe(
        map(({ data, errors }) => {
            if (!data?.createAccessToken || (errors && errors.length > 0)) {
                EVENT_LOGGER.log('CreateAccessTokenFailed')
                telemetryRecorder.recordEvent('settings.token', 'createFail')
                throw createAggregateError(errors)
            }
            EVENT_LOGGER.log('AccessTokenCreated')
            telemetryRecorder.recordEvent('settings.token', 'create')
            return data.createAccessToken
        })
    )
}
