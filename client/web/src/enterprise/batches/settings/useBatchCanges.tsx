import { queryUserBatchChangesCodeHosts } from './backend'
import fetcher from '../../../client'
import { QueryObserverResult, useQuery } from 'react-query'

export const useBatchChanges = (userID: string): QueryObserverResult => {
    const data = useQuery([queryUserBatchChangesCodeHosts, { user: userID, first: 10, after: '1' }], fetcher)
    return data
}
