import { NamespaceAreaRoute } from '../../namespaces/NamespaceArea'
import { namespaceAreaRoutes } from '../../namespaces/routes'
import { graphOwnerAreaRoutes } from '../graphs/graphOwner/routes'

export const enterpriseNamespaceAreaRoutes: readonly NamespaceAreaRoute[] = [
    ...namespaceAreaRoutes,
    ...graphOwnerAreaRoutes,
]
