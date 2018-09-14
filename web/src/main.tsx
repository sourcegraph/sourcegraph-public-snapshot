// This is the entry point for the web app

// Order is important here
// Don't remove the empty lines between these imports

import './polyfills'

import './highlight'

import React from 'react'
import { render } from 'react-dom'
import { enterpriseExtensionAreaHeaderNavItems } from './enterprise/extensions/extension/extensionAreaHeaderNavItems'
import { enterpriseExtensionAreaRoutes } from './enterprise/extensions/extension/routes'
import { enterpriseExtensionsAreaHeaderActionButtons } from './enterprise/extensions/extensionsAreaHeaderActionButtons'
import { enterpriseExtensionsAreaRoutes } from './enterprise/extensions/routes'
import { enterpriseRepoHeaderActionButtons } from './enterprise/repo/repoHeaderActionButtons'
import { enterpriseRepoRevContainerRoutes } from './enterprise/repo/routes'
import { enterpriseSiteAdminAreaRoutes } from './enterprise/site-admin/routes'
import { enterpriseSiteAdminSidebarNavItems } from './enterprise/site-admin/sidebaritems'
import { enterpriseUserAccountAreaRoutes } from './enterprise/user/account/routes'
import { enterpriseUserAccountSideBarItems } from './enterprise/user/account/sidebaritems'
import { SourcegraphWebApp } from './SourcegraphWebApp'

window.addEventListener('DOMContentLoaded', () => {
    render(
        <SourcegraphWebApp
            extensionAreaRoutes={enterpriseExtensionAreaRoutes}
            extensionAreaHeaderNavItems={enterpriseExtensionAreaHeaderNavItems}
            extensionsAreaRoutes={enterpriseExtensionsAreaRoutes}
            extensionsAreaHeaderActionButtons={enterpriseExtensionsAreaHeaderActionButtons}
            siteAdminAreaRoutes={enterpriseSiteAdminAreaRoutes}
            siteAdminSideBarItems={enterpriseSiteAdminSidebarNavItems}
            userAccountSideBarItems={enterpriseUserAccountSideBarItems}
            userAccountAreaRoutes={enterpriseUserAccountAreaRoutes}
            repoRevContainerRoutes={enterpriseRepoRevContainerRoutes}
            repoHeaderActionButtons={enterpriseRepoHeaderActionButtons}
        />,
        document.querySelector('#root')
    )
})
