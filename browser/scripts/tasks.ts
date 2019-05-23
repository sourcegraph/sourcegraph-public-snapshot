import fs from 'fs'
import { omit } from 'lodash'
import path from 'path'
import shelljs from 'shelljs'
import signale from 'signale'
import utcVersion from 'utc-version'
import { Stats } from 'webpack'
import extensionInfo from '../src/extension/manifest.spec.json'
import schema from '../src/extension/schema.json'

export type BuildEnv = 'dev' | 'prod'

type Browser = 'firefox' | 'chrome'

const BUILDS_DIR = 'build'

export const WEBPACK_STATS_OPTIONS: Stats.ToStringOptions = {
    all: false,
    timings: true,
    errors: true,
    warnings: true,
    colors: true,
}

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
    shelljs.cp('build/dist/js/extensionHostWorker.bundle.js', 'build/phabricator/dist/scripts')
    shelljs.cp('build/dist/css/style.bundle.css', 'build/phabricator/dist/css')
}

const BROWSER_TITLES = {
    firefox: 'Firefox',
    chrome: 'Chrome',
}

const BROWSER_BUNDLE_ZIPS = {
    firefox: 'firefox-bundle.xpi',
    chrome: 'chrome-bundle.zip',
}

const BROWSER_BLACKLIST = {
    chrome: ['applications'],
    firefox: ['key'],
}

function writeSchema(env: BuildEnv, browser: Browser, writeDir: string): void {
    fs.writeFileSync(`${writeDir}/schema.json`, JSON.stringify(schema, null, 4))
}

const version = utcVersion()

function writeManifest(env: BuildEnv, browser: Browser, writeDir: string): void {
    const manifest = {
        ...omit(extensionInfo, ['dev', 'prod', ...BROWSER_BLACKLIST[browser]]),
        ...omit(extensionInfo[env], BROWSER_BLACKLIST[browser]),
    }

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

function buildForBrowser(browser: Browser): (env: BuildEnv) => () => void {
    ensurePaths()
    return env => {
        const title = BROWSER_TITLES[browser]

        const buildDir = path.resolve(process.cwd(), `${BUILDS_DIR}/${browser}`)

        writeManifest(env, browser, buildDir)
        writeSchema(env, browser, buildDir)

        return () => {
            // Allow only building for specific browser targets. Useful in local dev for faster
            // builds.
            if (process.env.TARGETS && !process.env.TARGETS.includes(browser)) {
                return
            }

            signale.await(`Building the ${title} ${env} bundle`)

            copyDist(buildDir)

            const zipDest = path.resolve(process.cwd(), `${BUILDS_DIR}/bundles/${BROWSER_BUNDLE_ZIPS[browser]}`)
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
