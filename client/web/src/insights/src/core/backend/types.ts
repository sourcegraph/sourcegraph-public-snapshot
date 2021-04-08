import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi';
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract';
import { Observable } from 'rxjs';
import { Remote } from 'comlink';

export enum ViewInsightProviderSourceType {
    Backend = 'Backend',
    Extension = 'Extension',
}

export interface ViewInsightProviderResult extends ViewProviderResult {
    /** The source of view provider to distinguish between data from extension and data from backend */
    source: ViewInsightProviderSourceType
}

export interface ApiService {
    getCombinedViews: (getExtensionsInsights: () => Observable<ViewProviderResult[]>) => Observable<ViewInsightProviderResult[]>
    getInsightCombinedViews: (extensionApi: Promise<Remote<FlatExtensionHostAPI>>) => Observable<ViewInsightProviderResult[]>
}
