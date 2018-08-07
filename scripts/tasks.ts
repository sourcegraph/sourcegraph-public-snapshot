import fs from 'fs'
import _ from 'lodash'
import path from 'path'
import shelljs from 'shelljs'
import extensionInfo from '../chrome/extension.info.json'
import schema from '../chrome/schema.json'

export type BuildEnv = 'dev' | 'prod'

const omit = _.omit
const pick = _.pick

const BUILDS_DIR = 'build'

function ensurePaths(): void {
    shelljs.mkdir('-p', 'build/dist')
    shelljs.mkdir('-p', 'build/bundles')
    shelljs.mkdir('-p', 'build/chrome')
    shelljs.mkdir('-p', 'build/firefox')
}

export function copyAssets(env: BuildEnv): void {
    const dir = 'build/dist'
    shelljs.rm('-rf', dir)
    shelljs.mkdir('-p', dir)
    shelljs.cp('-R', 'chrome/assets/*', dir)
    shelljs.cp('-R', 'chrome/views/*', dir)
}

function copyDist(toDir: string): void {
    shelljs.mkdir('-p', toDir)
    shelljs.cp('-R', 'build/dist/*', toDir)
}

const browserTitles = {
    firefox: 'FireFox',
    chrome: 'Chrome',
}

const browserBundleZips = {
    firefox: 'firefox-bundle.xpi',
    chrome: 'chrome-bundle.zip',
}

const browserBlacklist = {
    chrome: ['applications'],
    firefox: ['key'],
    safari: [],
}

const browserWhitelist = {
    safari: ['version'],
}

function writeSchema(env, browser, writeDir): void {
    fs.writeFileSync(`${writeDir}/schema.json`, JSON.stringify(schema, null, 4))
}

function writeManifest(env, browser, writeDir): void {
    let envInfo = omit(extensionInfo[env], browserBlacklist[browser])

    let manifest
    const whitelist = browserWhitelist[browser]
    if (whitelist) {
        manifest = pick(extensionInfo, whitelist)
        envInfo = pick(envInfo, whitelist)
    } else {
        manifest = omit(extensionInfo, ['dev', 'prod', ...browserBlacklist[browser]])
    }

    manifest = { ...manifest, ...envInfo }

    if (browser === 'firefox') {
        manifest.permissions.push('<all_urls>')
        delete manifest.storage
    }

    delete manifest.$schema

    fs.writeFileSync(`${writeDir}/manifest.json`, JSON.stringify(manifest, null, 4))
}

function buildForBrowser(browser): (env: string) => () => void {
    ensurePaths()
    return env => {
        const title = browserTitles[browser]

        const buildDir = path.resolve(process.cwd(), `${BUILDS_DIR}/${browser}`)

        writeManifest(env, browser, buildDir)
        writeSchema(env, browser, buildDir)

        return () => {
            console.log(`Building ${title} ${env} bundle...`)

            copyDist(buildDir)

            const zipDest = path.resolve(process.cwd(), `${BUILDS_DIR}/bundles/${browserBundleZips[browser]}`)
            if (zipDest) {
                shelljs.mkdir('-p', `./${BUILDS_DIR}/bundles`)
                shelljs.exec(`cd ${buildDir} && zip -r ${zipDest} *`)
            }

            console.log(`${title} ${env} bundle built.`)
        }
    }
}

export const buildFirefox = buildForBrowser('firefox')
export const buildChrome = buildForBrowser('chrome')

export function buildSafari(env: BuildEnv): void {
    console.log(`Building Safari ${env} bundle...`)

    shelljs.exec('cp -r build/dist/* Sourcegraph.safariextension')
    writeManifest(env, 'safari', 'Sourcegraph.safariextension')

    console.log(`Safari ${env} bundle built.`)
}
