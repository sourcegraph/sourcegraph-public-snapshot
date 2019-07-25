import { OrgAreaRoute } from '../../org/area/OrgArea'
import { orgAreaRoutes } from '../../org/area/routes'
import { enterpriseNamespaceAreaRoutes } from '../namespaces/routes'

export const enterpriseOrganizationAreaRoutes: ReadonlyArray<OrgAreaRoute> = [
    ...orgAreaRoutes,
    ...enterpriseNamespaceAreaRoutes,
]
