import { ensureEvent, getClient, EventOptions } from './google-calendar'
import { postMessage } from './slack'
import { ensureTrackingIssue, getTrackingIssueURL, getAuthenticatedGitHubClient } from './github'
import * as persistedConfig from './config.json'
import { addHours, addMinutes, subMinutes } from 'date-fns'
import { spawn } from 'child_process'
import * as semver from 'semver'

interface Config {
    teamEmail: string

    captainSlackUsername: string
    captainGitHubUsername: string

    majorMinorVersion: string
    releaseDateTime: string
    oneWorkingDayBeforeRelease: string
    threeWorkingDaysBeforeRelease: string
    fourWorkingDaysBeforeRelease: string
    fiveWorkingDaysBeforeRelease: string
    retrospectiveReminderDateTime: string
    retrospectiveDateTime: string
    retrospectiveDocURL: string
}

type StepID =
    | 'add-timeline-to-calendar'
    | 'help'
    | 'tracking-issue:announce'
    | 'tracking-issue:create'
    | 'release-candidate:create'
    | 'release-candidate:dev-announce'
    | 'qa-start:dev-announce'
    | '_test:google-calendar'

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
                            (argNames && argNames.length > 0 ? ' ' + argNames.map(n => `<${n}>`).join(' ') : '')
                    )
                    .join('\n')
            )
        },
    },
    {
        id: '_test:google-calendar',
        run: async c => {
            const googleCalendar = await getClient()
            await ensureEvent(
                {
                    title: 'TEST EVENT',
                    startDateTime: new Date(c.releaseDateTime).toISOString(),
                    endDateTime: addMinutes(new Date(c.releaseDateTime), 1).toISOString(),
                },
                googleCalendar
            )
        },
    },
    {
        id: 'add-timeline-to-calendar',
        run: async c => {
            const googleCalendar = await getClient()
            const events: EventOptions[] = [
                {
                    title: 'Release captain reminder: 5 working days before release',
                    description: 'See the release tracking issue for TODOs',
                    startDateTime: new Date(c.fiveWorkingDaysBeforeRelease).toISOString(),
                    endDateTime: addMinutes(new Date(c.fiveWorkingDaysBeforeRelease), 1).toISOString(),
                },
                {
                    title: 'Release captain reminder: 4 working days before release',
                    description: 'See the release tracking issue for TODOs',
                    startDateTime: new Date(c.fourWorkingDaysBeforeRelease).toISOString(),
                    endDateTime: addMinutes(new Date(c.fourWorkingDaysBeforeRelease), 1).toISOString(),
                },
                {
                    title: 'Release captain reminder: 3 working days before release',
                    description: 'See the release tracking issue for TODOs',
                    startDateTime: new Date(c.threeWorkingDaysBeforeRelease).toISOString(),
                    endDateTime: addMinutes(new Date(c.threeWorkingDaysBeforeRelease), 1).toISOString(),
                },
                {
                    title: 'Release captain reminder: 1 working day before release',
                    description: 'See the release tracking issue for TODOs',
                    startDateTime: new Date(c.oneWorkingDayBeforeRelease).toISOString(),
                    endDateTime: addMinutes(new Date(c.oneWorkingDayBeforeRelease), 1).toISOString(),
                },
                {
                    title: `Cut release branch ${c.majorMinorVersion}`,
                    description: '(This is not an actual event to attend, just a calendar marker.)',
                    anyoneCanAddSelf: true,
                    attendees: [c.teamEmail],
                    startDateTime: new Date(c.fourWorkingDaysBeforeRelease).toISOString(),
                    endDateTime: addMinutes(new Date(c.fourWorkingDaysBeforeRelease), 1).toISOString(),
                },
                {
                    title: `Release Sourcegraph ${c.majorMinorVersion}`,
                    description: '(This is not an actual event to attend, just a calendar marker.)',
                    anyoneCanAddSelf: true,
                    attendees: [c.teamEmail],
                    startDateTime: new Date(c.releaseDateTime).toISOString(),
                    endDateTime: addMinutes(new Date(c.releaseDateTime), 1).toISOString(),
                },
                {
                    title: `Reminder to submit feedback for ${c.majorMinorVersion} Engineering Retrospective`,
                    description: `Retrospective document: ${c.retrospectiveDocURL}\n\n(This is not an actual event to attend, just a calendar marker.)`,
                    anyoneCanAddSelf: true,
                    attendees: [c.teamEmail],
                    startDateTime: new Date(c.retrospectiveReminderDateTime).toISOString(),
                    endDateTime: addMinutes(new Date(c.retrospectiveReminderDateTime), 1).toISOString(),
                },
                {
                    title: 'Release captain reminder: set up Zoom for engineering retrospective',
                    description:
                        'Go to https://zoom.us/, click "Host a meeting > With Video On", and add the link to the Retrospective calendar event',
                    startDateTime: subMinutes(new Date(c.retrospectiveDateTime), 15).toISOString(),
                    endDateTime: new Date(c.retrospectiveDateTime).toISOString(),
                },
                {
                    title: `Engineering Retrospective ${c.majorMinorVersion}`,
                    description: `Retrospective document: ${c.retrospectiveDocURL}`,
                    anyoneCanAddSelf: true,
                    attendees: [c.teamEmail],
                    startDateTime: new Date(c.retrospectiveDateTime).toISOString(),
                    endDateTime: addHours(new Date(c.retrospectiveDateTime), 1).toISOString(),
                },
            ]

            for (const event of events) {
                console.log(`Create calendar event: ${event.title}`)
                await ensureEvent(event, googleCalendar)
            }
        },
    },
    {
        id: 'tracking-issue:create',
        run: async ({
            majorMinorVersion: version,
            releaseDateTime,
            captainGitHubUsername,
            oneWorkingDayBeforeRelease,
            threeWorkingDaysBeforeRelease,
            fourWorkingDaysBeforeRelease,
            fiveWorkingDaysBeforeRelease,
        }: Config) => {
            const { url, created } = await ensureTrackingIssue({
                version,
                assignees: [captainGitHubUsername],
                releaseDateTime: new Date(releaseDateTime),
                oneWorkingDayBeforeRelease: new Date(oneWorkingDayBeforeRelease),
                threeWorkingDaysBeforeRelease: new Date(threeWorkingDaysBeforeRelease),
                fourWorkingDaysBeforeRelease: new Date(fourWorkingDaysBeforeRelease),
                fiveWorkingDaysBeforeRelease: new Date(fiveWorkingDaysBeforeRelease),
            })
            console.log(created ? `Created tracking issue ${url}` : `Tracking issue already exists: ${url}`)
        },
    },
    {
        id: 'tracking-issue:announce',
        run: async c => {
            const trackingIssueURL = await getTrackingIssueURL(
                await getAuthenticatedGitHubClient(),
                c.majorMinorVersion
            )
            if (!trackingIssueURL) {
                throw new Error(`Tracking issue for version ${c.majorMinorVersion} not found--has it been create yet?`)
            }
            const formatDate = (d: Date): string =>
                `${d.toLocaleString('en-US', {
                    timeZone: 'America/Los_Angeles',
                    dateStyle: 'medium',
                    timeStyle: 'short',
                } as Intl.DateTimeFormatOptions)} (SF time) / ${d.toLocaleString('en-US', {
                    timeZone: 'Europe/Berlin',
                    dateStyle: 'medium',
                    timeStyle: 'short',
                } as Intl.DateTimeFormatOptions)} (Berlin time)`
            await postMessage(`:captain: ${c.majorMinorVersion} Release :captain:
Release captain: @${c.captainSlackUsername}
Tracking issue: ${trackingIssueURL}
Key dates:
- Release branch cut, testing commences: ${formatDate(new Date(c.fourWorkingDaysBeforeRelease))}
- Final release tag: ${formatDate(new Date(c.oneWorkingDayBeforeRelease))}
- Release: ${formatDate(new Date(c.releaseDateTime))}}
- Retrospective: ${formatDate(new Date(c.retrospectiveDateTime))}`)
        },
    },
    {
        id: 'release-candidate:create',
        argNames: ['version'],
        run: (_config, version) => {
            const parsedVersion = semver.parse(version, { loose: false })
            if (!parsedVersion) {
                throw new Error(`version ${version} is not valid semver`)
            }
            const tag = JSON.stringify(`v${parsedVersion.version}`)
            console.log(`Creating and pushing tag ${tag}`)
            const child = spawn('bash', ['-c', `git tag -a ${tag} -m ${tag} && git push origin ${tag}`])
            child.stdout.pipe(process.stdout)
            child.stderr.pipe(process.stderr)
        },
    },
    {
        id: 'release-candidate:dev-announce',
        run: () => {
            console.log('NOT YET IMPLEMENTED')
            process.exit(1)
        },
    },
    {
        id: 'qa-start:dev-announce',
        run: () => {
            console.log('NOT YET IMPLEMENTED')
            process.exit(1)
        },
    },
]

async function run(config: Config, stepIDToRun: StepID, ...stepArgs: string[]): Promise<void> {
    await Promise.all(
        steps
            .filter(({ id }) => id === stepIDToRun)
            .map(async step => {
                if (step.run) {
                    await step.run(config, ...stepArgs)
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
    if (args.length < 1) {
        console.error('This command expects at least 1 argument')
        await run(config, 'help')
        return
    }
    const step = args[0]
    if (!steps.map(({ id }) => id as string).includes(step)) {
        console.error('Unrecognized step', JSON.stringify(step))
        return
    }
    const stepArgs = args.slice(1)
    await run(config, step as StepID, ...stepArgs)
}

main().catch(err => console.error(err))
