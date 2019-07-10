import H from 'history'
import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../../shared/src/platform/context'
import { ThemeProps } from '../../../../../theme'
import { CheckAreaContext } from '../../CheckArea'
import { CheckDiagnosticGroup } from './CheckDiagnosticGroup'

interface Props
    extends Pick<CheckAreaContext, 'checkProvider'>,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    diagnosticGroups: sourcegraph.DiagnosticGroup[]
    expandedDiagnosticGroup?: Pick<sourcegraph.DiagnosticGroup, 'id'>
    checkDiagnosticsURL: string

    className?: string
    itemClassName?: string
    history: H.History
    location: H.Location
}

/**
 * A list of check diagnostics groups.
 */
export const CheckDiagnosticGroupsList: React.FunctionComponent<Props> = ({
    diagnosticGroups,
    expandedDiagnosticGroup,
    checkProvider,
    className = '',
    itemClassName = '',
    ...props
}) => (
    <ul className={`list-unstyled mb-0 ${className}`}>
        {diagnosticGroups.map((diagnosticGroup, i) => (
            <li key={i}>
                <CheckDiagnosticGroup
                    {...props}
                    diagnosticGroup={diagnosticGroup}
                    isExpanded={diagnosticGroup === expandedDiagnosticGroup}
                    className={`${itemClassName}`}
                    contentClassName="container"
                />
            </li>
        ))}
    </ul>
)
