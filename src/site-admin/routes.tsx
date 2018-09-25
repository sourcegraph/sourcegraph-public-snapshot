import { siteAdminAreaRoutes } from '@sourcegraph/webapp/dist/site-admin/routes'
import { SiteAdminAreaRoute } from '@sourcegraph/webapp/dist/site-admin/SiteAdminArea'
import React from 'react'
import { SiteAdminGenerateLicensePage } from './dotcom/SiteAdminGenerateLicensePage'
import { SiteAdminLicensePage } from './license/SiteAdminLicensePage'
import { SiteAdminAuthenticationProvidersPage } from './SiteAdminAuthenticationProvidersPage'
import { SiteAdminExternalAccountsPage } from './SiteAdminExternalAccountsPage'
import { SiteAdminRegistryExtensionsPage } from './SiteAdminRegistryExtensionsPage'

export const enterpriseSiteAdminAreaRoutes: ReadonlyArray<SiteAdminAreaRoute> = [
    ...siteAdminAreaRoutes,
    {
        path: '/license',
        render: props => <SiteAdminLicensePage {...props} />,
        exact: true,
    },
    {
        path: '/dotcom/generate-license',
        render: props => <SiteAdminGenerateLicensePage {...props} />,
        exact: true,
    },
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
