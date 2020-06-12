import { ensureEvent, getClient, EventOptions } from './google-calendar'
import { postMessage } from './slack'
import {
    ensureTrackingIssue,
    getAuthenticatedGitHubClient,
    listIssues,
    getIssueByTitle,
    trackingIssueTitle,
    ensurePatchReleaseIssue,
    createBranchWithChanges,
    createPR,
    CreateBranchWithChangesOptions,
} from './github'
import * as persistedConfig from './config.json'
import { addMinutes, isWeekend, eachDayOfInterval, addDays, subDays } from 'date-fns'
import * as semver from 'semver'
import commandExists from 'command-exists'
import { PullsCreateParams } from '@octokit/rest'
import execa from 'execa'

const sed = process.platform === 'linux' ? 'sed' : 'gsed'

const formatDate = (date: Date): string =>
    `${date.toLocaleString('en-US', {
        timeZone: 'America/Los_Angeles',
        dateStyle: 'medium',
        timeStyle: 'short',
    } as Intl.DateTimeFormatOptions)} (SF time) / ${date.toLocaleString('en-US', {
        timeZone: 'Europe/Berlin',
        dateStyle: 'medium',
        timeStyle: 'short',
    } as Intl.DateTimeFormatOptions)} (Berlin time)`

interface Config {
    teamEmail: string

    captainSlackUsername: string
    captainGitHubUsername: string

    majorVersion: string
    minorVersion: string
    releaseDateTime: string
    oneWorkingDayBeforeRelease: string
    fourWorkingDaysBeforeRelease: string
    fiveWorkingDaysBeforeRelease: string

    slackAnnounceChannel: string
}

type StepID =
    | 'add-timeline-to-calendar'
    | 'help'
    | 'tracking-issue:announce'
    | 'tracking-issue:create'
    | 'release-candidate:create'
    | 'release-candidate:dev-announce'
    | 'qa-start:dev-announce'
    | 'patch:issue'
    | 'release:publish'
    | '_test:google-calendar'
    | '_test:slack'

interface Step {
    id: StepID
    run?: ((config: Config, ...args: string[]) => Promise<void>) | ((config: Config, ...args: string[]) => void)
    argNames?: string[]
}

const steps: Step[] = [
    {
        id: 'help',
        run: () => {
            console.error('Steps are:')
            console.error(
                steps
                    .filter(({ id }) => !id.startsWith('_'))
                    .map(
                        ({ id, argNames }) =>
                            '\t' +
                            id +
                            (argNames && argNames.length > 0
                                ? ' ' + argNames.map(argumentName => `<${argumentName}>`).join(' ')
                                : '')
                    )
                    .join('\n')
            )
        },
    },
    {
        id: '_test:google-calendar',
        run: async config => {
            const googleCalendar = await getClient()
            await ensureEvent(
                {
                    title: 'TEST EVENT',
                    startDateTime: new Date(config.releaseDateTime).toISOString(),
                    endDateTime: addMinutes(new Date(config.releaseDateTime), 1).toISOString(),
                },
                googleCalendar
            )
        },
    },
    {
        id: '_test:slack',
        run: async (_config, message) => {
            await postMessage(message, '_test-channel')
        },
    },
    {
        id: 'add-timeline-to-calendar',
        run: async config => {
            const googleCalendar = await getClient()
            const events: EventOptions[] = [
                {
                    title: 'Release captain: prepare for branch cut (5 working days until release)',
                    description: 'See the release tracking issue for TODOs',
                    startDateTime: new Date(config.fiveWorkingDaysBeforeRelease).toISOString(),
                    endDateTime: addMinutes(new Date(config.fiveWorkingDaysBeforeRelease), 1).toISOString(),
                },
                {
                    title: 'Release captain: branch cut (4 working days until release)',
                    description: 'See the release tracking issue for TODOs',
                    startDateTime: new Date(config.fourWorkingDaysBeforeRelease).toISOString(),
                    endDateTime: addMinutes(new Date(config.fourWorkingDaysBeforeRelease), 1).toISOString(),
                },
                ...eachDayOfInterval({
                    start: addDays(new Date(config.fourWorkingDaysBeforeRelease), 1),
                    end: subDays(new Date(config.oneWorkingDayBeforeRelease), 1),
                })
                    .filter(date => !isWeekend(date))
                    .map(date => ({
                        title: 'Release captain: cut new release candidate',
                        description: 'See release tracking issue for TODOs',
                        startDateTime: date.toISOString(),
                        endDateTime: addMinutes(date, 1).toISOString(),
                    })),
                {
                    title: 'Release captain: tag final release (1 working day before release)',
                    description: 'See the release tracking issue for TODOs',
                    startDateTime: new Date(config.oneWorkingDayBeforeRelease).toISOString(),
                    endDateTime: addMinutes(new Date(config.oneWorkingDayBeforeRelease), 1).toISOString(),
                },
                {
                    title: `Cut release branch ${config.majorVersion}.${config.minorVersion}`,
                    description: '(This is not an actual event to attend, just a calendar marker.)',
                    anyoneCanAddSelf: true,
                    attendees: [config.teamEmail],
                    startDateTime: new Date(config.fourWorkingDaysBeforeRelease).toISOString(),
                    endDateTime: addMinutes(new Date(config.fourWorkingDaysBeforeRelease), 1).toISOString(),
                },
                {
                    title: `Release Sourcegraph ${config.majorVersion}.${config.minorVersion}`,
                    description: '(This is not an actual event to attend, just a calendar marker.)',
                    anyoneCanAddSelf: true,
                    attendees: [config.teamEmail],
                    startDateTime: new Date(config.releaseDateTime).toISOString(),
                    endDateTime: addMinutes(new Date(config.releaseDateTime), 1).toISOString(),
                },
            ]

            for (const event of events) {
                console.log(`Create calendar event: ${event.title}: ${event.startDateTime!}`)
                await ensureEvent(event, googleCalendar)
            }
        },
    },
    {
        id: 'tracking-issue:create',
        run: async ({
            majorVersion,
            minorVersion,
            releaseDateTime,
            captainGitHubUsername,
            oneWorkingDayBeforeRelease,
            fourWorkingDaysBeforeRelease,
            fiveWorkingDaysBeforeRelease,
        }: Config) => {
            const { url, created } = await ensureTrackingIssue({
                majorVersion,
                minorVersion,
                assignees: [captainGitHubUsername],
                releaseDateTime: new Date(releaseDateTime),
                oneWorkingDayBeforeRelease: new Date(oneWorkingDayBeforeRelease),
                fourWorkingDaysBeforeRelease: new Date(fourWorkingDaysBeforeRelease),
                fiveWorkingDaysBeforeRelease: new Date(fiveWorkingDaysBeforeRelease),
            })
            console.log(created ? `Created tracking issue ${url}` : `Tracking issue already exists: ${url}`)
        },
    },
    {
        id: 'tracking-issue:announce',
        run: async config => {
            const trackingIssueURL = await getIssueByTitle(
                await getAuthenticatedGitHubClient(),
                trackingIssueTitle(config.majorVersion, config.minorVersion)
            )
            if (!trackingIssueURL) {
                throw new Error(
                    `Tracking issue for version ${config.majorVersion}.${config.minorVersion} not found--has it been create yet?`
                )
            }
            await postMessage(
                `:captain: ${config.majorVersion}.${config.minorVersion} Release :captain:
Release captain: @${config.captainSlackUsername}
Tracking issue: ${trackingIssueURL}
Key dates:
- Release branch cut, testing commences: ${formatDate(new Date(config.fourWorkingDaysBeforeRelease))}
- Final release tag: ${formatDate(new Date(config.oneWorkingDayBeforeRelease))}
- Release: ${formatDate(new Date(config.releaseDateTime))}`,
                config.slackAnnounceChannel
            )
        },
    },
    {
        id: 'release-candidate:create',
        argNames: ['version'],
        run: async (_config, version) => {
            const parsedVersion = semver.parse(version, { loose: false })
            if (!parsedVersion) {
                throw new Error(`version ${version} is not valid semver`)
            }
            const tag = JSON.stringify(`v${parsedVersion.version}`)
            const branch = JSON.stringify(`${parsedVersion.major}.${parsedVersion.minor}`)
            console.log(`Creating and pushing tag ${tag}`)
            await execa(
                'bash',
                [
                    '-c',
                    `git diff --quiet && git checkout ${branch} && git pull --rebase && git tag -a ${tag} -m ${tag} && git push origin ${tag}`,
                ],
                { stdio: 'inherit' }
            )
        },
    },
    {
        id: 'release-candidate:dev-announce',
        run: async (config, version) => {
            const parsedVersion = semver.parse(version, { loose: false })
            if (!parsedVersion) {
                throw new Error(`version ${version} is not valid semver`)
            }

            const query = `is:open is:issue milestone:${config.majorVersion}.${config.minorVersion} label:release-blocker`
            const issues = await listIssues(await getAuthenticatedGitHubClient(), query)
            const issuesURL = `https://github.com/issues?q=${encodeURIComponent(query)}`
            const releaseBlockerMessage =
                issues.length === 0
                    ? 'There are currently ZERO release blocking issues'
                    : issues.length === 1
                    ? `There is 1 release-blocking issue: ${issuesURL}`
                    : `There are ${issues.length} release-blocking issues: ${issuesURL}`

            const message = `:captain: Release \`${version}\` has been cut :captain:

- Please ensure \`CHANGELOG.md\` on \`main\` is up-to-date.
- Run this release locally with \`IMAGE=sourcegraph/server:${version} ./dev/run-server-image.sh\`
- It will be deployed to k8s.sgdev.org within approximately one hour (https://k8s.sgdev.org/site-admin/updates)
- ${releaseBlockerMessage}
            `
            await postMessage(message, config.slackAnnounceChannel)
        },
    },
    {
        id: 'patch:issue',
        run: async ({ captainGitHubUsername, slackAnnounceChannel }, version) => {
            const parsedVersion = semver.parse(version, { loose: false })
            if (!parsedVersion) {
                throw new Error(`version ${version} is not valid semver`)
            }
            if (parsedVersion.prerelease.length > 0) {
                throw new Error(`version ${version} is pre-release`)
            }

            // Create issue
            const { url, created } = await ensurePatchReleaseIssue({
                version: parsedVersion,
                assignees: [captainGitHubUsername],
            })
            const existsText = created ? '' : ' (already exists)'
            console.log(`Patch release issue URL${existsText}: ${url}`)
            if (!created) {
                return
            }

            // - Announce issue if issue does not already exist
            await postMessage(
                `:captain: Patch release ${parsedVersion.version} will be published soon. If you have changes that should go into this patch release, please add your item to the checklist in the issue description: ${url}`,
                slackAnnounceChannel
            )
            console.log(`Posted to Slack channel ${slackAnnounceChannel}`)
        },
    },
    {
        id: 'release:publish',
        run: async ({ slackAnnounceChannel }, version) => {
            const parsedVersion = semver.parse(version, { loose: false })
            if (!parsedVersion) {
                throw new Error(`version ${version} is not valid semver`)
            }
            if (parsedVersion.prerelease.length > 0) {
                throw new Error(`version ${version} is pre-release`)
            }
            const requiredCommands = ['comby', sed, 'find']
            for (const command of requiredCommands) {
                try {
                    await commandExists(command)
                } catch {
                    throw new Error(`Required command ${command} does not exist`)
                }
            }

            const changes: (PullsCreateParams & CreateBranchWithChangesOptions)[] = [
                {
                    owner: 'sourcegraph',
                    repo: 'sourcegraph',
                    base: 'main',
                    head: `publish-${parsedVersion.version}`,
                    commitMessage: `Update latest release to ${parsedVersion.version}`,
                    bashEditCommands: [
                        `find . -type f -name '*.md' ! -name 'CHANGELOG.md' -exec ${sed} -i -E 's/sourcegraph\\/server:[0-9]+\\.[0-9]+\\.[0-9]+/sourcegraph\\/server:${parsedVersion.version}/g' {} +`,
                        `${sed} -i -E 's/version \`[0-9]+\\.[0-9]+\\.[0-9]+\`/version \`${parsedVersion.version}\`/g' doc/index.md`,
                        parsedVersion.patch === 0
                            ? `comby -in-place '{{$previousReleaseRevspec := ":[1]"}} {{$previousReleaseVersion := ":[2]"}} {{$currentReleaseRevspec := ":[3]"}} {{$currentReleaseVersion := ":[4]"}}' '{{$previousReleaseRevspec := ":[3]"}} {{$previousReleaseVersion := ":[4]"}} {{$currentReleaseRevspec := "v${parsedVersion.version}"}} {{$currentReleaseVersion := "${parsedVersion.major}.${parsedVersion.minor}"}}' doc/_resources/templates/document.html`
                            : `comby -in-place 'currentReleaseRevspec := ":[1]"' 'currentReleaseRevspec := "v${parsedVersion.version}"' doc/_resources/templates/document.html`,
                        `comby -in-place 'latestReleaseKubernetesBuild = newBuild(":[1]")' "latestReleaseKubernetesBuild = newBuild(\\"${parsedVersion.version}\\")" cmd/frontend/internal/app/pkg/updatecheck/handler.go`,
                        `comby -in-place 'latestReleaseDockerServerImageBuild = newBuild(":[1]")' "latestReleaseDockerServerImageBuild = newBuild(\\"${parsedVersion.version}\\")" cmd/frontend/internal/app/pkg/updatecheck/handler.go`,
                    ],
                    title: `Update latest release to ${parsedVersion.version}`,
                },
                {
                    owner: 'sourcegraph',
                    repo: 'deploy-sourcegraph-aws',
                    base: 'master',
                    head: `publish-${parsedVersion.version}`,
                    commitMessage: `Update latest release to ${parsedVersion.version}`,
                    bashEditCommands: [
                        `${sed} -i -E 's/export SOURCEGRAPH_VERSION=[0-9]+\\.[0-9]+\\.[0-9]+/export SOURCEGRAPH_VERSION=${parsedVersion.version}/g' resources/amazon-linux2.sh`,
                    ],
                    title: `Update latest release to ${parsedVersion.version}`,
                },
                {
                    owner: 'sourcegraph',
                    repo: 'deploy-sourcegraph-digitalocean',
                    base: 'master',
                    head: `publish-${parsedVersion.version}`,
                    commitMessage: `Update latest release to ${parsedVersion.version}`,
                    bashEditCommands: [
                        `${sed} -i -E 's/export SOURCEGRAPH_VERSION=[0-9]+\\.[0-9]+\\.[0-9]+/export SOURCEGRAPH_VERSION=${parsedVersion.version}/g' resources/user-data.sh`,
                    ],
                    title: `Update latest release to ${parsedVersion.version}`,
                },
            ]
            for (const changeset of changes) {
                await createBranchWithChanges(changeset)
                const prURL = await createPR(changeset)
                console.log(`Pull request created: ${prURL}`)
            }
            await postMessage(
                `${parsedVersion.version} has been released, update deploy-sourcegraph-docker as needed, cc @stephen`,
                slackAnnounceChannel
            )
        },
    },
]

async function run(config: Config, stepIDToRun: StepID, ...stepArguments: string[]): Promise<void> {
    await Promise.all(
        steps
            .filter(({ id }) => id === stepIDToRun)
            .map(async step => {
                if (step.run) {
                    await step.run(config, ...stepArguments)
                }
            })
    )
}

/**
 * Release captain automation
 */
async function main(): Promise<void> {
    const config = persistedConfig
    const args = process.argv.slice(2)
    if (args.length === 0) {
        console.error('This command expects at least 1 argument')
        await run(config, 'help')
        return
    }
    const step = args[0]
    if (!steps.map(({ id }) => id as string).includes(step)) {
        console.error('Unrecognized step', JSON.stringify(step))
        return
    }
    const stepArguments = args.slice(1)
    await run(config, step as StepID, ...stepArguments)
}

main().catch(error => console.error(error))
