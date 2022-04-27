import { ApolloClient } from '@apollo/client'
import { map } from 'rxjs/operators'

import { UiFeaturesConfig } from '../code-insights-backend-types'
import { CodeInsightsGqlBackend } from '../gql-backend'

export class CodeInsightsGqlBackendLimited extends CodeInsightsGqlBackend {
    public readonly UIFeatures: UiFeaturesConfig = {
        licensed: false,
        insightsLimit: 2,
    }

    constructor(apolloClient: ApolloClient<object>) {
        super(apolloClient)

        const getInsights = this.getInsights

        this.getInsights = input =>
            getInsights(input).pipe(
                map(insights =>
                    insights.map(insight => ({
                        ...insight,
                        locked: true,
                    }))
                )
            )
    }
}
