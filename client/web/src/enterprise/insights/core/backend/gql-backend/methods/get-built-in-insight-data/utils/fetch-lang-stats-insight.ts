import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../../../../../../../backend/graphql'
import {
    LangStatsInsightContentResult,
    LangStatsInsightContentVariables,
} from '../../../../../../../../graphql-operations'

export function fetchLangStatsInsight(query: string): Observable<LangStatsInsightContentResult> {
    return requestGraphQL<LangStatsInsightContentResult, LangStatsInsightContentVariables>(
        gql`
            query LangStatsInsightContent($query: String!) {
                search(query: $query) {
                    results {
                        limitHit
                    }
                    stats {
                        languages {
                            name
                            totalLines
                        }
                    }
                }
            }
        `,
        { query }
    ).pipe(map(dataOrThrowErrors))
}
