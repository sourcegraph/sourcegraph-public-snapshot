// This is the entry point for the web app

import React from 'react'
import { render } from 'react-dom'
import { enterpriseSiteAdminAreaRoutes } from './enterprise/site-admin/routes'
import { SourcegraphWebApp } from './SourcegraphWebApp'

window.addEventListener('DOMContentLoaded', () => {
    render(<SourcegraphWebApp siteAdminAreaRoutes={enterpriseSiteAdminAreaRoutes} />, document.querySelector('#root'))
})
