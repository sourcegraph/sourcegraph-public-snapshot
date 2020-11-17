import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import { requestGraphQL } from '../../backend/graphql'
import { CreateCodeMonitorResult, CreateCodeMonitorVariables } from '../../graphql-operations'

export const createCodeMonitor = ({
    namespace,
    description,
    enabled,
    trigger,
    actions,
}: CreateCodeMonitorVariables): Observable<CreateCodeMonitorResult['createCodeMonitor']> => {
    const query = gql`
        mutation CreateCodeMonitor(
            $namespace: ID!
            $description: String!
            $enabled: Boolean!
            $trigger: MonitorTriggerInput!
            $actions: [MonitorActionInput!]!
        ) {
            createCodeMonitor(
                namespace: $namespace
                description: $description
                enabled: $enabled
                trigger: $trigger
                actions: $actions
            ) {
                description
            }
        }
    `

    return requestGraphQL<CreateCodeMonitorResult, CreateCodeMonitorVariables>(query, {
        namespace,
        description,
        enabled,
        trigger,
        actions,
    }).pipe(
        map(dataOrThrowErrors),
        map(data => data.createCodeMonitor)
    )
}
