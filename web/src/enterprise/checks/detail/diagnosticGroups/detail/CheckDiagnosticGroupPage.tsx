import H from 'history'
import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../../shared/src/platform/context'
import { ThemeProps } from '../../../../../theme'
import { DiagnosticsList } from '../../../../tasks/list/DiagnosticsList'
import { CheckAreaContext } from '../../CheckArea'
import { CheckDiagnosticGroup } from '../list/CheckDiagnosticGroup'
import { useDiagnostics } from './useDiagnostics'

interface Props
    extends Pick<CheckAreaContext, 'checkProvider'>,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    diagnosticGroup: sourcegraph.DiagnosticGroup
    checkDiagnosticsURL: string

    className?: string
    history: H.History
    location: H.Location
}

/**
 * A page for a check diagnostic group.
 */
export const CheckDiagnosticGroupPage: React.FunctionComponent<Props> = ({
    diagnosticGroup,
    checkProvider,
    className = '',
    extensionsController,
    ...props
}) => {
    const diagnosticsOrError = useDiagnostics(extensionsController, diagnosticGroup.query)
    return (
        <div className={`${className}`}>
            <CheckDiagnosticGroup
                {...props}
                diagnosticGroup={diagnosticGroup}
                className="card my-5"
                contentClassName="card-body"
                extensionsController={extensionsController}
            />
            <DiagnosticsList
                {...props}
                diagnosticsOrError={diagnosticsOrError}
                itemClassName="container-fluid"
                extensionsController={extensionsController}
            />
        </div>
    )
}
