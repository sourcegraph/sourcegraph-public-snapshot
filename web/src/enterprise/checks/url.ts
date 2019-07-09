import { CheckID } from '../../../../shared/src/api/client/services/checkService'

export const urlToCheck = (checksURL: string, check: CheckID): string => `${checksURL}/${check.type}/${check.id}`
