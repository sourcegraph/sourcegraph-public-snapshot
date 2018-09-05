// This is the entry point for the web app

import React from 'react'
import { render } from 'react-dom'
import { siteAdminAreaRoutes } from './site-admin/routes'
import { SourcegraphWebApp } from './SourcegraphWebApp'

window.addEventListener('DOMContentLoaded', () => {
    render(<SourcegraphWebApp siteAdminAreaRoutes={siteAdminAreaRoutes} />, document.querySelector('#root'))
})
