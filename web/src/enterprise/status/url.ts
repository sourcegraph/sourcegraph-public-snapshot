import { WrappedStatus } from '../../../../shared/src/api/client/services/statusService'

export const urlToStatus = (statusesURL: string, status: Pick<WrappedStatus, 'name'> | string): string =>
    `${statusesURL}/${typeof status === 'string' ? status : status.name}`
