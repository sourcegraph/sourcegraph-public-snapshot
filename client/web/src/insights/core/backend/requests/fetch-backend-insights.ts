import { Observable } from 'rxjs';
import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import { InsightFields, InsightsResult } from '../../../../graphql-operations';
import { requestGraphQL } from '../../../../backend/graphql';
import { map } from 'rxjs/operators';

const insightFieldsFragment = gql`
    fragment InsightFields on Insight {
        title
        description
        series {
            label
            points {
                dateTime
                value
            }
        }
    }
`
export function fetchBackendInsights(): Observable<InsightFields[]> {
    return requestGraphQL<InsightsResult>(gql`
        query Insights {
            insights {
                nodes {
                    ...InsightFields
                }
            }
        }
        ${insightFieldsFragment}
    `).pipe(
        map(dataOrThrowErrors),
        map(data => data.insights?.nodes ?? [])
    )
}
