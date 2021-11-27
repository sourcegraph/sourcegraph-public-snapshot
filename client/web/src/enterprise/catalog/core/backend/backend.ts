import { Observable } from 'rxjs'

import { ListComponentsResult } from '../../../../graphql-operations'

/**
 * The main interface for the catalog. Each backend version should implement this interface in order
 * to support all functionality for the catalog.
 */
export interface CatalogBackend {
    listComponents: () => Observable<ListComponentsResult['catalog']['components']>
}
