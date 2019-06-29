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
    history: H.History
    location: H.Location
}

/**
 * An item in a checklist.
 */
export const ChecklistListItem: React.FunctionComponent<Props> = ({ item, className = '', headerClassName = '' }) => (
    <div className={`d-flex flex-wrap align-items-stretch ${className}`}>
        <header className={headerClassName}>
            <h3 className="mb-0 font-weight-normal font-size-base">{item.title}</h3>
        </header>
    </div>
)
