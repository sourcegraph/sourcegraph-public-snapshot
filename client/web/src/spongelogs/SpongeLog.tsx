import * as React from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { useParams } from 'react-router-dom'

//import { useTable } from 'react-table'

import { renderMarkdown } from '@sourcegraph/common'
import { ErrorMessage, LoadingSpinner } from '@sourcegraph/wildcard'

import { HeroPage } from '../components/HeroPage'

import { useSpongeLog } from './backend'

import styles from './SpongeLog.module.scss'

export const SpongeLog: React.FunctionComponent<{}> = () => {
    const { uuid } = useParams<{ uuid: string }>()
    const { data, loading, error } = useSpongeLog(uuid ?? '')
    if (uuid === undefined) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle="UUID undefined LOLOLO" />
    }
    if (loading) {
        return <LoadingSpinner />
    }
    if (error) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={<ErrorMessage error={error} />} />
    }
    // This is custom for autocomplete - should later dispatch on data.spongeLog.interpreter.
    const log = JSON.parse(data?.spongeLog?.log || '[]')
    const codeToCompletions: any = {}
    for (const [_, { completions, timestamp, sample, ...rest }] of log.entries()) {
        if (!codeToCompletions[sample.content]) {
            codeToCompletions[sample.content] = []
        }
        codeToCompletions[sample.content].push({ ...rest, sample, completions, timestamp })
    }
    console.log(codeToCompletions)
    // Return JSON pretty printed in a <code> tag
    return <Completions completions={codeToCompletions} />
    // return (
    //     <pre>
    //         <code>{JSON.stringify(log)}</code>
    //     </pre>
    // )
}

interface Column {
    Header: string
    accessor: string
}

interface Row {
    code: string
    bgColor: string
    [key: string]: string
}

const Completions: React.FunctionComponent<{ completions: any }> = ({ completions }) => {
    const columns: Column[] = [
        {
            Header: 'Code',
            accessor: 'code',
        },
    ]

    // Dynamically generated columns based on the number of generated files.
    const extraColumns = new Set<string>()

    const formatData = (inputData: typeof completions): Row[] => {
        const rows: Row[] = []

        for (const [code, entries] of Object.entries(inputData)) {
            const row: Row = {
                code: renderMarkdown('```\n' + code + '\n```'),
                bgColor: '#f7f7f7',
            }

            // For now assume just one entry
            for (const { completions, elapsed } of entries as any[]) {
                const columnKey = 'Completions'

                extraColumns.add(columnKey)

                // All completions are displayed in the same cell separated by <hr />.
                row[columnKey] = `<div class="elapsed">${elapsed}ms</div>${completions
                    .map(renderMarkdown)
                    .join('<hr />')}`
            }

            rows.push(row)
        }

        columns.push(...Array.from(extraColumns).map(columnKey => ({ Header: columnKey, accessor: columnKey })))

        return rows
    }

    const rows = formatData(completions)

    // const { getTableProps, getTableBodyProps, headerGroups, rows, prepareRow } = useTable({
    //     columns,
    //     data: formatData(completions),
    // })

    return (
        <div className={styles.tableWrapper}>
            <table /*{...getTableProps()}*/ className={styles.table}>
                <thead>
                    <tr>
                        {columns.map(column => (
                            <th key={column.accessor}>{column.Header}</th>
                        ))}
                        <th></th>
                    </tr>
                </thead>

                <tbody /*{...getTableBodyProps()}*/>
                    {rows.map(row => (
                        <tr>
                            {columns.map(column => (
                                <td
                                    /* eslint-disable react/forbid-dom-props */
                                    style={{ maxWidth: 800, overflow: 'auto' }}
                                    key={column.accessor}
                                    dangerouslySetInnerHTML={{
                                        __html: row[column.accessor],
                                    }}
                                />
                            ))}
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    )
}
