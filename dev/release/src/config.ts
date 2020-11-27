import { readLine, getWeekNumber } from './util'
import * as semver from 'semver'
import { readFileSync } from 'fs'
import { parse as parseJSONC } from '@sqs/jsonc-parser'

/**
 * Release configuration file format
 */
export interface Config {
    teamEmail: string

    captainSlackUsername: string
    captainGitHubUsername: string

    previousRelease: string
    upcomingRelease: string

    releaseDateTime: string
    oneWorkingDayBeforeRelease: string
    fourWorkingDaysBeforeRelease: string
    fiveWorkingDaysBeforeRelease: string

    slackAnnounceChannel: string

    dryRun: {
        tags?: boolean
        changesets?: boolean
        trackingIssues?: boolean
    }
}

/**
 * Loads configuration from predefined path. It does not do any special validation.
 */
export function loadConfig(): Config {
    return parseJSONC(readFileSync('release-config.jsonc').toString()) as Config
}

/**
 * Convenience function for getting relevant configured releases as semver.SemVer
 *
 * It prompts for a confirmation of the `upcomingRelease` that is cached for a week.
 */
export async function releaseVersions(
    config: Config
): Promise<{
    previous: semver.SemVer
    upcoming: semver.SemVer
}> {
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
    const cachedVersion = `.secrets/current_release_${now.getUTCFullYear()}_${getWeekNumber(now)}.txt`
    const confirmVersion = await readLine(
        `Please confirm the upcoming release version (configured: '${config.upcomingRelease}'): `,
        cachedVersion
    )
    const parsedConfirmed = semver.parse(confirmVersion, parseOptions)
    if (!parsedConfirmed) {
        throw new Error(`Provided version '${confirmVersion}' is not valid semver (in ${cachedVersion})`)
    }
    if (semver.neq(parsedConfirmed, parsedUpcoming)) {
        throw new Error(
            `Provided version '${confirmVersion}' and config.upcomingRelease '${config.upcomingRelease}' to not match - please update the release configuration, or confirm the version in your cached answer (in ${cachedVersion})`
        )
    }

    const versions = {
        previous: parsedPrevious,
        upcoming: parsedUpcoming,
    }
    console.log(`Using versions: { upcoming: ${versions.upcoming.format()}, previous: ${versions.previous.format()} }`)
    return versions
}
