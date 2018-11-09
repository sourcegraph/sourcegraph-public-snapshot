import { Observable } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { IssueResult } from 'src/protocol/plainTypes'
import { FeatureProviderRegistry } from './registry'

export type ProvideIssueResultsSignature = (query: string) => Observable<IssueResult[] | null>

export class IssueResultsProviderRegistry extends FeatureProviderRegistry<{}, ProvideIssueResultsSignature> {
    public provideIssueResults(query: string): Observable<IssueResult[] | null> {
        return provideIssueResults(this.providers, query)
    }
}

export function provideIssueResults(
    providers: Observable<ProvideIssueResultsSignature[]>,
    query: string
): Observable<IssueResult[] | null> {
    return providers.pipe(
        switchMap(providers => {
            if (providers.length === 0) {
                return [null]
            }
            return providers[0](query)
        })
    )
}
