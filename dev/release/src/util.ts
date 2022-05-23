import * as path from 'path'
import * as readline from 'readline'
import { URL } from 'url'

import execa from 'execa'
import { readFile, writeFile, mkdir } from 'mz/fs'

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

export function hubSpotFeedbackFormStub(version: string): string {
    const link = `[this feedback form](${hubSpotFeedbackFormURL(version)})`
    return `*How smooth was this upgrade process for you? You can give us your feedback on this upgrade by filling out ${link}.*`
}

function hubSpotFeedbackFormURL(version: string): string {
    const url = new URL('https://share.hsforms.com/1aGeG7ALQQEGO6zyfauIiCA1n7ku')
    url.searchParams.set('update_version', version)

    return url.toString()
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
