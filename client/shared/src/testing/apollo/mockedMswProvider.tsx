import { type FC, type PropsWithChildren, useMemo } from 'react'

import { ApolloClient, ApolloProvider } from '@apollo/client'

import { generateCache } from '../../backend/apolloCache'

export const MockedMSWProvider: FC<PropsWithChildren<{}>> = ({ children }) => {
    const client = useMemo(
        () =>
            new ApolloClient({
                uri: 'http://0.0.0.0:8080/graphql',
                cache: generateCache(),
            }),
        []
    )
    return <ApolloProvider client={client}>{children}</ApolloProvider>
}
