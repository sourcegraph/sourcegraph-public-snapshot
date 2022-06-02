import { FunctionComponent, ReactNode } from 'react'

import classNames from 'classnames'
import DatabaseEditIcon from 'mdi-react/DatabaseEditIcon'
import DatabasePlusIcon from 'mdi-react/DatabasePlusIcon'

import { Code, Container } from '@sourcegraph/wildcard'

import { Timeline } from '../../../../components/Timeline'
import { AuditLogOperation, LsifUploadsAuditLogsFields } from '../../../../graphql-operations'

export interface UploadAuditLogTimelineProps {
    logs: LsifUploadsAuditLogsFields[]
}

export const UploadAuditLogTimeline: FunctionComponent<React.PropsWithChildren<UploadAuditLogTimelineProps>> = ({
    logs,
}) => {
    const stages = logs?.map(log => ({
        icon: log.operation === AuditLogOperation.CREATE ? <DatabasePlusIcon /> : <DatabaseEditIcon />,
        text: stageText(log),
        date: log.logTimestamp,
        details: (
            <>
                {log.reason && (
                    <>
                        <Container>
                            <b>Reason</b>: {log.reason}
                        </Container>
                        <br />
                    </>
                )}
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
                                    <pre className={classNames('mb-0 position-relative')}>{change.column}</pre>
                                </td>
                                <td className="mr-2">
                                    <pre className={classNames('mb-0 position-relative')}>{change.old || 'NULL'}</pre>
                                </td>
                                <td className="mr-2">
                                    <pre className={classNames('mb-0 position-relative')}>{change.new || 'NULL'}</pre>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </>
        ),
        className: log.operation === AuditLogOperation.CREATE ? 'bg-success' : 'bg-warning',
    }))

    return <Timeline withoutPreviousDate={true} stages={stages} />
}

function stageText(log: LsifUploadsAuditLogsFields): ReactNode {
    if (log.operation === AuditLogOperation.CREATE) {
        return 'Upload created'
    }

    return (
        <>
            Altered columns:{' '}
            {formatReactNodeList(log.changedColumns.map(change => <Code key={change.column}>{change.column}</Code>))}
        </>
    )
}

function formatReactNodeList(list: ReactNode[]): ReactNode {
    if (list.length === 0) {
        return <></>
    }
    if (list.length === 1) {
        return list[0]
    }

    return (
        <>
            {list.slice(0, -1).reduce((previous, current) => [previous, ', ', current])} and {list[list.length - 1]}
        </>
    )
}
