import fs from 'fs'
import path from 'path'

import utils from '@percy/sdk-utils'
import puppeteer, { JSONObject } from 'puppeteer'

// import shelljs from 'shelljs'

import { WORKSPACES_PATH } from '@sourcegraph/build-config'

const outsiteOfBazel = require('/Users/val/Desktop/sourcegraph-root/sourcegraph/client/testing/src/percySnapshotToDisk2.js')
console.log(path, fs, utils, WORKSPACES_PATH)
// const puppeteerPkg = require('puppeteer/package.json')
// const CLIENT_INFO = 'percySnapshotToDisk/0.0.1'
// const ENV_INFO = `${puppeteerPkg.name}/${puppeteerPkg.version}`

// console.log('CWD in ', process.cwd())
// const SNAPSHOT_DIR = path.join(WORKSPACES_PATH, 'shared', 'percy-snapshots')
// console.log({ SNAPSHOT_DIR })
// const SNAPSHOT_FILE = path.join(SNAPSHOT_DIR, 'snapshots.snap')
// shelljs.mkdir('-p', SNAPSHOT_DIR)

// Take a DOM snapshot and save it to disk
export async function percySnapshotToDisk(page: puppeteer.Page, name: string, options: JSONObject) {
    outsiteOfBazel.percySnapshotToDisk(page, name, options)
    // if (!page) throw new Error('A Puppeteer `page` object is required.')
    // if (!name) throw new Error('The `name` argument is required.')
    // if (!(await utils.isPercyEnabled())) return
    // let log = utils.logger('puppeteer')

    // try {
    //     // Inject the DOM serialization script
    //     await page.evaluate(await utils.fetchPercyDOM())

    //     // Serialize and capture the DOM
    //     let domSnapshot = await page.evaluate(options => {
    //         return (window as any).PercyDOM.serialize(options)
    //     }, options)

    //     appendJsonToFile(SNAPSHOT_FILE, {
    //         ...options,
    //         environmentInfo: ENV_INFO,
    //         clientInfo: CLIENT_INFO,
    //         url: page.url(),
    //         domSnapshot,
    //         name,
    //     })
    // } catch (err) {
    //     log.error(`Could not take DOM snapshot "${name}"`)
    //     log.error(err)
    // }
}

// function appendJsonToFile(filePath: string, jsonObject: JSONObject) {
//     const jsonString = JSON.stringify(jsonObject)
//     const separator = fs.existsSync(filePath) ? '\n' : ''

//     try {
//         fs.appendFileSync(filePath, separator + jsonString)
//         console.log(`JSON object appended to the file: ${filePath}`)
//     } catch (err) {
//         console.error(`Error appending JSON object to the file: ${err}`)
//     }
// }
