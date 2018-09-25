import { siteAdminOverviewComponents } from '@sourcegraph/webapp/dist/site-admin/overviewComponents'
import { SourcegraphLicense } from './license/SourcegraphLicense'

export const enterpriseSiteAdminOverviewComponents: ReadonlyArray<React.ComponentType> = [
    ...siteAdminOverviewComponents,
    SourcegraphLicense,
]
