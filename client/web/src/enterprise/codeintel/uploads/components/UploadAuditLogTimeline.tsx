import { FunctionComponent } from 'react'

import classNames from 'classnames'
import DatabaseEditIcon from 'mdi-react/DatabaseEditIcon'
import DatabasePlusIcon from 'mdi-react/DatabasePlusIcon'

import { Timeline } from '../../../../components/Timeline'
import { AuditLogOperation, LsifUploadsAuditLogsFields } from '../../../../graphql-operations'

export interface UploadAuditLogTimelineProps {
    logs: LsifUploadsAuditLogsFields[]
}

interface stateChange {
    column: string
    old?: string
    new?: string
}

export const UploadAuditLogTimeline: FunctionComponent<React.PropsWithChildren<UploadAuditLogTimelineProps>> = ({
    logs,
}) => {
    const stages = logs?.map(log => ({
        icon: log.operation === AuditLogOperation.CREATE ? <DatabasePlusIcon /> : <DatabaseEditIcon />,
        text: log.logTimestamp,
        date: log.logTimestamp,
        details: (
            <table className="table mb-0">
                <thead>
                    <tr>
                        <th>Column</th>
                        <th>Old</th>
                        <th>New</th>
                    </tr>
                </thead>
                <tbody>
                    {log.changedColumns.map((change, index) => (
                        // eslint-disable-next-line react/no-array-index-key
                        <tr key={index}>
                            <td className="mr-2">
                                <pre className={classNames('mb-0 position-relative')}>
                                    {(change as stateChange).column}
                                </pre>
                            </td>
                            <td className="mr-2">
                                <pre className={classNames('mb-0 position-relative')}>
                                    {(change as stateChange).old || 'NULL'}
                                </pre>
                            </td>
                            <td className="mr-2">
                                <pre className={classNames('mb-0 position-relative')}>
                                    {(change as stateChange).new || 'NULL'}
                                </pre>
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
        ),
        className: log.operation === AuditLogOperation.CREATE ? 'bg-success' : 'bg-warning',
    }))

    return <Timeline withoutPreviousDate={true} stages={stages} />
}
