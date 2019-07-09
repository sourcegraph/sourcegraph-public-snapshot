import { WrappedStatus } from '../../../../shared/src/api/client/services/statusService'

export const urlToStatus = (checksURL: string, status: Pick<WrappedStatus, 'name'> | string): string =>
    `${checksURL}/${typeof status === 'string' ? status : status.name}`
