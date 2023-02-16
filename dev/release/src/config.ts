import {readFileSync, unlinkSync, writeFileSync} from 'fs'

import chalk from 'chalk'
import { parse as parseJSONC } from 'jsonc-parser'
import {DateTime} from 'luxon';
import * as semver from 'semver'
import {SemVer} from 'semver';

import { getPreviousVersion } from './git';
import {cacheFolder, readLine, getWeekNumber, retryInput} from './util'

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
// export async function releaseVersions(config: ReleaseConfig): Promise<ScheduledReleaseDefinition> {
//     const parseOptions: semver.Options = { loose: false }
//     const parsedPrevious = semver.parse(config.previousRelease, parseOptions)
//     if (!parsedPrevious) {
//         throw new Error(`config.previousRelease '${config.previousRelease}' is not valid semver`)
//     }
//     const parsedUpcoming = semver.parse(config.upcomingRelease, parseOptions)
//     if (!parsedUpcoming) {
//         throw new Error(`config.upcomingRelease '${config.upcomingRelease}' is not valid semver`)
//     }
//
//     // Verify the configured upcoming release. The response is cached and expires in a
//     // week, after which the captain is required to confirm again.
//     const now = new Date()
//     const cachedVersionResponse = `${cacheFolder}/current_release_${now.getUTCFullYear()}_${getWeekNumber(now)}.txt`
//     const confirmVersion = await readLine(
//         `Please confirm the upcoming release version configured in '${configPath}' (currently '${config.upcomingRelease}') by entering it again: `,
//         cachedVersionResponse
//     )
//     const parsedConfirmed = semver.parse(confirmVersion, parseOptions)
//     let error = ''
//     if (!parsedConfirmed) {
//         error = `Provided version '${confirmVersion}' is not valid semver`
//     } else if (semver.neq(parsedConfirmed, parsedUpcoming)) {
//         error = `Provided version '${confirmVersion}' and config.upcomingRelease '${config.upcomingRelease}' do not match - please update the release configuration at '${configPath}' and try again`
//     }
//
//     // If error, abort and remove the cached response (since it is invalid anyway)
//     if (error !== '') {
//         unlinkSync(cachedVersionResponse)
//         throw new Error(error)
//     }
//
//     const versions = {
//         previous: parsedPrevious,
//         upcoming: parsedUpcoming,
//     }
//     console.log(`Using versions: { upcoming: ${versions.upcoming.format()}, previous: ${versions.previous.format()} }`)
//     return versions
// }

export async function getActiveRelease(config: ReleaseConfig): Promise<ActiveRelease & ScheduledReleaseDefinition> {
    if (!config.in_progress || config.in_progress.releases.length === 0) {
        console.log(chalk.yellow('No active releases are defined! Attempting to activate...'))
        await activateRelease(config)
    }
    if (config.in_progress.releases.length > 1) {
        throw new Error(chalk.red('The release config has multiple versions activated. This feature is not yet supported by the release tool! Please activate only a single release.'))
    }
    const rel = config.in_progress.releases[0]
    return {...rel, ...config.scheduledReleases[rel.version]}
}

const releaseConfigPath = 'release-config-new.jsonc'

export function loadReleaseConfig(): ReleaseConfig {
    const config = parseJSONC(readFileSync(releaseConfigPath).toString()) as ReleaseConfig
    return config
}

export function saveReleaseConfig(config: ReleaseConfig): void {
    writeFileSync(releaseConfigPath, JSON.stringify(config, null, 2))
}

export function newRelease(version: SemVer, releaseDate: DateTime, captainGithub: string, captainSlack: string): ScheduledReleaseDefinition {
    return {
        ...releaseDates(releaseDate),
        current: version.version,
        captainGitHubUsername: captainGithub,
        captainSlackUsername:captainSlack
    } as ScheduledReleaseDefinition
}

export async function newReleaseFromInput(versionOverride?: SemVer): Promise<ScheduledReleaseDefinition> {
    let version = versionOverride
    if (!version) {
        version = new SemVer(await retryInput('Enter the next version number (ex. 5.4.0): ', val => !!semver.parse(val)))
    }

    const releaseDateStr = await retryInput('Enter the release date (YYYY-MM-DD). Enter blank to use current date: ', val => {
        if (val && /^\d{4}-\d{2}-\d{2}$/.test(val)) {
            return true
        }
        // this will return false if the input doesn't match the regexp above but does exist, allowing blank input to still be valid
        return !val;
    }, 'invalid date, expected format YYYY-MM-DD')
    let releaseTime:DateTime
    if (!releaseDateStr) {
        releaseTime = DateTime.now().setZone('America/Los_Angeles')
        console.log(chalk.blue(`Using current time: ${releaseTime.toString()}`))
    } else {
        releaseTime = DateTime.fromISO(releaseDateStr, {zone: 'America/Los_Angeles'})
    }

    const captainGithubUsername = await retryInput('Enter the github username of the release captain: ', val => !!val)
    const captainSlackUsername = await retryInput('Enter the slack username of the release captain: ', val => !!val)

    const rel = newRelease(version, releaseTime, captainGithubUsername, captainSlackUsername)
    console.log(chalk.green('Version created:'))
    console.log(chalk.green(JSON.stringify(rel, null, 2)))
    return rel
}

function releaseDates(releaseDate: DateTime): ReleaseDates {
    releaseDate = releaseDate.set({hour: 10})
    return {
        codeFreezeDate: releaseDate.plus({days:-7}).toString(),
        securityApprovalDate: releaseDate.plus({days:-7}).toString(),
        releaseDate: releaseDate.toString()
    } as ReleaseDates
}

export function addScheduledRelease(config: ReleaseConfig, release: ScheduledReleaseDefinition): ReleaseConfig {
    config.scheduledReleases[release.current] = release
    return config
}

export function removeScheduledRelease(config: ReleaseConfig, version: string): ReleaseConfig {
    delete config.scheduledReleases[version]
    return config
}

interface ReleaseDates {
    releaseDate: string
    codeFreezeDate: string
    securityApprovalDate: string
}

export interface ActiveRelease {
    version: string
    previous: string
}

export interface InProgress {
    captainSlackUsername: string
    captainGitHubUsername: string
    releases: ActiveRelease[]
}

export interface ReleaseConfig {
    metadata: {
        teamEmail: string
        slackAnnounceChannel: string
    },
    scheduledReleases: {
        [version: string]: ScheduledReleaseDefinition
    }
    in_progress: InProgress
    dryRun: {
        tags?: boolean
        changesets?: boolean
        trackingIssues?: boolean
        slack?: boolean
        calendar?: boolean
    }
}

export interface ScheduledReleaseDefinition extends ReleaseDates {
    current: string

    captainSlackUsername: string
    captainGitHubUsername: string
}

export async function activateRelease(config: ReleaseConfig): Promise<void> {
    const probably = getPreviousVersion().inc('minor')
    const input = await retryInput(`Enter the feature version to activate (probably ${probably.version}): `, val => {
        const version = semver.parse(val)
        return (!(!version || version.patch !== 0))
    })
    const next = new SemVer(input)
    console.log('Attempting to detect previous version...')
    const previous = getPreviousVersion(next)
    console.log(chalk.blue(`Detected previous version: ${previous.version}`))

    let scheduled = config.scheduledReleases[next.version]
    if (!scheduled) {
        console.log(chalk.yellow(`Release definition not found for: ${next.version}, enter release information.\n`))
        scheduled = await newReleaseFromInput(next)
        addScheduledRelease(config, scheduled)
    }
    config.in_progress = {
        captainGitHubUsername: scheduled.captainGitHubUsername,
        captainSlackUsername: scheduled.captainSlackUsername,
        releases: [{version: next.version, previous: previous.version}]
    } as InProgress
    saveReleaseConfig(config)
    console.log(chalk.green(`Release: ${next.version} activated!`))
}
