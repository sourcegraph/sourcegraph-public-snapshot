import React from 'react'
import { DiagnosticSeverity } from 'sourcegraph'
import { TreeFilterSidebar } from '../../../components/treeFilterSidebar/TreeFilterSidebar'
import { QueryParameterProps } from '../../../components/withQueryParameter/WithQueryParameter'
import { DiagnosticInfo } from '../item/ThreadChangedFileItem'

interface Props extends QueryParameterProps {
    diagnostics: DiagnosticInfo[]

    className?: string
}

/**
 * The sidebar for the thread inbox.
 */
export const ThreadInboxSidebar: React.FunctionComponent<Props> = ({ diagnostics, ...props }) => (
    <TreeFilterSidebar {...props}>{({ query, className }) => <>TODO</>}</TreeFilterSidebar>
)
