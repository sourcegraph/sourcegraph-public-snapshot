import type { FC } from 'react'

import { Route, Routes } from 'react-router-dom'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { NotFoundPage } from '../../components/HeroPage'

import type { ExecutorsListPageProps } from './instances/ExecutorsListPage'
import type { GlobalExecutorSecretsListPageProps } from './secrets/ExecutorSecretsListPage'

const ExecutorsListPage = lazyComponent<ExecutorsListPageProps, 'ExecutorsListPage'>(
    () => import('./instances/ExecutorsListPage'),
    'ExecutorsListPage'
)

const GlobalExecutorSecretsListPage = lazyComponent<
    GlobalExecutorSecretsListPageProps,
    'GlobalExecutorSecretsListPage'
>(() => import('./secrets/ExecutorSecretsListPage'), 'GlobalExecutorSecretsListPage')

/** The page area for all executors settings in site-admin. */
export const ExecutorsSiteAdminArea: FC = () => (
    <Routes>
        <Route index={true} element={<ExecutorsListPage />} />
        <Route path="secrets" element={<GlobalExecutorSecretsListPage />} />
        <Route path="*" element={<NotFoundPage pageType="settings" />} />
    </Routes>
)
