import H from 'history'
import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { ThemeProps } from '../../../../theme'
import { CheckAreaContext } from '../CheckArea'
import { DiagnosticsListPage } from './DiagnosticsListPage'

interface Props
    extends Pick<CheckAreaContext, 'checkID' | 'checkProvider' | 'checkInfo'>,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    className?: string
    history: H.History
    location: H.Location
}

/**
 * The check diagnostics page.
 */
export const CheckDiagnosticsPage: React.FunctionComponent<Props> = ({
    checkID,
    checkProvider,
    checkInfo,
    className = '',
    ...props
}) => {
    const baseDiagnosticQuery: sourcegraph.DiagnosticQuery = { type: checkID.type }
    return <DiagnosticsListPage {...props} baseDiagnosticQuery={baseDiagnosticQuery} />
}
