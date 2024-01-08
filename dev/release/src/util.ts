import { readdirSync, readFileSync, writeFileSync } from 'fs'
import * as path from 'path'
import * as readline from 'readline'

import type Octokit from '@octokit/rest'
import chalk from 'chalk'
import execa from 'execa'
import { mkdir, readFile, writeFile } from 'mz/fs'
import fetch from 'node-fetch'
import * as semver from 'semver'
import { SemVer } from 'semver'

import type { ReleaseConfig } from './config'
import { getPreviousVersionExecutor, getPreviousVersionSrcCli } from './git'
import { cloneRepo, type EditFunc, getAuthenticatedGitHubClient, listIssues } from './github'
import * as update from './update'

const SOURCEGRAPH_RELEASE_INSTANCE_URL = 'https://sourcegraph.sourcegraph.com'

export interface ReleaseTag {
    repo: string
    nextTag: string
    workDir: string
}

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

export async function verifyWithInput(prompt: string): Promise<void> {
    await readLineNoCache(chalk.yellow(`${prompt}\nInput yes to confirm: `)).then(val => {
        if (!(val === 'yes' || val === 'y')) {
            console.log(chalk.red('Aborting!'))
            process.exit(0)
        }
    })
}

// similar to verifyWithInput but will not exit and will allow the caller to decide what to do
export async function softVerifyWithInput(prompt: string): Promise<boolean> {
    return readLineNoCache(chalk.yellow(`${prompt}\nInput yes to confirm: `)).then(val => val === 'yes' || val === 'y')
}

export async function ensureDocker(): Promise<execa.ExecaReturnValue<string>> {
    return execa('docker', ['version'], { stdout: 'ignore' })
}

export function changelogURL(version: string): string {
    const versionAnchor = version.replaceAll('.', '-')
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

export async function getLatestSrcCliGithubRelease(): Promise<string> {
    return fetch('https://api.github.com/repos/sourcegraph/src-cli/releases/latest', {
        method: 'GET',
        headers: {
            Accept: 'application/json',
        },
    })
        .then(response => response.json())
        .then(json => json.tag_name)
}

export async function ensureSrcCliUpToDate(): Promise<void> {
    const latestTag = await getLatestSrcCliGithubRelease()
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
    return fetch(`https://api.github.com/repos/${owner}/${repo}/tags`, {
        method: 'GET',
        headers: {
            Accept: 'application/json',
        },
    })
        .then(response => response.json())
        .then(json => json[0].name)
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
    const registryPassword = await readLine(
        `Enter your container registry (${registryHostname} ) access token: `,
        `${cacheFolder}/cr_${registryHostname.replace('.', '_')}_password.txt`
    )
    const credential: ContainerRegistryCredential = {
        username: registryUsername,
        password: registryPassword,
        hostname: registryHostname,
    }
    return credential
}

export type ContentFunc = (previousVersion?: string, nextVersion?: string) => string

const upgradeContentGenerators: { [s: string]: ContentFunc } = {
    docker_compose: () => '',
    kubernetes: () => '',
    server: () => '',
    pure_docker: (previousVersion?: string, nextVersion?: string) => {
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
                continue
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

export async function retryInput(
    prompt: string,
    delegate: (val: string) => boolean,
    errorMessage?: string
): Promise<string> {
    while (true) {
        const val = await readLine(prompt).then(value => value)
        if (delegate(val)) {
            return val
        }
        if (errorMessage) {
            console.log(chalk.red(errorMessage))
        } else {
            console.log(chalk.red('invalid input'))
        }
    }
}

const blockingQuery = 'is:open org:sourcegraph label:release-blocker'

export async function getReleaseBlockers(
    octokit: Octokit
): Promise<Octokit.SearchIssuesAndPullRequestsResponseItemsItem[]> {
    return listIssues(octokit, blockingQuery)
}

export function backportIssueQuery(version: SemVer): string {
    return `is:open is:pr repo:sourcegraph org:sourcegraph label:"backported-to-${version.major}.${version.minor}"`
}

export async function getBackportsForVersion(
    octokit: Octokit,
    version: SemVer
): Promise<Octokit.SearchIssuesAndPullRequestsResponseItemsItem[]> {
    return listIssues(octokit, backportIssueQuery(version))
}

export function releaseBlockerUri(): string {
    return issuesQueryUri(blockingQuery)
}

function issuesQueryUri(query: string): string {
    return `https://github.com/issues?q=${encodeURIComponent(query)}`
}

export async function validateNoOpenBackports(octokit: Octokit, version: SemVer): Promise<void> {
    const backports = await getBackportsForVersion(octokit, version)
    if (backports.length > 0) {
        await verifyWithInput(`${backportWarning(backports.length, version)})\nConfirm to proceed`)
    } else {
        console.log('No backports found!')
    }
}

export async function backportStatus(octokit: Octokit, version: SemVer): Promise<string> {
    const backports = await getBackportsForVersion(octokit, version)
    return backportWarning(backports.length, version)
}

export function backportWarning(numBackports: number, version: SemVer): string {
    return `Warning! There are ${chalk.red(numBackports)} backport pull requests open!\n${issuesQueryUri(
        backportIssueQuery(version)
    )}`
}

export async function validateNoReleaseBlockers(octokit: Octokit): Promise<void> {
    const blockers = await getReleaseBlockers(octokit)
    if (blockers.length > 0) {
        await verifyWithInput(
            `Warning! There are ${chalk.red(
                blockers.length
            )} release blocking issues open!\n${releaseBlockerUri()}\nConfirm to proceed`
        )
    }
}

export async function nextSrcCliVersionInputWithAutodetect(config: ReleaseConfig, repoPath?: string): Promise<SemVer> {
    let next: SemVer
    if (!config.in_progress?.srcCliVersion) {
        if (!repoPath) {
            const client = await getAuthenticatedGitHubClient()
            const { workdir } = await cloneRepo(client, 'sourcegraph', 'src-cli', {
                revision: 'main',
                revisionMustExist: true,
            })
            repoPath = workdir
        }
        console.log('Attempting to detect previous src-cli version...')
        const previous = getPreviousVersionSrcCli(repoPath)
        console.log(chalk.blue(`Detected previous src-cli version: ${previous.version}`))
        next = previous.inc('minor')
    } else {
        next = new SemVer(config.in_progress.srcCliVersion)
    }

    if (!(await softVerifyWithInput(`Confirm next version of src-cli should be: ${next.version}`))) {
        return new SemVer(
            await retryInput(
                'Enter the next version of src-cli: ',
                val => !!semver.parse(val),
                'Expected semver format'
            )
        )
    }
    return next
}

export async function nextGoogleExecutorVersionInputWithAutodetect(
    config: ReleaseConfig,
    repoPath?: string
): Promise<SemVer> {
    let next: SemVer
    if (!config.in_progress?.googleExecutorVersion) {
        if (!repoPath) {
            const client = await getAuthenticatedGitHubClient()
            const { workdir } = await cloneRepo(client, 'sourcegraph', 'terraform-google-executors', {
                revision: 'main',
                revisionMustExist: true,
            })
            repoPath = workdir
        }
        console.log('Attempting to detect previous executor version...')
        const previous = getPreviousVersionExecutor(repoPath)
        console.log(chalk.blue(`Detected previous executor version: ${previous.version}`))
        next = previous.inc('minor')
    } else {
        next = new SemVer(config.in_progress.googleExecutorVersion)
    }

    if (
        !(await softVerifyWithInput(
            `Confirm next version of sourcegraph/terraform-google-executors should be: ${next.version}`
        ))
    ) {
        return new SemVer(
            await retryInput(
                'Enter the next version of executor: ',
                val => !!semver.parse(val),
                'Expected semver format'
            )
        )
    }
    return next
}

export async function nextAWSExecutorVersionInputWithAutodetect(
    config: ReleaseConfig,
    repoPath?: string
): Promise<SemVer> {
    let next: SemVer
    if (!config.in_progress?.awsExecutorVersion) {
        if (!repoPath) {
            const client = await getAuthenticatedGitHubClient()
            const { workdir } = await cloneRepo(client, 'sourcegraph', 'terraform-aws-executors', {
                revision: 'main',
                revisionMustExist: true,
            })
            repoPath = workdir
        }
        console.log('Attempting to detect previous executor version...')
        const previous = getPreviousVersionExecutor(repoPath)
        console.log(chalk.blue(`Detected previous sourcegraph/terraform-aws-executors version: ${previous.version}`))
        next = previous.inc('minor')
    } else {
        next = new SemVer(config.in_progress.awsExecutorVersion)
    }

    if (!(await softVerifyWithInput(`Confirm next version of executor should be: ${next.version}`))) {
        return new SemVer(
            await retryInput(
                'Enter the next version of executor: ',
                val => !!semver.parse(val),
                'Expected semver format'
            )
        )
    }
    return next
}

export function pullRequestBody(content: string): string {
    const header = 'This pull request was automatically generated by the release-tool.\n'
    const testPlan = '\n## Test Plan:\nN/A'
    return `${header}${content}${testPlan}`
}
