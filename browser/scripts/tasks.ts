/* eslint no-sync: warn */
import fs from 'fs'
import { omit } from 'lodash'
import path from 'path'
import shelljs from 'shelljs'
import signale from 'signale'
import utcVersion from 'utc-version'
import { Stats } from 'webpack'
import extensionInfo from '../src/extension/manifest.spec.json'
import schema from '../src/extension/schema.json'

/**
 * If true, add <all_urls> to the permissions in the manifest.
 * This is needed for e2e tests because it is not possible to accept the permission prompt with puppeteer.
 */
const EXTENSION_PERMISSIONS_ALL_URLS = Boolean(
    process.env.EXTENSION_PERMISSIONS_ALL_URLS && JSON.parse(process.env.EXTENSION_PERMISSIONS_ALL_URLS)
)

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

export function copyAssets(): void {
    signale.await('Copy assets')
    const dir = 'build/dist'
    shelljs.rm('-rf', dir)
    shelljs.mkdir('-p', dir)
    shelljs.cp('-R', 'src/extension/assets/*', dir)
    shelljs.cp('-R', 'src/extension/views/*', dir)
    signale.success('Assets copied')
}

function copyExtensionAssets(toDir: string): void {
    shelljs.mkdir('-p', `${toDir}/js`, `${toDir}/css`, `${toDir}/img`)
    shelljs.cp('build/dist/js/background.bundle.js', `${toDir}/js`)
    shelljs.cp('build/dist/js/inject.bundle.js', `${toDir}/js`)
    shelljs.cp('build/dist/js/options.bundle.js', `${toDir}/js`)
    shelljs.cp('build/dist/css/style.bundle.css', `${toDir}/css`)
    shelljs.cp('build/dist/css/options-style.bundle.css', `${toDir}/css`)
    shelljs.cp('build/dist/css/options-style.bundle.css', `${toDir}/css`)
    shelljs.cp('-R', 'build/dist/img/*', `${toDir}/img`)
    shelljs.cp('build/dist/background.html', toDir)
    shelljs.cp('build/dist/options.html', toDir)
}

export function copyIntegrationAssets(): void {
    shelljs.mkdir('-p', 'build/integration/scripts')
    shelljs.mkdir('-p', 'build/integration/css')
    shelljs.cp('build/dist/js/phabricator.bundle.js', 'build/integration/scripts')
    shelljs.cp('build/dist/js/integration.bundle.js', 'build/integration/scripts')
    shelljs.cp('build/dist/js/extensionHostWorker.bundle.js', 'build/integration/scripts')
    shelljs.cp('build/dist/css/style.bundle.css', 'build/integration/css')
    shelljs.cp('src/phabricator/extensionHostFrame.html', 'build/integration')
    // Copy to the ui/assets directory so that these files can be served by the webapp.
    shelljs.mkdir('-p', '../ui/assets/extension')
    shelljs.cp('-r', 'build/integration/*', '../ui/assets/extension')
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
    chrome: ['applications'] as const,
    firefox: ['key'] as const,
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

    if (EXTENSION_PERMISSIONS_ALL_URLS) {
        manifest.permissions!.push('<all_urls>')
        signale.info('Adding <all_urls> to permissions because of env var setting')
    }

    if (browser === 'firefox') {
        manifest.permissions!.push('<all_urls>')
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
            // Allow only building for specific browser targets.
            // Useful in local dev for faster builds.
            if (process.env.TARGETS && !process.env.TARGETS.includes(browser)) {
                return
            }

            signale.await(`Building the ${title} ${env} bundle`)

            copyExtensionAssets(buildDir)

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
