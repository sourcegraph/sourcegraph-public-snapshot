import { Dispatch, SetStateAction } from 'react'

import type { ClientInit } from '@sourcegraph/cody-shared/src/chat/client'
import { useLocalStorage } from '@sourcegraph/wildcard'

const DEFAULT_WEB_CONFIGURATION: WebConfiguration = {
    serverEndpoint: 'https://sourcegraph.com',
    accessToken: null,
    useContext: 'embeddings',
    customHeaders: {},
}

export interface WebConfiguration
    extends Pick<ClientInit, 'accessToken' | 'customHeaders'>,
        Omit<ClientInit['config'], never> {}

export function useConfig(): [WebConfiguration, Dispatch<SetStateAction<WebConfiguration>>] {
    // eslint-disable-next-line no-restricted-syntax
    return useLocalStorage<WebConfiguration>('cody-web.config', DEFAULT_WEB_CONFIGURATION)
}
