import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import AlertIcon from 'mdi-react/AlertIcon'
import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import React from 'react'
import { DiagnosticSeverity as DiagnosticSeverityType } from 'sourcegraph'
import { DiagnosticSeverity } from '../../../../shared/src/api/types/diagnosticCollection'

const DEFAULT_SEVERITY: DiagnosticSeverityType = DiagnosticSeverity.Error

interface SeverityInfo {
    icon: React.ComponentType<{ className?: string }>
    tooltip: string
    className: string
}

const INFO: Record<DiagnosticSeverityType, SeverityInfo> = {
    [DiagnosticSeverity.Error]: { icon: AlertCircleIcon, tooltip: 'Error', className: 'text-danger' },
    [DiagnosticSeverity.Warning]: { icon: AlertIcon, tooltip: 'Warning', className: 'text-warning' },
    [DiagnosticSeverity.Information]: { icon: InformationOutlineIcon, tooltip: 'Info', className: 'text-info' },
    [DiagnosticSeverity.Hint]: { icon: HelpCircleOutlineIcon, tooltip: 'Hint', className: 'text-success' },
}

/**
 * An icon representing the severity of a diagnostic.
 */
export const DiagnosticSeverityIcon: React.FunctionComponent<{
    severity: DiagnosticSeverityType
    className?: string
}> = ({ severity, className = '' }) => {
    const { icon: Icon, tooltip, className: severityClassName } = INFO[severity] || INFO[DEFAULT_SEVERITY]
    return <Icon className={`${severityClassName} ${className}`} aria-label={tooltip} />
}
