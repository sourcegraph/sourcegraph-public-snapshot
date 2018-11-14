import createWebstoreClient from 'chrome-webstore-upload'
import * as fs from 'fs'
import * as path from 'path'

const release = async (extensionId: string, asset: string) => {
    const {
        GOOGLE_CLIENT_ID: clientId,
        GOOGLE_CLIENT_SECRET: clientSecret,
        GOOGLE_REFRESH_TOKEN: refreshToken,
    } = process.env

    if (!clientId) {
        throw new Error('GOOGLE_CLIENT_ID not set')
    }

    if (!clientSecret) {
        throw new Error('GOOGLE_CLIENT_SECRET not set')
    }

    if (!refreshToken) {
        throw new Error('GOOGLE_REFRESH_TOKEN not set')
    }

    const webStore = await createWebstoreClient({
        clientId,
        clientSecret,
        extensionId,
        refreshToken,
    })

    const token = await webStore.fetchToken()

    const zipFile = fs.createReadStream(asset)

    const uploadRes = await webStore.uploadExisting(zipFile, token)

    if (uploadRes.uploadState === 'FAILURE') {
        throw Object.assign(new Error(uploadRes.itemError.map(e => e.error_detail).join('\n')), {
            name: 'AggregateError' as 'AggregateError',
        })
    }

    const publishRes = await webStore.publish('default', token)

    if (!publishRes.status.includes('OK')) {
        throw Object.assign(new Error(publishRes.statusList.join('\n')), {
            name: 'AggregateError' as 'AggregateError',
        })
    }

    return `Publish complete: https://chrome.google.com/webstore/detail/${extensionId}`
}

release('dgjhfomjieaadpoljlnidmbgkdffpack', path.resolve(__dirname, '../build/bundles/chrome-bundle.zip'))
    .then(res => console.log(res))
    .catch(err => {
        throw err
    })
