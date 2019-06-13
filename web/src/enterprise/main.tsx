// This is the entry point for the enterprise web app

// Order is important here
// Don't remove the empty lines between these imports

import '../../../shared/src/polyfills'

import '../sentry'

import React from 'react'
import { render } from 'react-dom'
import '../enterprise.scss'
import { keybindings } from '../keybindings'
import { SourcegraphWebApp } from '../SourcegraphWebApp'
import { enterpriseExploreSections } from './explore/exploreSections'
import { enterpriseExtensionAreaHeaderNavItems } from './extensions/extension/extensionAreaHeaderNavItems'
import { enterpriseExtensionAreaRoutes } from './extensions/extension/routes'
import { enterpriseExtensionsAreaHeaderActionButtons } from './extensions/extensionsAreaHeaderActionButtons'
import { enterpriseExtensionsAreaRoutes } from './extensions/routes'
import { enterpriseRepoHeaderActionButtons } from './repo/repoHeaderActionButtons'
import { enterpriseRepoContainerRoutes, enterpriseRepoRevContainerRoutes } from './repo/routes'
import { enterpriseRoutes } from './routes'
import { enterpriseSiteAdminOverviewComponents } from './site-admin/overviewComponents'
import { enterpriseSiteAdminAreaRoutes } from './site-admin/routes'
import { enterpriseSiteAdminSidebarGroups } from './site-admin/sidebaritems'
import { enterpriseUserAreaHeaderNavItems } from './user/navitems'
import { enterpriseUserAreaRoutes } from './user/routes'
import { enterpriseUserSettingsAreaRoutes } from './user/settings/routes'
import { enterpriseUserSettingsSideBarItems } from './user/settings/sidebaritems'

window.addEventListener('DOMContentLoaded', () => {
    render(
        <SourcegraphWebApp
            exploreSections={enterpriseExploreSections}
            extensionAreaRoutes={enterpriseExtensionAreaRoutes}
            extensionAreaHeaderNavItems={enterpriseExtensionAreaHeaderNavItems}
            extensionsAreaRoutes={enterpriseExtensionsAreaRoutes}
            extensionsAreaHeaderActionButtons={enterpriseExtensionsAreaHeaderActionButtons}
            siteAdminAreaRoutes={enterpriseSiteAdminAreaRoutes}
            siteAdminSideBarGroups={enterpriseSiteAdminSidebarGroups}
            siteAdminOverviewComponents={enterpriseSiteAdminOverviewComponents}
            userAreaHeaderNavItems={enterpriseUserAreaHeaderNavItems}
            userAreaRoutes={enterpriseUserAreaRoutes}
            userSettingsSideBarItems={enterpriseUserSettingsSideBarItems}
            userSettingsAreaRoutes={enterpriseUserSettingsAreaRoutes}
            repoContainerRoutes={enterpriseRepoContainerRoutes}
            repoRevContainerRoutes={enterpriseRepoRevContainerRoutes}
            repoHeaderActionButtons={enterpriseRepoHeaderActionButtons}
            routes={enterpriseRoutes}
            keybindings={keybindings}
        />,
        document.querySelector('#root')
    )
})
