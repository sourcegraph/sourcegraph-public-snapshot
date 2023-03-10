import React, { FC } from 'react'

import { Routes, Route, Navigate } from 'react-router-dom'
import type { Observable } from 'rxjs'

import { FetchFileParameters } from '@sourcegraph/shared/src/backend/file'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SearchContextProps } from '@sourcegraph/shared/src/search'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../auth'
import { EnterprisePageRoutes } from '../routes.constants'
import { SearchStreamingProps } from '../search'

import { NotebookProps } from '.'

const NotebookPage = lazyComponent(() => import('./notebookPage/NotebookPage'), 'NotebookPage')
const CreateNotebookPage = lazyComponent(() => import('./createPage/CreateNotebookPage'), 'CreateNotebookPage')
const NotebooksListPage = lazyComponent(() => import('./listPage/NotebooksListPage'), 'NotebooksListPage')

export interface GlobalNotebooksAreaProps
    extends TelemetryProps,
        PlatformContextProps,
        SettingsCascadeProps,
        NotebookProps,
        SearchStreamingProps,
        Pick<SearchContextProps, 'searchContextsEnabled'> {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
    isSourcegraphApp: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    globbing: boolean
}
/**
 * The global code monitoring area.
 */
export const GlobalNotebooksArea: FC<React.PropsWithChildren<GlobalNotebooksAreaProps>> = ({
    authenticatedUser,
    ...outerProps
}) => (
    <Routes>
        <Route index={true} element={<NotebooksListPage authenticatedUser={authenticatedUser} {...outerProps} />} />
        <Route
            path="new"
            element={
                authenticatedUser ? (
                    <CreateNotebookPage authenticatedUser={authenticatedUser} {...outerProps} />
                ) : (
                    <Navigate to={EnterprisePageRoutes.Notebooks} replace={true} />
                )
            }
        />
        <Route path=":id" element={<NotebookPage authenticatedUser={authenticatedUser} {...outerProps} />} />
    </Routes>
)
