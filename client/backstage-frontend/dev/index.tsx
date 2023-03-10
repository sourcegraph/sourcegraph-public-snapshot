import React from 'react'

import { createDevApp } from '@backstage/dev-utils'

import { sourcegraphPlugin, SourcegraphPage } from '../src/plugin'

createDevApp()
    .registerPlugin(sourcegraphPlugin)
    .addPage({
        element: <SourcegraphPage />,
        title: 'Root Page',
        path: '/william',
    })
    .render()
