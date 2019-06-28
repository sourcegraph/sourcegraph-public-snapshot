import H from 'history'
import React from 'react'
import { LinkOrSpan } from '../../../../../../shared/src/components/LinkOrSpan'
import { displayRepoName } from '../../../../../../shared/src/components/RepoFileLink'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { DiagnosticSeverityIcon } from '../../../../diagnostics/components/DiagnosticSeverityIcon'
import { DiagnosticInfo } from '../../../threads/detail/backend'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    diagnostic: DiagnosticInfo

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
    diagnostic,
    className = '',
    headerClassName = '',
    headerStyle,
}) => (
    <div className={`d-flex flex-wrap align-items-stretch ${className}`}>
        {/* tslint:disable-next-line: jsx-ban-props */}
        <div style={{ flex: '1 1 40%', minWidth: '400px' }} className="pr-5">
            {/* tslint:disable-next-line: jsx-ban-props */}
            <header className={`d-flex align-items-start ${headerClassName}`} style={headerStyle}>
                <div className={`flex-1 d-flex align-items-center`}>
                    <h3 className="mb-0 small">
                        <LinkOrSpan to={diagnostic.entry.url} className="d-block">
                            {diagnostic.entry.path ? (
                                <>
                                    <span className="font-weight-normal">
                                        {displayRepoName(diagnostic.entry.repository.name)}
                                    </span>{' '}
                                    › {diagnostic.entry.path}
                                </>
                            ) : (
                                displayRepoName(diagnostic.entry.repository.name)
                            )}
                        </LinkOrSpan>
                    </h3>
                </div>
            </header>
            <div className={`d-flex align-items-start mt-2 mb-1`}>
                <DiagnosticSeverityIcon severity={diagnostic.severity} className="icon-inline mr-2" />
                <span>{diagnostic.message}</span>
            </div>
        </div>
    </div>
)
