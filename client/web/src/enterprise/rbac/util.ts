import { startCase } from 'lodash'

export const prettifyNamespace = (namespace: string): string => startCase(namespace.replace(/_/g, ' ').toLowerCase())
export const prettifyAction = (action: string): string => startCase(action.replace(/_/g, ' ').toLowerCase())
