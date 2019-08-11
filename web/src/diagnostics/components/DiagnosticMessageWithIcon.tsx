import { Diagnostic } from '@sourcegraph/extension-api-types'
import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { DiagnosticSeverityIcon } from './DiagnosticSeverityIcon'

interface Props {
    diagnostic: Diagnostic | sourcegraph.Diagnostic

    className?: string
}

export const DiagnosticMessageWithIcon: React.FunctionComponent<Props> = ({ diagnostic, className = '' }) => (
    <div className={`d-flex align-items-start ${className}`}>
        <DiagnosticSeverityIcon severity={diagnostic.severity} className="icon-inline mr-2" />
        <span>{diagnostic.message}</span>
    </div>
)
