import { GraphQLRequestOptions } from '../../../../../shared/src/graphql/graphql'
import { getHeaders } from './headers'

export const requestOptions: GraphQLRequestOptions = {
    headers: getHeaders(),
    requestOptions: {
        crossDomain: true,
        withCredentials: true,
        async: true,
    },
}
