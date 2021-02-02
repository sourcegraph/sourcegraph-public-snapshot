import React from 'react'
import { RepoContainerRoute } from '../../repo/RepoContainer'
import { RepoRevisionContainerRoute } from '../../repo/RepoRevisionContainer'
import { repoContainerRoutes, repoRevisionContainerRoutes } from '../../repo/routes'
import { SymbolsArea } from '../symbols/SymbolsArea'

export const enterpriseRepoContainerRoutes: readonly RepoContainerRoute[] = repoContainerRoutes

export const enterpriseRepoRevisionContainerRoutes: readonly RepoRevisionContainerRoute[] = [
    ...repoRevisionContainerRoutes,
    {
        path: '/-/docs',
        render: context => {
            console.log('# HERe')
            return <SymbolsArea {...context} />
        },
    },
]
