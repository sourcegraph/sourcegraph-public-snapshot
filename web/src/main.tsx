// This is the entry point for the web app

// Order is important here
// Don't remove the empty lines between these imports

import '../../shared/src/polyfills'

import 'focus-visible'

import './sentry'

import React from 'react'
import { render } from 'react-dom'
import { exploreSections } from './explore/exploreSections'
import { extensionAreaHeaderNavItems } from './extensions/extension/extensionAreaHeaderNavItems'
import { extensionAreaRoutes } from './extensions/extension/routes'
import { extensionsAreaHeaderActionButtons } from './extensions/extensionsAreaHeaderActionButtons'
import { extensionsAreaRoutes } from './extensions/routes'
import { keybindings } from './keybindings'
import { repoHeaderActionButtons } from './repo/repoHeaderActionButtons'
import { repoContainerRoutes, repoRevContainerRoutes } from './repo/routes'
import { routes } from './routes'
import { siteAdminOverviewComponents } from './site-admin/overviewComponents'
import { siteAdminAreaRoutes } from './site-admin/routes'
import { siteAdminSidebarGroups } from './site-admin/sidebaritems'
import { SourcegraphWebApp } from './SourcegraphWebApp'
import { userAreaHeaderNavItems } from './user/area/navitems'
import { userAreaRoutes } from './user/area/routes'
import { userSettingsAreaRoutes } from './user/settings/routes'
import { userSettingsSideBarItems } from './user/settings/sidebaritems'

window.addEventListener('DOMContentLoaded', () => {
    render(
        <SourcegraphWebApp
            exploreSections={exploreSections}
            extensionAreaRoutes={extensionAreaRoutes}
            extensionAreaHeaderNavItems={extensionAreaHeaderNavItems}
            extensionsAreaRoutes={extensionsAreaRoutes}
            extensionsAreaHeaderActionButtons={extensionsAreaHeaderActionButtons}
            siteAdminAreaRoutes={siteAdminAreaRoutes}
            siteAdminSideBarGroups={siteAdminSidebarGroups}
            siteAdminOverviewComponents={siteAdminOverviewComponents}
            userAreaRoutes={userAreaRoutes}
            userAreaHeaderNavItems={userAreaHeaderNavItems}
            userSettingsSideBarItems={userSettingsSideBarItems}
            userSettingsAreaRoutes={userSettingsAreaRoutes}
            repoContainerRoutes={repoContainerRoutes}
            repoRevContainerRoutes={repoRevContainerRoutes}
            repoHeaderActionButtons={repoHeaderActionButtons}
            routes={routes}
            keybindings={keybindings}
        />,
        document.querySelector('#root')
    )
})
