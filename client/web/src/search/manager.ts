import { History } from 'history'
import { aggregateStreamingSearch } from '@sourcegraph/shared/src/search/stream'
import { Controller } from '@sourcegraph/shared/src/extensions/controller'

export class SearchManager {
    constructor(private history: History, private extensionController: Controller) {
        console.debug('initiate manager', performance.now())
        //console.timeLog('search')
    }
}
