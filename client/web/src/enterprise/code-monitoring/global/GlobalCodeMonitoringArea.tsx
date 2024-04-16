import React from 'react'

import { Routes, Route } from 'react-router-dom'

import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import type { AuthenticatedUser } from '../../../auth'
import { Page } from '../../../components/Page'

interface Props extends TelemetryProps, PlatformContextProps, SettingsCascadeProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
}

const CodeMonitoringPage = lazyComponent(() => import('../CodeMonitoringPage'), 'CodeMonitoringPage')
const CreateCodeMonitorPage = lazyComponent(() => import('../CreateCodeMonitorPage'), 'CreateCodeMonitorPage')
const ManageCodeMonitorPage = lazyComponent(() => import('../ManageCodeMonitorPage'), 'ManageCodeMonitorPage')

/**
 * The global code monitoring area.
 */
export const GlobalCodeMonitoringArea: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    ...outerProps
}) => (
    <div className="w-100">
        <Page>
            <Routes>
                <Route
                    path=""
                    element={
                        <CodeMonitoringPage
                            {...outerProps}
                            telemetryRecorder={outerProps.platformContext.telemetryRecorder}
                        />
                    }
                />
                <Route
                    path="new"
                    element={
                        <CreateCodeMonitorPage
                            {...outerProps}
                            telemetryRecorder={outerProps.platformContext.telemetryRecorder}
                        />
                    }
                />
                <Route
                    path=":id"
                    element={
                        <ManageCodeMonitorPage
                            {...outerProps}
                            telemetryRecorder={outerProps.platformContext.telemetryRecorder}
                        />
                    }
                />
            </Routes>
        </Page>
    </div>
)
