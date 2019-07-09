import { CheckWithType } from '../../../../shared/src/api/client/services/checkService'

export const urlToStatus = (checksURL: string, status: Pick<CheckWithType, 'name'> | string): string =>
    `${checksURL}/${typeof status === 'string' ? status : status.name}`
