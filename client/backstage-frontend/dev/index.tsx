import { createDevApp } from '@backstage/dev-utils'
import React from 'react'
import { SourcegraphPage, sourcegraphPlugin } from '../src/plugin'

createDevApp()
    .registerPlugin(sourcegraphPlugin)
    .addPage({
        element: <SourcegraphPage />,
        title: 'Root Page',
        path: '/william',
    })
    .render()
