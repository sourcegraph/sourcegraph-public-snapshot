import { Polly } from '@pollyjs/core'
import { CdpAdapter } from './shared/src/testing/integration/polly/CdpAdapter'

import { ResourceType } from 'puppeteer'
import FSPersister from '@pollyjs/persister-fs'
import puppeteer from 'puppeteer'

async function run() {
    // register uses `adapter.id()` to register a service by its id
    Polly.register(CdpAdapter as any)
    Polly.register(FSPersister)

    const recordingsDirectory = './__fixtures__1'

    // This is copied as-is
    const requestResourceTypes: ResourceType[] = [
        'xhr',
        'fetch',
        'document',
        'script',
        'stylesheet',
        'image',
        'font',
        'other', // Favicon
    ]

    const record = false

    const browser = await puppeteer.launch()
    const page = await browser.newPage()

    const polly = new Polly('testing CDP adapter', {
        adapters: ['cdp'], // provided by CdpAdapter
        adapterOptions: {
            cdp: {
                page: page,
                requestResourceTypes,
            },
        },
        persister: 'fs', // provided by FSPersister
        persisterOptions: {
            fs: {
                recordingsDir: recordingsDirectory,
            },
        },
        expiryStrategy: 'warn',
        recordIfMissing: record,
        recordFailedRequests: true,
        matchRequestsBy: {
            method: true,
            body: true,
            order: true,
            // Origin header will change when running against a test instance
            headers: false,
        },
        mode: record ? 'record' : 'replay',
        logging: true,
    })

    await page.goto('https://example.com')
    console.log('Done loading URL,...')
    await page.close()

    await browser.close()

    await polly.stop()
}

run().catch(error => {
    console.error(error)
})
