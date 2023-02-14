import {readFileSync, unlinkSync, writeFileSync} from 'fs'

import chalk from 'chalk'
import { parse as parseJSONC } from 'jsonc-parser'
import {DateTime} from 'luxon';
import * as semver from 'semver'

import { cacheFolder, readLine, getWeekNumber } from './util'

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

/**
 * Default path of JSONC containing release configuration.
 */
const configPath = 'release-config.jsonc'

/**
 * Loads configuration from predefined path. It does not do any special validation.
 */
export function loadConfig(): Config {
    return parseJSONC(readFileSync(configPath).toString()) as Config
}

/**
 * Convenience function for getting relevant configured releases as semver.SemVer
 *
 * It prompts for a confirmation of the `upcomingRelease` that is cached for a week.
 */
export async function releaseVersions(config: ReleaseConfig): Promise<ReleaseDefinition> {
    const parseOptions: semver.Options = { loose: false }
    const parsedPrevious = semver.parse(config.previousRelease, parseOptions)
    if (!parsedPrevious) {
        throw new Error(`config.previousRelease '${config.previousRelease}' is not valid semver`)
    }
    const parsedUpcoming = semver.parse(config.upcomingRelease, parseOptions)
    if (!parsedUpcoming) {
        throw new Error(`config.upcomingRelease '${config.upcomingRelease}' is not valid semver`)
    }

    // Verify the configured upcoming release. The response is cached and expires in a
    // week, after which the captain is required to confirm again.
    const now = new Date()
    const cachedVersionResponse = `${cacheFolder}/current_release_${now.getUTCFullYear()}_${getWeekNumber(now)}.txt`
    const confirmVersion = await readLine(
        `Please confirm the upcoming release version configured in '${configPath}' (currently '${config.upcomingRelease}') by entering it again: `,
        cachedVersionResponse
    )
    const parsedConfirmed = semver.parse(confirmVersion, parseOptions)
    let error = ''
    if (!parsedConfirmed) {
        error = `Provided version '${confirmVersion}' is not valid semver`
    } else if (semver.neq(parsedConfirmed, parsedUpcoming)) {
        error = `Provided version '${confirmVersion}' and config.upcomingRelease '${config.upcomingRelease}' do not match - please update the release configuration at '${configPath}' and try again`
    }

    // If error, abort and remove the cached response (since it is invalid anyway)
    if (error !== '') {
        unlinkSync(cachedVersionResponse)
        throw new Error(error)
    }

    const versions = {
        previous: parsedPrevious,
        upcoming: parsedUpcoming,
    }
    console.log(`Using versions: { upcoming: ${versions.upcoming.format()}, previous: ${versions.previous.format()} }`)
    return versions
}

const releaseConfigPath = 'release-config-new.jsonc'

export function loadReleaseConfig(): ReleaseConfig {
    return parseJSONC(readFileSync(releaseConfigPath).toString()) as ReleaseConfig
}

export function saveReleaseConfig(config: ReleaseConfig): void {
    writeFileSync(releaseConfigPath, JSON.stringify(config, null, 2))
}

export function newRelease(version: string, previous: string, releaseDate: DateTime, captainGithub: string, captainSlack: string): ReleaseDefinition {
    return {
        ...releaseDates(releaseDate),
        current: version,
        previous,
        captainGitHubUsername: captainGithub,
        captainSlackUsername:captainSlack
    } as ReleaseDefinition
}

export async function newReleaseFromInput(): Promise<ReleaseDefinition> {

    const version = await retryInput('Enter the next version number (ex. 5.4.0): ', val => !!semver.parse(val))
    const previous = await retryInput('Enter the previous version number (ex. 5.4.0): ', val => !!semver.parse(val))

    const releaseDate = await retryInput('Enter the release date (YYYY-MM-DD): ', val => /^\d{4}-\d{2}-\d{2}$/.test(val), 'invalid date, expected format YYYY-MM-DD')

    const captainGithubUsername = await retryInput('Enter the github username of the release captain: ', val => !!val)
    const captainSlackUsername = await retryInput('Enter the slack username of the release captain: ', val => !!val)

    const rel = newRelease(version, previous, DateTime.fromISO(releaseDate, {zone: 'America/Los_Angeles'}), captainGithubUsername, captainSlackUsername)
    console.log(chalk.green('Version created:'))
    console.log(chalk.green(JSON.stringify(rel, null, 2)))
    return rel
}

async function retryInput(prompt: string, delegate: (val: string) => boolean, errorMessage?: string): Promise<string> {
    while (true) {
        const val = await readLine(prompt)
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
function releaseDates(releaseDate: DateTime): ReleaseDates {
    releaseDate = releaseDate.set({hour: 10})
    return {
        codeFreezeDate: releaseDate.plus({days:-7}).toString(),
        securityApprovalDate: releaseDate.plus({days:-7}).toString(),
        releaseDate: releaseDate.toString()
    } as ReleaseDates
}

export function addRelease(config: ReleaseConfig, release: ReleaseDefinition): ReleaseConfig {
    config.releases[release.current] = release
    return config
}

export function removeRelease(config: ReleaseConfig, version: string): ReleaseConfig {
    delete config.releases[version]
    return config
}

interface ReleaseDates {
    releaseDate: string
    codeFreezeDate: string
    securityApprovalDate: string
}

export interface ReleaseConfig {
    metadata: {
        teamEmail: string
        slackAnnounceChannel: string
    },
    releases: {
        [version: string]: ReleaseDefinition
    }
    dryRun: {
        tags?: boolean
        changesets?: boolean
        trackingIssues?: boolean
        slack?: boolean
        calendar?: boolean
    }
}

export interface ReleaseDefinition extends ReleaseDates {
    current: string
    previous: string

    captainSlackUsername: string
    captainGitHubUsername: string
}
