import { Diagnostic } from '@sourcegraph/extension-api-types'
import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { DiagnosticMessageWithIcon } from '../components/DiagnosticMessageWithIcon'

interface Props extends ExtensionsControllerProps {
    diagnostic: Diagnostic | sourcegraph.Diagnostic
}

export const DiagnosticListItem: React.FunctionComponent<Props> = ({ diagnostic, ...props }) => (
    <DiagnosticMessageWithIcon diagnostic={diagnostic} />
)
