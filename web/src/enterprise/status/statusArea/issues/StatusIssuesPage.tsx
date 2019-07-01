import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { TasksList } from '../../../tasks/list/TasksList'
import { StatusAreaContext } from '../StatusArea'

interface Props extends Pick<StatusAreaContext, 'status'>, ExtensionsControllerProps, PlatformContextProps {
    className?: string
    history: H.History
    location: H.Location
    isLightTheme: boolean
}

/**
 * The status issues page.
 */
export const StatusIssuesPage: React.FunctionComponent<Props> = ({ status, className = '', ...props }) => (
    <div className={`status-issues-page ${className}`}>
        <TasksList {...props} itemClassName="container-fluid" />
    </div>
)
