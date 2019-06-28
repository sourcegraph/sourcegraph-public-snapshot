import H from 'history'
import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { LinkOrSpan } from '../../../../../../shared/src/components/LinkOrSpan'
import { displayRepoName } from '../../../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { DiagnosticSeverityIcon } from '../../../../diagnostics/components/DiagnosticSeverityIcon'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    item: sourcegraph.ChecklistItem

    className?: string
    headerClassName?: string
    headerStyle?: React.CSSProperties
    isLightTheme: boolean
    history: H.History
    location: H.Location
}

/**
 * An item in a checklist.
 */
export const ChecklistListItem: React.FunctionComponent<Props> = ({
    item,
    className = '',
    headerClassName = '',
    headerStyle,
}) => (
    <div className={`d-flex flex-wrap align-items-stretch ${className}`}>
        <h3 className="mb-0">{item.title}</h3>
    </div>
)
