import { WrappedStatus } from '../../../../shared/src/api/client/services/statusService'

export const urlToStatus = (areaURL: string, status: Pick<WrappedStatus, 'name'> | 'string'): string =>
    `${areaURL}/${typeof status === 'string' ? status : status.name}`
