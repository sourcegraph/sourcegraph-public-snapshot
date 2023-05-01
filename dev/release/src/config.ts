import { readFileSync, writeFileSync } from 'fs'

import chalk from 'chalk'
import { parse as parseJSONC } from 'jsonc-parser'
import { DateTime } from 'luxon'
import * as semver from 'semver'
import { SemVer } from 'semver'

import { getPreviousVersion } from './git'
import { retryInput } from './util'

const releaseConfigPath = 'release-config.jsonc'

/**
 * Release configuration file format
 */
export interface Config {
    teamEmail: string

    captainSlackUsername: string
    captainGitHubUsername: string

    previousRelease: string
    upcomingRelease: string

    oneWorkingWeekBeforeRelease: string
    threeWorkingDaysBeforeRelease: string
    releaseDate: string
    oneWorkingDayAfterRelease: string
    oneWorkingWeekAfterRelease: string

    slackAnnounceChannel: string

    dryRun: {
        tags?: boolean
        changesets?: boolean
        trackingIssues?: boolean
        slack?: boolean
        calendar?: boolean
    }
}

export async function getActiveRelease(config: ReleaseConfig): Promise<ActiveRelease> {
    if (!config.in_progress || config.in_progress.releases.length === 0) {
        console.log(chalk.yellow('No active releases are defined! Attempting to activate...'))
        await activateRelease(config)
    }
    if (!config.in_progress) {
        throw new Error('unable to activate a release!')
    }
    if (config.in_progress.releases.length > 1) {
        throw new Error(
            chalk.red(
                'The release config has multiple versions activated. This feature is not yet supported by the release tool! Please activate only a single release.'
            )
        )
    }
    const rel = config.in_progress.releases[0]
    const def = config.scheduledReleases[rel.version]
    const version = new SemVer(rel.version)
    return {
        version,
        previous: new SemVer(rel.previous),
        ...(def as ReleaseDates),
        ...(def as ReleaseCaptainInformation),
        branch: `${version.major}.${version.minor}`,
        srcCliVersion: config.in_progress.srcCliVersion ? new SemVer(config.in_progress.srcCliVersion) : undefined,
    }
}

export function loadReleaseConfig(): ReleaseConfig {
    return parseJSONC(readFileSync(releaseConfigPath).toString()) as ReleaseConfig
}

export function saveReleaseConfig(config: ReleaseConfig): void {
    writeFileSync(releaseConfigPath, JSON.stringify(config, null, 2))
}

export function newRelease(
    version: SemVer,
    releaseDate: DateTime,
    captainGithub: string,
    captainSlack: string
): ScheduledReleaseDefinition {
    return {
        ...releaseDates(releaseDate, version.patch === 0),
        current: version.version,
        captainGitHubUsername: captainGithub,
        captainSlackUsername: captainSlack,
    }
}

export async function newReleaseFromInput(versionOverride?: SemVer): Promise<ScheduledReleaseDefinition> {
    let version = versionOverride
    if (!version) {
        version = await selectVersionWithSuggestion('Enter the desired version number')
    }

    const releaseDateStr = await retryInput(
        'Enter the release date (YYYY-MM-DD). Enter blank to use current date: ',
        val => {
            if (val && /^\d{4}-\d{2}-\d{2}$/.test(val)) {
                return true
            }
            // this will return false if the input doesn't match the regexp above but does exist, allowing blank input to still be valid
            return !val
        },
        'invalid date, expected format YYYY-MM-DD'
    )
    let releaseTime: DateTime
    if (!releaseDateStr) {
        releaseTime = DateTime.now().setZone('America/Los_Angeles')
        console.log(chalk.blue(`Using current time: ${releaseTime.toString()}`))
    } else {
        releaseTime = DateTime.fromISO(releaseDateStr, { zone: 'America/Los_Angeles' })
    }

    const captainGithubUsername = await retryInput('Enter the github username of the release captain: ', val => !!val)
    const captainSlackUsername = await retryInput('Enter the slack username of the release captain: ', val => !!val)

    const rel = newRelease(version, releaseTime, captainGithubUsername, captainSlackUsername)
    console.log(chalk.green('Version created:'))
    console.log(chalk.green(JSON.stringify(rel, null, 2)))
    return rel
}

function releaseDates(releaseDate: DateTime, includePatches?: boolean): ReleaseDates {
    releaseDate = releaseDate.set({ hour: 10 })
    return {
        codeFreezeDate: releaseDate.plus({ days: -7 }).toString(),
        securityApprovalDate: releaseDate.plus({ days: -7 }).toString(),
        releaseDate: releaseDate.toString(),
        patches: includePatches
            ? generatePatchDates(releaseDate, releaseDate.plus({ months: 3 }), 2).map(rdate => rdate.toString())
            : undefined,
    }
}

function generatePatchDates(start: DateTime, end: DateTime, intervalWeeks: number): DateTime[] {
    const patches = []
    let current: DateTime = start.plus({ weeks: intervalWeeks })
    while (current < end.minus({ weeks: 1 })) {
        patches.push(current)
        current = current.plus({ weeks: intervalWeeks })
    }
    return patches
}

export function addScheduledRelease(config: ReleaseConfig, release: ScheduledReleaseDefinition): ReleaseConfig {
    config.scheduledReleases[release.current] = release
    return config
}

export function removeScheduledRelease(config: ReleaseConfig, version: string): ReleaseConfig {
    delete config.scheduledReleases[version]
    return config
}

export interface ReleaseDates {
    releaseDate: string
    codeFreezeDate: string
    securityApprovalDate: string
    patches?: string[]
}

export interface ActiveRelease extends ReleaseCaptainInformation, ReleaseDates {
    version: SemVer
    previous: SemVer
    branch: string
    srcCliVersion?: SemVer
}

export interface ActiveReleaseDefinition {
    version: string
    previous: string
}

export interface ReleaseCaptainInformation {
    captainSlackUsername: string
    captainGitHubUsername: string
}

export interface InProgress extends ReleaseCaptainInformation {
    releases: ActiveReleaseDefinition[]
    srcCliVersion?: string
    googleExecutorVersion?: string
    awsExecutorVersion?: string
}

export interface ReleaseConfig {
    metadata: {
        teamEmail: string
        slackAnnounceChannel: string
    }
    scheduledReleases: {
        [version: string]: ScheduledReleaseDefinition
    }
    in_progress?: InProgress
    dryRun: {
        tags?: boolean
        changesets?: boolean
        trackingIssues?: boolean
        slack?: boolean
        calendar?: boolean
    }
}

export interface ScheduledReleaseDefinition extends ReleaseDates, ReleaseCaptainInformation {
    current: string
}

// Prompt a user for input and activate the given release version if possible. Will redirect to release creation input if
// the version isn't defined.
export async function activateRelease(config: ReleaseConfig): Promise<void> {
    const next = await selectVersionWithSuggestion('Enter the version to activate')
    console.log('Attempting to detect previous version...')
    const previous = getPreviousVersion(next)
    console.log(chalk.blue(`Detected previous version: ${previous.version}`))

    const scheduled = await getScheduledReleaseWithInput(config, next)
    config.in_progress = {
        captainGitHubUsername: scheduled.captainGitHubUsername,
        captainSlackUsername: scheduled.captainSlackUsername,
        releases: [{ version: next.version, previous: previous.version }],
    }
    saveReleaseConfig(config)
    console.log(chalk.green(`Release: ${next.version} activated!`))
}

export function deactivateAllReleases(config: ReleaseConfig): void {
    delete config.in_progress
    saveReleaseConfig(config)
}

// Prompt a user for a major / minor version input with automation suggestion by adding a minor version to the previous version.
async function selectVersionWithSuggestion(prompt: string): Promise<SemVer> {
    const probablyMinor = getPreviousVersion().inc('minor')
    const probablyPatch = getPreviousVersion().inc('patch')
    const input = await retryInput(
        `Next minor release: ${probablyMinor.version}\nNext patch release: ${probablyPatch.version}\n${chalk.blue(
            prompt
        )}: `,
        val => {
            const version = semver.parse(val)
            return !!version
        }
    )
    return new SemVer(input)
}

// Prompt a user for a release definition input, and redirect to creation input if it doesn't exist.
export async function getReleaseDefinition(config: ReleaseConfig): Promise<ScheduledReleaseDefinition> {
    const next = await selectVersionWithSuggestion('Enter the version number to select')
    return getScheduledReleaseWithInput(config, next)
}

// Helper function to get a release definition from the release config, redirecting to creation input if it doesn't exist.
async function getScheduledReleaseWithInput(
    config: ReleaseConfig,
    releaseVersion: SemVer
): Promise<ScheduledReleaseDefinition> {
    let scheduled = config.scheduledReleases[releaseVersion.version]
    if (!scheduled) {
        console.log(
            chalk.yellow(`Release definition not found for: ${releaseVersion.version}, enter release information.\n`)
        )
        scheduled = await newReleaseFromInput(releaseVersion)
        addScheduledRelease(config, scheduled)
        saveReleaseConfig(config)
    }
    return scheduled
}

export function setSrcCliVersion(config: ReleaseConfig, version: string): void {
    if (config.in_progress) {
        config.in_progress.srcCliVersion = version
    }
    saveReleaseConfig(config)
}

export function setGoogleExecutorVersion(config: ReleaseConfig, version: string): void {
    if (config.in_progress) {
        config.in_progress.googleExecutorVersion = version
    }
    saveReleaseConfig(config)
}

export function setAWSExecutorVersion(config: ReleaseConfig, version: string): void {
    if (config.in_progress) {
        config.in_progress.awsExecutorVersion = version
    }
    saveReleaseConfig(config)
}
