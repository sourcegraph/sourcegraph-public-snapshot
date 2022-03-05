import { Observable, of } from 'rxjs'

import { CodeInsightsGqlBackend } from './code-insights-gql-backend'

export class CodeInsightsGqlBackendLimited extends CodeInsightsGqlBackend {
    public isCodeInsightsLicensed = (): Observable<boolean> => of(false)
}
