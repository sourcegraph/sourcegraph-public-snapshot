// This is the entry point for the enterprise web app

// Order is important here
// Don't remove the empty lines between these imports

import '../../src/polyfills'

import '../../src/highlight'

import React from 'react'
import { render } from 'react-dom'
import { keybindings } from '../../src/keybindings'
import { SourcegraphWebApp } from '../../src/SourcegraphWebApp'
import { enterpriseExtensionAreaHeaderNavItems } from './extensions/extension/extensionAreaHeaderNavItems'
import { enterpriseExtensionAreaRoutes } from './extensions/extension/routes'
import { enterpriseExtensionsAreaHeaderActionButtons } from './extensions/extensionsAreaHeaderActionButtons'
import { enterpriseExtensionsAreaRoutes } from './extensions/routes'
import { enterpriseRepoHeaderActionButtons } from './repo/repoHeaderActionButtons'
import { enterpriseRepoRevContainerRoutes } from './repo/routes'
import { enterpriseRoutes } from './routes'
import { enterpriseSiteAdminOverviewComponents } from './site-admin/overviewComponents'
import { enterpriseSiteAdminAreaRoutes } from './site-admin/routes'
import { enterpriseSiteAdminSidebarGroups } from './site-admin/sidebaritems'
import { enterpriseUserAccountAreaRoutes } from './user/account/routes'
import { enterpriseUserAccountSideBarItems } from './user/account/sidebaritems'
import { enterpriseUserAreaHeaderNavItems } from './user/navitems'
import { enterpriseUserAreaRoutes } from './user/routes'

window.addEventListener('DOMContentLoaded', () => {
    render(
        <SourcegraphWebApp
            extensionAreaRoutes={enterpriseExtensionAreaRoutes}
            extensionAreaHeaderNavItems={enterpriseExtensionAreaHeaderNavItems}
            extensionsAreaRoutes={enterpriseExtensionsAreaRoutes}
            extensionsAreaHeaderActionButtons={enterpriseExtensionsAreaHeaderActionButtons}
            siteAdminAreaRoutes={enterpriseSiteAdminAreaRoutes}
            siteAdminSideBarGroups={enterpriseSiteAdminSidebarGroups}
            siteAdminOverviewComponents={enterpriseSiteAdminOverviewComponents}
            userAreaHeaderNavItems={enterpriseUserAreaHeaderNavItems}
            userAreaRoutes={enterpriseUserAreaRoutes}
            userAccountSideBarItems={enterpriseUserAccountSideBarItems}
            userAccountAreaRoutes={enterpriseUserAccountAreaRoutes}
            repoRevContainerRoutes={enterpriseRepoRevContainerRoutes}
            repoHeaderActionButtons={enterpriseRepoHeaderActionButtons}
            routes={enterpriseRoutes}
            keybindings={keybindings}
        />,
        document.querySelector('#root')
    )
})
