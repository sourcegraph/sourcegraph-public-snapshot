const fs = require('fs')
const path = require('path')
const utils = require('@percy/sdk-utils')
// const { ROOT_PATH } = require('@sourcegraph/build-config')

const puppeteerPkg = require('puppeteer/package.json')
const CLIENT_INFO = 'percySnapshotToDisk/0.0.1'
const ENV_INFO = `${puppeteerPkg.name}/${puppeteerPkg.version}`

console.log('--------------------------------------------------------------------')
console.log('CWD in:', process.cwd())
console.log(JSON.stringify(process.env, null, 2))
const SNAPSHOT_DIR = path.join(process.cwd(), process.env.JS_BINARY__PACKAGE)
console.log('SNAPSHOT_DIR', SNAPSHOT_DIR)
const SNAPSHOT_FILE = path.join(SNAPSHOT_DIR, 'snapshots.snap')
console.log('SNAPSHOT_FILE', SNAPSHOT_FILE)
console.log('--------------------------------------------------------------------')
// shelljs.mkdir('-p', SNAPSHOT_DIR)

// Take a DOM snapshot and save it to disk
async function percySnapshotToDisk(page, name, options) {
  console.log('--------------------------------------------------------------------')
  //   const isPercyEnabled = await utils.isPercyEnabled()
  //   console.log(utils.isPercyEnabled)
  //   console.log('percySnapshotToDisk', page, name, isPercyEnabled)

  if (!page) throw new Error('A Puppeteer `page` object is required.')
  if (!name) throw new Error('The `name` argument is required.')
  //   if (!(await utils.isPercyEnabled())) return
  let log = utils.logger('puppeteer')
  console.log('PERCY ADDRESS', utils.percy.address)

  try {
    console.log('TRYING TO TAKE SNAPSHOT')
    // Inject the DOM serialization script
    const x = await fetch(`${utils.percy.address}/percy/healthcheck`)
    console.log('X', x)
    await page.evaluate(await utils.fetchPercyDOM())
    console.log('GOT PERCY DOM')

    // Serialize and capture the DOM
    let domSnapshot = await page.evaluate(options => {
      return window.PercyDOM.serialize(options)
    }, options)

    console.log('TRYING TO WRITE A FILE', SNAPSHOT_FILE)
    appendToFile(SNAPSHOT_FILE, {
      ...options,
      environmentInfo: ENV_INFO,
      clientInfo: CLIENT_INFO,
      url: page.url(),
      domSnapshot,
      name,
    })
  } catch (err) {
    console.log('FAILED TO TAKE SNAPSHOT', err)
    log.error(`Could not take DOM snapshot "${name}"`)
    log.error(err)
  }
  process.exit(1)
  console.log('--------------------------------------------------------------------')
}

function appendToFile(filePath, jsonObject) {
  const jsonString = JSON.stringify(jsonObject)
  const separator = fs.existsSync(filePath) ? '\n' : ''

  try {
    fs.appendFileSync(filePath, separator + jsonString)
    console.log(`JSON object appended to the file: ${filePath}`)
  } catch (err) {
    console.error(`Error appending JSON object to the file: ${err}`)
  }
}

module.exports = { percySnapshotToDisk }
