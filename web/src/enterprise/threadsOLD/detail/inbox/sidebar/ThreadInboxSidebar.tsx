import React from 'react'
import { DiagnosticSeverity } from 'sourcegraph'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { QueryParameterProps } from '../../../../../components/withQueryParameter/WithQueryParameter'
import { TreeFilterSidebar } from '../../../components/treeFilterSidebar/TreeFilterSidebar'
import { DiagnosticInfo } from '../../backend'
import { ThreadInboxSidebarFilterListDiagnosticItem } from './ThreadInboxSidebarFilterListDiagnosticItem'
import { ThreadInboxSidebarFilterListPathItem } from './ThreadInboxSidebarFilterListPathItem'
import { ThreadInboxSidebarFilterListRepositoryItem } from './ThreadInboxSidebarFilterListRepositoryItem'

/**
 * Reports whether the diagnostic passes the filter.
 */
export const diagnosticFilter = (query: string, diagnostic: DiagnosticInfo): boolean => {
    query = query.toLowerCase()
    return (
        diagnostic.message.toLowerCase().includes(query) ||
        diagnostic.entry.repository.name.toLowerCase().includes(query) ||
        diagnostic.entry.path.toLowerCase().includes(query)
    )
}

interface Props extends QueryParameterProps {
    diagnostics: DiagnosticInfo[]

    className?: string
}

/**
 * The sidebar for the thread inbox.
 */
export const ThreadInboxSidebar: React.FunctionComponent<Props> = ({
    diagnostics,
    query,
    onQueryChange,
    className,
    ...props
}) => {
    diagnostics = diagnostics.filter(diagnostic => diagnosticFilter(query, diagnostic))
    return (
        <TreeFilterSidebar query={query} onQueryChange={onQueryChange} className={className} {...props}>
            {({ query, className }) => (
                <>
                    {uniqueMessages(diagnostics).map(([{ message, severity }, count], i) => (
                        <ThreadInboxSidebarFilterListDiagnosticItem
                            key={i}
                            diagnostic={{ message, severity }}
                            count={count}
                            query={query}
                            onQueryChange={onQueryChange}
                            className={className}
                        />
                    ))}
                    {uniqueRepos(diagnostics).map(([repository, count], i) => (
                        <ThreadInboxSidebarFilterListRepositoryItem
                            key={i}
                            repository={repository}
                            count={count}
                            query={query}
                            onQueryChange={onQueryChange}
                            className={className}
                        />
                    ))}
                    {uniqueFiles(diagnostics).map(([path, count], i) => (
                        <ThreadInboxSidebarFilterListPathItem
                            key={i}
                            path={path}
                            count={count}
                            query={query}
                            onQueryChange={onQueryChange}
                            className={className}
                        />
                    ))}
                </>
            )}
        </TreeFilterSidebar>
    )
}

function uniqueMessages(diagnostics: DiagnosticInfo[]): [Pick<DiagnosticInfo, 'message' | 'severity'>, number][] {
    const messages = new Map<string, number>()
    const severity = new Map<string, DiagnosticSeverity>()
    for (const d of diagnostics) {
        const count = messages.get(d.message) || 0 // TODO!(sqs): hacky, doesnt support multi repos
        messages.set(d.message, count + 1)
        severity.set(d.message, d.severity)
    }
    return Array.from(messages.entries())
        .sort((a, b) => a[1] - b[1])
        .map(([message, count]) => [{ message, severity: severity.get(message)! }, count])
}

function uniqueRepos(diagnostics: DiagnosticInfo[]): [Pick<GQL.IRepository, 'name'>, number][] {
    const files = new Map<string, number>()
    for (const d of diagnostics) {
        const count = files.get(d.entry.path) || 0 // TODO!(sqs): hacky, doesnt support multi repos
        files.set(d.entry.repository.name, count + 1)
    }
    return Array.from(files.entries())
        .map(([repo, count]) => [{ name: repo }, count] as [Pick<GQL.IRepository, 'name'>, number])
        .sort((a, b) => a[1] - b[1])
}

function uniqueFiles(diagnostics: DiagnosticInfo[]): [string, number][] {
    const files = new Map<string, number>()
    for (const d of diagnostics) {
        const count = files.get(d.entry.path) || 0 // TODO!(sqs): hacky, doesnt support multi repos
        files.set(d.entry.path, count + 1)
    }
    return Array.from(files.entries()).sort((a, b) => a[1] - b[1])
}
