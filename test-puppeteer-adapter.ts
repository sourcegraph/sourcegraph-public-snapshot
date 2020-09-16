import { Polly } from '@pollyjs/core'
import { PuppeteerAdapter } from './shared/src/testing/integration/polly/PuppeteerAdapter'
import { ResourceType } from 'puppeteer'
import FSPersister from '@pollyjs/persister-fs'
import puppeteer from 'puppeteer'

async function run() {
    // register uses `adapter.id()` to register a service by its id
    Polly.register(PuppeteerAdapter as any)
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

    const record = true

    const browser = await puppeteer.launch()
    const page = await browser.newPage()

    const polly = new Polly('testing puppeteer adapter', {
        adapters: ['puppeteer'], // provided by PuppeteerAdapter
        adapterOptions: {
            puppeteer: {
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
        matchRequestsBy: {
            method: true,
            body: true,
            order: true,
            // Origin header will change when running against a test instance
            headers: false,
        },
        mode: record ? 'record' : 'replay',
        logging: false,
    })

    polly

    await page.goto('https://sourcegraph.com')
    console.log('Done')
    await page.close()

    await browser.close()
}

run().catch(error => {
    console.error(error)
})
