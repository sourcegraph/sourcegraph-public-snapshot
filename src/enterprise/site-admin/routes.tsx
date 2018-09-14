import React from 'react'
import { siteAdminAreaRoutes } from '../../site-admin/routes'
import { SiteAdminAreaRoute } from '../../site-admin/SiteAdminArea'
import { SiteAdminAuthenticationProvidersPage } from './SiteAdminAuthenticationProvidersPage'
import { SiteAdminExternalAccountsPage } from './SiteAdminExternalAccountsPage'
import { SiteAdminRegistryExtensionsPage } from './SiteAdminRegistryExtensionsPage'

export const enterpriseSiteAdminAreaRoutes: ReadonlyArray<SiteAdminAreaRoute> = [
    ...siteAdminAreaRoutes,
    {
        path: '/auth/providers',
        render: props => <SiteAdminAuthenticationProvidersPage {...props} />,
        exact: true,
    },
    {
        path: '/auth/external-accounts',
        render: props => <SiteAdminExternalAccountsPage {...props} />,
        exact: true,
    },
    {
        path: '/registry/extensions',
        render: props => <SiteAdminRegistryExtensionsPage {...props} />,
        exact: true,
    },
]
