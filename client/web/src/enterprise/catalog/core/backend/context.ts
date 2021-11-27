import React from 'react'
import { throwError } from 'rxjs'

import { CatalogBackend } from './backend'

const errorMockMethod = (methodName: string) => () => throwError(new Error(`Implement ${methodName} method first`))

/**
 * Default context API class. Provides mock methods only.
 */
export class FakeDefaultCatalogBackend implements CatalogBackend {
    public listComponents = errorMockMethod('listComponents')
}

export const CatalogBackendContext = React.createContext<CatalogBackend>(new FakeDefaultCatalogBackend())
