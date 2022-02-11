import { of } from 'rxjs'

import { proxySubscribable } from '@sourcegraph/shared/src/api/extension/api/common'

import { SearchSidebarAPI } from '../../contract'

export function createSearchSidebarAPI(): SearchSidebarAPI {
    return {
        ping: () => {
            console.log('ping called')
            return proxySubscribable(of('pong'))
        },
        addTextDocumentIfNotExists: () => {
            console.log('addTextDocumentIfNotExists called')
        },
    }
}
