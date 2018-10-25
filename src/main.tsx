// This is the entry point for the web app

// Order is important here
// Don't remove the empty lines between these imports

import './polyfills'

import './highlight'

import React from 'react'
import { render } from 'react-dom'
import { extensionAreaHeaderNavItems } from './extensions/extension/extensionAreaHeaderNavItems'
import { extensionAreaRoutes } from './extensions/extension/routes'
import { extensionsAreaHeaderActionButtons } from './extensions/extensionsAreaHeaderActionButtons'
import { extensionsAreaRoutes } from './extensions/routes'
import { keybindings } from './keybindings'
import { repoHeaderActionButtons } from './repo/repoHeaderActionButtons'
import { repoRevContainerRoutes } from './repo/routes'
import { siteAdminOverviewComponents } from './site-admin/overviewComponents'
import { siteAdminAreaRoutes } from './site-admin/routes'
import { siteAdminSidebarGroups } from './site-admin/sidebaritems'
import { SourcegraphWebApp } from './SourcegraphWebApp'
import { userAccountAreaRoutes } from './user/account/routes'
import { userAccountSideBarItems } from './user/account/sidebaritems'
import { userAreaHeaderNavItems } from './user/area/navitems'
import { userAreaRoutes } from './user/area/routes'

window.addEventListener('DOMContentLoaded', () => {
    render(
        <SourcegraphWebApp
            extensionAreaRoutes={extensionAreaRoutes}
            extensionAreaHeaderNavItems={extensionAreaHeaderNavItems}
            extensionsAreaRoutes={extensionsAreaRoutes}
            extensionsAreaHeaderActionButtons={extensionsAreaHeaderActionButtons}
            siteAdminAreaRoutes={siteAdminAreaRoutes}
            siteAdminSideBarGroups={siteAdminSidebarGroups}
            siteAdminOverviewComponents={siteAdminOverviewComponents}
            userAreaRoutes={userAreaRoutes}
            userAreaHeaderNavItems={userAreaHeaderNavItems}
            userAccountSideBarItems={userAccountSideBarItems}
            userAccountAreaRoutes={userAccountAreaRoutes}
            repoRevContainerRoutes={repoRevContainerRoutes}
            repoHeaderActionButtons={repoHeaderActionButtons}
            keybindings={keybindings}
        />,
        document.querySelector('#root')
    )
})
