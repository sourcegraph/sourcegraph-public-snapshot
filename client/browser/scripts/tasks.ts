import fs from 'fs'
import { omit, pick } from 'lodash'
import path from 'path'
import shelljs from 'shelljs'
import signale from 'signale'
import utcVersion from 'utc-version'
import { Stats } from 'webpack'
import extensionInfo from '../src/extension/manifest.spec.json'
import schema from '../src/extension/schema.json'

export type BuildEnv = 'dev' | 'prod'

const BUILDS_DIR = 'build'

export const WEBPACK_STATS_OPTIONS = {
    all: false,
    timings: true,
    errors: true,
    warnings: true,
    colors: true,
} as Stats.ToStringOptions

function ensurePaths(): void {
    shelljs.mkdir('-p', 'build/dist')
    shelljs.mkdir('-p', 'build/bundles')
    shelljs.mkdir('-p', 'build/chrome')
    shelljs.mkdir('-p', 'build/firefox')
}

export function copyAssets(env: BuildEnv): void {
    signale.await('Copy assets')
    const dir = 'build/dist'
    shelljs.rm('-rf', dir)
    shelljs.mkdir('-p', dir)
    shelljs.cp('-R', 'src/extension/assets/*', dir)
    shelljs.cp('-R', 'src/extension/views/*', dir)
    signale.success('Assets copied')
}

function copyDist(toDir: string): void {
    shelljs.mkdir('-p', toDir)
    shelljs.cp('-R', 'build/dist/*', toDir)
}

export function copyPhabricator(): void {
    shelljs.mkdir('-p', 'build/phabricator/dist/scripts')
    shelljs.mkdir('-p', 'build/phabricator/dist/css')
    shelljs.cp('build/dist/js/phabricator.bundle.js', 'build/phabricator/dist/scripts')
    shelljs.cp('build/dist/css/style.bundle.css', 'build/phabricator/dist/css')
}

const browserTitles = {
    firefox: 'Firefox',
    chrome: 'Chrome',
}

const browserBundleZips = {
    firefox: 'firefox-bundle.xpi',
    chrome: 'chrome-bundle.zip',
}

const browserBlacklist = {
    chrome: ['applications'],
    firefox: ['key'],
}

const browserWhitelist = {}

function writeSchema(env, browser, writeDir): void {
    fs.writeFileSync(`${writeDir}/schema.json`, JSON.stringify(schema, null, 4))
}

const version = utcVersion()

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

    if (env === 'prod') {
        manifest.version = version
    }

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
            signale.await(`Building the ${title} ${env} bundle`)

            copyDist(buildDir)

            const zipDest = path.resolve(process.cwd(), `${BUILDS_DIR}/bundles/${browserBundleZips[browser]}`)
            if (zipDest) {
                shelljs.mkdir('-p', `./${BUILDS_DIR}/bundles`)
                shelljs.exec(`cd ${buildDir} && zip -q -r ${zipDest} *`)
            }

            signale.success(`Done building the ${title} ${env} bundle`)
        }
    }
}

export const buildFirefox = buildForBrowser('firefox')
export const buildChrome = buildForBrowser('chrome')
