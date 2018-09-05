// This is the entry point for the web app

import React from 'react'
import { render } from 'react-dom'
import { enterpriseSiteAdminAreaRoutes } from './enterprise/site-admin/routes'
import { enterpriseSiteAdminSidebarNavItems } from './enterprise/site-admin/sidebaritems'
import { SourcegraphWebApp } from './SourcegraphWebApp'

window.addEventListener('DOMContentLoaded', () => {
    render(
        <SourcegraphWebApp
            siteAdminAreaRoutes={enterpriseSiteAdminAreaRoutes}
            siteAdminSideBarItems={enterpriseSiteAdminSidebarNavItems}
        />,
        document.querySelector('#root')
    )
})
