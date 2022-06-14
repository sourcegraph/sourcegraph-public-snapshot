import { FunctionComponent, ReactNode } from 'react'

import DatabaseEditIcon from 'mdi-react/DatabaseEditIcon'
import DatabasePlusIcon from 'mdi-react/DatabasePlusIcon'

import { Container, Icon } from '@sourcegraph/wildcard'

import { Timeline } from '../../../../components/Timeline'
import { AuditLogOperation, LsifUploadsAuditLogsFields } from '../../../../graphql-operations'

import styles from './UploadAuditLogTimeline.module.scss'

export interface UploadAuditLogTimelineProps {
    logs: LsifUploadsAuditLogsFields[]
}

export const UploadAuditLogTimeline: FunctionComponent<React.PropsWithChildren<UploadAuditLogTimelineProps>> = ({
    logs,
}) => {
    const stages = logs?.map(log => ({
        icon:
            log.operation === AuditLogOperation.CREATE ? (
                <Icon as={DatabasePlusIcon} aria-label="Success" />
            ) : (
                <Icon as={DatabaseEditIcon} aria-label="Warn" />
            ),
        text: stageText(log),
        className: log.operation === AuditLogOperation.CREATE ? 'bg-success' : 'bg-warning',
        expanded: true,
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
                <div className={styles.tableContainer}>
                    <table className="table mb-0 table-striped">
                        <thead>
                            <tr>
                                <th className={styles.dbColumnCol} scope="column">
                                    Column
                                </th>
                                <th className={styles.dataColumnCol} scope="column">
                                    Old
                                </th>
                                <th scope="column">New</th>
                            </tr>
                        </thead>
                        <tbody>
                            {log.changedColumns.map((change, index) => (
                                // eslint-disable-next-line react/no-array-index-key
                                <tr key={index} className="overflow-scroll">
                                    <td className="mr-2">{change.column}</td>
                                    <td className="mr-2">{change.old || 'NULL'}</td>
                                    <td className="mr-2">{change.new || 'NULL'}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </>
        ),
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
            {formatReactNodeList(log.changedColumns.map(change => <span key={change.column}>{change.column}</span>))}
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
