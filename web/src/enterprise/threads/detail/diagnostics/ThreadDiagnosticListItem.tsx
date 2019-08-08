import H from 'history'
import React from 'react'
import { toDiagnostic } from '../../../../../../shared/src/api/types/diagnostic'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { ThemeProps } from '../../../../theme'
import { DiagnosticsListItem } from '../../../tasks/list/item/DiagnosticsListItem'

interface Props extends ExtensionsControllerProps, PlatformContextProps, ThemeProps {
    threadDiagnostic: GQL.IThreadDiagnosticEdge

    className?: string
    history: H.History
    location: H.Location
}

export const ThreadDiagnosticListItem: React.FunctionComponent<Props> = ({
    threadDiagnostic,
    className = '',
    ...props
}) => (
    <div className={`thread-diagnostic-list-item ${className}`}>
        <DiagnosticsListItem
            {...props}
            diagnostic={{ ...threadDiagnostic.diagnostic.data, ...toDiagnostic(threadDiagnostic.diagnostic.data) }}
            selectedAction={null}
            // tslint:disable-next-line: jsx-no-lambda
            onActionSelect={() => void 0}
        />
    </div>
)
