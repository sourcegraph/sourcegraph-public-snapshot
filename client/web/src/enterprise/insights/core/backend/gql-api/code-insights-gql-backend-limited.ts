import { Observable, forkJoin, of } from 'rxjs'

import { UiFeatures } from '../code-insights-backend'

import { CodeInsightsGqlBackend } from './code-insights-gql-backend'

export class CodeInsightsGqlBackendLimited extends CodeInsightsGqlBackend {
    public getUiFeatures = (): Observable<UiFeatures> =>
        forkJoin({
            licensed: of(false),
        })
}
