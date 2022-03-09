import { UiFeatures } from '../code-insights-backend'

import { CodeInsightsGqlBackend } from './code-insights-gql-backend'

export class CodeInsightsGqlBackendLimited extends CodeInsightsGqlBackend {
    public getUiFeatures = (): UiFeatures => ({
        licensed: false,
        getDashboardsContent: () => ({
            addRemoveInsightsButton: {
                disabled: true,
                tooltip: 'TODO: Need tooltip message',
            },
        }),
    })
}
