import { Observable } from 'rxjs'

/**
 * The main interface for the catalog. Each backend version should implement this interface in order
 * to support all functionality for the catalog.
 */
export interface CatalogBackend {
    getFoo: () => Observable<string[]>
}
