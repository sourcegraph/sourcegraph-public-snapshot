import React, { type FC } from 'react'

import { Routes, Route, Navigate } from 'react-router-dom'
import type { Observable } from 'rxjs'

import type { FetchFileParameters } from '@sourcegraph/shared/src/backend/file'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { SearchContextProps } from '@sourcegraph/shared/src/search'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import type { AuthenticatedUser } from '../auth'
import type { OwnConfigProps } from '../own/OwnConfigProps'
import { PageRoutes } from '../routes.constants'
import type { SearchStreamingProps } from '../search'

import type { NotebookProps } from '.'

const NotebookPage = lazyComponent(() => import('./notebookPage/NotebookPage'), 'NotebookPage')
const CreateNotebookPage = lazyComponent(() => import('./createPage/CreateNotebookPage'), 'CreateNotebookPage')
const NotebooksListPage = lazyComponent(() => import('./listPage/NotebooksListPage'), 'NotebooksListPage')

export interface GlobalNotebooksAreaProps
    extends TelemetryProps,
        PlatformContextProps,
        SettingsCascadeProps,
        NotebookProps,
        SearchStreamingProps,
        Pick<SearchContextProps, 'searchContextsEnabled'>,
        OwnConfigProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
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
                    <Navigate to={PageRoutes.Notebooks} replace={true} />
                )
            }
        />
        <Route path=":id" element={<NotebookPage authenticatedUser={authenticatedUser} {...outerProps} />} />
    </Routes>
)
