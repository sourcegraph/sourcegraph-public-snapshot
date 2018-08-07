import * as querystring from 'query-string'
import storage from '../../extension/storage'

const search = window.location.search
const searchParams = querystring.parse(search)

if (searchParams && searchParams.sourceurl) {
    storage.getSync(items => {
        const serverUrls = items.serverUrls || []
        serverUrls.push(searchParams.sourceurl)
        storage.setSync({
            serverUrls: [...new Set([...serverUrls, 'https://sourcegraph.com'])],
            serverUserId: searchParams.userId || items.serverUserId,
        })
    })
}
