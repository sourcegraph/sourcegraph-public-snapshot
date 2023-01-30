import { readdirSync, readFileSync, writeFileSync } from 'fs'
import * as path from 'path'
import * as readline from 'readline'

import execa from 'execa'
import { readFile, writeFile, mkdir } from 'mz/fs'
import fetch from 'node-fetch'

import { EditFunc } from './github'
import * as update from './update'

const SOURCEGRAPH_RELEASE_INSTANCE_URL = 'https://k8s.sgdev.org'

/* eslint-disable @typescript-eslint/consistent-type-assertions */
export function formatDate(date: Date): string {
    return `${date.toLocaleString('en-US', {
        timeZone: 'UTC',
        dateStyle: 'medium',
        timeStyle: 'short',
    } as Intl.DateTimeFormatOptions)} (UTC)`
}
/* eslint-enable @typescript-eslint/consistent-type-assertions */

const addZero = (index: number): string => (index < 10 ? `0${index}` : `${index}`)

/**
 * Generates a link for comparing given Date with local time.
 */
export function timezoneLink(date: Date, linkName: string): string {
    const timeString = `${addZero(date.getUTCHours())}${addZero(date.getUTCMinutes())}`
    return `https://time.is/${timeString}_${date.getUTCDate()}_${date.toLocaleString('en-US', {
        month: 'short',
    })}_${date.getUTCFullYear()}_in_UTC?${encodeURI(linkName)}`
}

export const cacheFolder = './.secrets'

export async function readLine(prompt: string, cacheFile?: string): Promise<string> {
    if (!cacheFile) {
        return readLineNoCache(prompt)
    }

    try {
        return (await readFile(cacheFile, { encoding: 'utf8' })).trimEnd()
    } catch {
        const userInput = await readLineNoCache(prompt)
        await mkdir(path.dirname(cacheFile), { recursive: true })
        await writeFile(cacheFile, userInput)
        return userInput
    }
}

async function readLineNoCache(prompt: string): Promise<string> {
    const readlineInterface = readline.createInterface({
        input: process.stdin,
        output: process.stdout,
    })
    const userInput = await new Promise<string>(resolve => readlineInterface.question(prompt, resolve))
    readlineInterface.close()
    return userInput
}

export function getWeekNumber(date: Date): number {
    const firstJan = new Date(date.getFullYear(), 0, 1)
    const day = 86400000
    return Math.ceil(((date.valueOf() - firstJan.valueOf()) / day + firstJan.getDay() + 1) / 7)
}

export async function ensureDocker(): Promise<execa.ExecaReturnValue<string>> {
    return execa('docker', ['version'], { stdout: 'ignore' })
}

export function changelogURL(version: string): string {
    const versionAnchor = version.replace(/\./g, '-')
    return `https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/CHANGELOG.md#${versionAnchor}`
}

function ensureBranchUpToDate(baseBranch: string, targetBranch: string): boolean {
    const [behind, ahead] = execa
        .sync('git', ['rev-list', '--left-right', '--count', targetBranch + '...' + baseBranch])
        .stdout.split('\t')

    if (behind === '0' && ahead === '0') {
        return true
    }

    const countCommits = function (numberOfCommits: string, aheadOrBehind: string): string {
        return numberOfCommits === '1'
            ? numberOfCommits + ' commit ' + aheadOrBehind
            : numberOfCommits + ' commits ' + aheadOrBehind
    }

    if (behind !== '0' && ahead !== '0') {
        console.log(
            `Your branch is ${countCommits(ahead, 'ahead')} and ${countCommits(
                behind,
                'behind'
            )} the branch ${targetBranch}.`
        )
    } else if (behind !== '0') {
        console.log(`Your branch is ${countCommits(behind, 'behind')} the branch ${targetBranch}.`)
    } else if (ahead !== '0') {
        console.log(`Your branch is ${countCommits(ahead, 'ahead')} the branch ${targetBranch}.`)
    }

    return false
}

export function ensureMainBranchUpToDate(): void {
    const mainBranch = 'main'
    const remoteMainBranch = 'origin/main'
    const currentBranch = execa.sync('git', ['rev-parse', '--abbrev-ref', 'HEAD']).stdout.trim()
    if (currentBranch !== mainBranch) {
        console.log(
            `Expected to be on branch ${mainBranch}, but was on ${currentBranch}. Run \`git checkout ${mainBranch}\` to switch to the main branch.`
        )
        process.exit(1)
    }
    execa.sync('git', ['remote', 'update'], { stdout: 'ignore' })
    if (!ensureBranchUpToDate(mainBranch, remoteMainBranch)) {
        process.exit(1)
    }
}

export function ensureReleaseBranchUpToDate(branch: string): void {
    const remoteBranch = 'origin/' + branch
    if (!ensureBranchUpToDate(branch, remoteBranch)) {
        process.exit(1)
    }
}

export async function ensureSrcCliUpToDate(): Promise<void> {
    const latestTag = await fetch('https://api.github.com/repos/sourcegraph/src-cli/releases/latest', {
        method: 'GET',
        headers: {
            Accept: 'application/json',
        },
    })
        .then(response => response.json())
        .then(json => json.tag_name)

    let installedTag = execa.sync('src', ['version']).stdout.split('\n')
    installedTag = installedTag[0].split(':')
    const trimmedInstalledTag = installedTag[1].trim()

    if (trimmedInstalledTag !== latestTag) {
        try {
            console.log('Uprading src-cli to the latest version.')
            execa.sync('brew', ['upgrade', 'src-cli'])
        } catch (error) {
            console.log('Trouble upgrading src-cli:', error)
            process.exit(1)
        }
    }
}

export function ensureSrcCliEndpoint(): void {
    const srcEndpoint = process.env.SRC_ENDPOINT
    if (srcEndpoint !== SOURCEGRAPH_RELEASE_INSTANCE_URL) {
        throw new Error(`the $SRC_ENDPOINT provided doesn't match what is expected by the release tool.
Expected $SRC_ENDPOINT to be "${SOURCEGRAPH_RELEASE_INSTANCE_URL}"`)
    }
}

export async function getLatestTag(owner: string, repo: string): Promise<string> {
    const latestTag = await fetch(`https://api.github.com/repos/${owner}/${repo}/tags`, {
        method: 'GET',
        headers: {
            Accept: 'application/json',
        },
    })
        .then(response => response.json())
        .then(json => json[0].name)
    return latestTag
}

interface ContainerRegistryCredential {
    username: string
    password: string
    hostname: string
}

export async function getContainerRegistryCredential(registryHostname: string): Promise<ContainerRegistryCredential> {
    const registryUsername = await readLine(
        `Enter your container registry (${registryHostname} ) username: `,
        `${cacheFolder}/cr_${registryHostname.replace('.', '_')}_username.txt`
    )
    const registryPassowrd = await readLine(
        `Enter your container registry (${registryHostname} ) password or access token: `,
        `${cacheFolder}/cr_${registryHostname.replace('.', '_')}_password.txt`
    )
    const credential: ContainerRegistryCredential = {
        username: registryUsername,
        password: registryPassowrd,
        hostname: registryHostname,
    }
    return credential
}

export type ContentFunc = (previousVersion: string, nextVersion: string) => string

const upgradeContentGenerators: { [s: string]: ContentFunc } = {
    docker_compose: (previousVersion: string, nextVersion: string) => '',
    kubernetes: (previousVersion: string, nextVersion: string) => '',
    server: (previousVersion: string, nextVersion: string) => '',
    pure_docker: (previousVersion: string, nextVersion: string) => {
        const compare = `compare/v${previousVersion}...v${nextVersion}`
        return `As a template, perform the same actions as the following diff in your own deployment: [\`Upgrade to v${nextVersion}\`](https://github.com/sourcegraph/deploy-sourcegraph-docker/${compare})
\nFor non-standard replica builds: 
- [\`Customer Replica 1: ➔ v${nextVersion}\`](https://github.com/sourcegraph/deploy-sourcegraph-docker-customer-replica-1/${compare})`
    },
}
export const getUpgradeGuide = (mode: string): ContentFunc => upgradeContentGenerators[mode]

export const getAllUpgradeGuides = (previous: string, next: string): string[] =>
    Object.keys(upgradeContentGenerators).map(
        key => `Guide for: ${key}\n\n${upgradeContentGenerators[key](previous, next)}`
    )

export const updateUpgradeGuides = (previous: string, next: string): EditFunc => {
    let updateDirectory = '/doc/admin/updates'
    const notPatchRelease = next.endsWith('.0')

    return (directory: string): void => {
        updateDirectory = directory + updateDirectory
        for (const file of readdirSync(updateDirectory)) {
            if (file === 'index.md') {
                continue
            }
            const mode = file.replace('.md', '')
            const updateFunc = getUpgradeGuide(mode)
            if (updateFunc === undefined) {
                console.log(`Skipping upgrade file: ${file} due to missing content generator`)
            }
            const guide = getUpgradeGuide(mode)(previous, next)

            const fullPath = path.join(updateDirectory, file)
            console.log(`Updating upgrade guide: ${fullPath}`)
            let updateContents = readFileSync(fullPath).toString()
            const releaseHeader = `## v${previous} ➔ v${next}`
            const notesHeader = '\n\n#### Notes:'

            if (notPatchRelease) {
                let content = `${update.releaseTemplate}\n\n${releaseHeader}`
                if (guide) {
                    content = `${content}\n\n${guide}`
                }
                content = content + notesHeader
                updateContents = updateContents.replace(update.releaseTemplate, content)
            } else {
                const prevReleaseHeaderPattern = `##\\s+v\\d\\.\\d(?:\\.\\d)? ➔ v${previous}\\s*`
                const matches = updateContents.match(new RegExp(prevReleaseHeaderPattern))
                if (!matches || matches.length === 0) {
                    console.log(`Unable to find header using pattern: ${prevReleaseHeaderPattern}. Skipping.`)
                    continue
                }
                const prevReleaseHeader = matches[0]
                let content = `${releaseHeader}`
                if (guide) {
                    content = `${content}\n\n${guide}`
                }
                content = content + notesHeader + `\n\n${prevReleaseHeader}`
                updateContents = updateContents.replace(prevReleaseHeader, content)
            }
            writeFileSync(fullPath, updateContents)
        }
    }
}
