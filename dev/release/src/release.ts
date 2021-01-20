import { ensureEvent, getClient, EventOptions } from './google-calendar'
import { postMessage } from './slack'
import {
    getAuthenticatedGitHubClient,
    listIssues,
    getTrackingIssue,
    ensureReleaseTrackingIssue,
    ensurePatchReleaseIssue,
    createChangesets,
    CreatedChangeset,
    createTag,
} from './github'
import * as changelog from './changelog'
import * as campaigns from './campaigns'
import { Config, releaseVersions, loadConfig } from './config'
import { formatDate, timezoneLink } from './util'
import { addMinutes, isWeekend, eachDayOfInterval, addDays, subDays } from 'date-fns'
import { readFileSync, writeFileSync } from 'fs'
import * as path from 'path'
import commandExists from 'command-exists'

const sed = process.platform === 'linux' ? 'sed' : 'gsed'

type StepID =
    | 'help'
    // release tracking
    | 'tracking:release-timeline'
    | 'tracking:release-issue'
    | 'tracking:patch-issue'
    // branch cut
    | 'changelog:cut'
    // release
    | 'release:status'
    | 'release:create-candidate'
    | 'release:stage'
    | 'release:add-to-campaign'
    | 'release:finalize'
    | 'release:close'
    // testing
    | '_test:google-calendar'
    | '_test:slack'
    | '_test:campaign-create-from-changes'
    | '_test:config'

interface Step {
    id: StepID
    description: string
    run?: ((config: Config, ...args: string[]) => Promise<void>) | ((config: Config, ...args: string[]) => void)
    argNames?: string[]
}

const steps: Step[] = [
    {
        id: 'help',
        description: 'Output help text about this tool',
        argNames: ['all'],
        run: (_config, all) => {
            console.error('Sourcegraph release tool - https://about.sourcegraph.com/handbook/engineering/releases')
            console.error('\nUSAGE\n')
            console.error('\tyarn run release <step>')
            console.error('\nAVAILABLE STEPS\n')
            console.error(
                steps
                    .filter(({ id }) => all || !id.startsWith('_'))
                    .map(
                        ({ id, argNames, description }) =>
                            '\t' +
                            id +
                            (argNames && argNames.length > 0
                                ? ' ' + argNames.map(argumentName => `<${argumentName}>`).join(' ')
                                : '') +
                            '\n\t\t' +
                            description
                    )
                    .join('\n') + '\n'
            )
        },
    },
    {
        id: 'tracking:release-timeline',
        description: 'Generate a set of Google Calendar events for a MAJOR.MINOR release',
        run: async config => {
            const googleCalendar = await getClient()
            const { upcoming: release } = await releaseVersions(config)
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
                    title: `Cut release branch ${release.major}.${release.minor}`,
                    description: '(This is not an actual event to attend, just a calendar marker.)',
                    anyoneCanAddSelf: true,
                    attendees: [config.teamEmail],
                    startDateTime: new Date(config.fourWorkingDaysBeforeRelease).toISOString(),
                    endDateTime: addMinutes(new Date(config.fourWorkingDaysBeforeRelease), 1).toISOString(),
                },
                {
                    title: `Release Sourcegraph ${release.major}.${release.minor}`,
                    description: '(This is not an actual event to attend, just a calendar marker.)',
                    anyoneCanAddSelf: true,
                    attendees: [config.teamEmail],
                    startDateTime: new Date(config.releaseDateTime).toISOString(),
                    endDateTime: addMinutes(new Date(config.releaseDateTime), 1).toISOString(),
                },
            ]

            for (const event of events) {
                console.log(`Create calendar event: ${event.title}: ${event.startDateTime || 'undefined'}`)
                await ensureEvent(event, googleCalendar)
            }
        },
    },
    {
        id: 'tracking:release-issue',
        description: 'Generate a GitHub tracking issue for a MAJOR.MINOR release',
        run: async (config: Config) => {
            const {
                releaseDateTime,
                captainGitHubUsername,
                oneWorkingDayBeforeRelease,
                fourWorkingDaysBeforeRelease,
                fiveWorkingDaysBeforeRelease,

                captainSlackUsername,
                slackAnnounceChannel,
                dryRun,
            } = config
            const { upcoming: release } = await releaseVersions(config)

            // Create issue
            const { url, created } = await ensureReleaseTrackingIssue({
                version: release,
                assignees: [captainGitHubUsername],
                releaseDateTime: new Date(releaseDateTime),
                oneWorkingDayBeforeRelease: new Date(oneWorkingDayBeforeRelease),
                fourWorkingDaysBeforeRelease: new Date(fourWorkingDaysBeforeRelease),
                fiveWorkingDaysBeforeRelease: new Date(fiveWorkingDaysBeforeRelease),
                dryRun: dryRun.trackingIssues || false,
            })
            if (url) {
                console.log(created ? `Created tracking issue ${url}` : `Tracking issue already exists: ${url}`)
            }

            // Announce issue if issue does not already exist
            if (created) {
                // Slack markdown links
                const majorMinor = `${release.major}.${release.minor}`
                const branchCutDate = new Date(fourWorkingDaysBeforeRelease)
                const branchCutDateString = `<${timezoneLink(branchCutDate, `${majorMinor} branch cut`)}|${formatDate(
                    branchCutDate
                )}>`
                const releaseDate = new Date(releaseDateTime)
                const releaseDateString = `<${timezoneLink(releaseDate, `${majorMinor} release`)}|${formatDate(
                    releaseDate
                )}>`
                await postMessage(
                    `:mega: *${majorMinor} Release*

:captain: Release captain: @${captainSlackUsername}
:pencil: Tracking issue: ${url}
:spiral_calendar_pad: Key dates:
* Branch cut: ${branchCutDateString}
* Release: ${releaseDateString}`,
                    slackAnnounceChannel
                )
                console.log(`Posted to Slack channel ${slackAnnounceChannel}`)
            }
        },
    },
    {
        id: 'tracking:patch-issue',
        description: 'Generate a GitHub tracking issue for a MAJOR.MINOR.PATCH release',
        run: async config => {
            const { captainGitHubUsername, captainSlackUsername, slackAnnounceChannel, dryRun } = config
            const { upcoming: release } = await releaseVersions(config)

            // Create issue
            const { url, created } = await ensurePatchReleaseIssue({
                version: release,
                assignees: [captainGitHubUsername],
                dryRun: dryRun.trackingIssues || false,
            })
            if (url) {
                console.log(created ? `Created tracking issue ${url}` : `Tracking issue already exists: ${url}`)
            }

            // Announce issue if issue does not already exist
            if (created) {
                const patchRequestTemplate = `https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=team%2Fdistribution&template=request_patch_release.md&title=${release.version}%3A+`
                await postMessage(
                    `:mega: *${release.version} Patch Release*

:captain: Release captain: @${captainSlackUsername}
:pencil: Tracking issue: ${url}

If you have changes that should go into this patch release, <${patchRequestTemplate}|please *file a patch request issue*>, or it will not be included.`,
                    slackAnnounceChannel
                )
                console.log(`Posted to Slack channel ${slackAnnounceChannel}`)
            }
        },
    },
    {
        id: 'changelog:cut',
        description: 'Generate pull requests to perform a changelog cut for branch cut',
        argNames: ['changelogFile'],
        run: async (config, changelogFile = 'CHANGELOG.md') => {
            const { upcoming: release } = await releaseVersions(config)
            const prMessage = `changelog: cut sourcegraph@${release.version}`
            await createChangesets({
                requiredCommands: [],
                changes: [
                    {
                        owner: 'sourcegraph',
                        repo: 'sourcegraph',
                        base: 'main',
                        head: `publish-${release.version}`,
                        title: prMessage,
                        commitMessage: prMessage,
                        edits: [
                            (directory: string) => {
                                console.log(`Updating '${changelogFile} for ${release.format()}'`)
                                const changelogPath = path.join(directory, changelogFile)
                                let changelogContents = readFileSync(changelogPath).toString()

                                // Convert 'unreleased' to a release
                                const releaseHeader = `## ${release.format()}`
                                const unreleasedHeader = '## Unreleased'
                                changelogContents = changelogContents.replace(unreleasedHeader, releaseHeader)

                                // Add a blank changelog template for the next release
                                changelogContents = changelogContents.replace(
                                    changelog.divider,
                                    changelog.releaseTemplate
                                )

                                // Update changelog
                                writeFileSync(changelogPath, changelogContents)
                            },
                        ], // Changes already done
                    },
                ],
                dryRun: config.dryRun.changesets,
            })
        },
    },
    {
        id: 'release:status',
        description: 'Post a message in Slack summarizing the progress of a release',
        run: async config => {
            const githubClient = await getAuthenticatedGitHubClient()
            const { upcoming: release } = await releaseVersions(config)

            const trackingIssue = await getTrackingIssue(githubClient, release)
            if (!trackingIssue) {
                throw new Error(`Tracking issue for version ${release.version} not found - has it been created yet?`)
            }

            const blockingQuery = 'is:open org:sourcegraph label:release-blocker'
            const blockingIssues = await listIssues(githubClient, blockingQuery)
            const blockingIssuesURL = `https://github.com/issues?q=${encodeURIComponent(blockingQuery)}`
            const blockingMessage =
                blockingIssues.length === 0
                    ? 'There are no release-blocking issues'
                    : `There ${
                          blockingIssues.length === 1
                              ? 'is 1 release-blocking issue'
                              : `are ${blockingIssues.length} release-blocking issues`
                      }`

            const message = `:mega: *${release.version} Release Status Update*

* Tracking issue: ${trackingIssue.url}
* ${blockingMessage}: ${blockingIssuesURL}`
            await postMessage(message, config.slackAnnounceChannel)
        },
    },
    {
        id: 'release:create-candidate',
        description: 'Generate the Nth release candidate. Set <candidate> to "final" to generate a final release',
        argNames: ['candidate'],
        run: async (config, candidate) => {
            if (!candidate) {
                throw new Error('Candidate information is required (either "final" or a number)')
            }
            const { upcoming: release } = await releaseVersions(config)
            const branch = `${release.major}.${release.minor}`
            const tag = `v${release.version}${candidate === 'final' ? '' : `-rc.${candidate}`}`
            await createTag(
                await getAuthenticatedGitHubClient(),
                {
                    owner: 'sourcegraph',
                    repo: 'sourcegraph',
                    branch,
                    tag,
                },
                config.dryRun.tags || false
            )
        },
    },
    {
        id: 'release:stage',
        description: 'Open pull requests and a campaign staging a release',
        run: async config => {
            const { slackAnnounceChannel, dryRun } = config
            const { upcoming: release, previous } = await releaseVersions(config)

            // set up campaign config
            const campaign = campaigns.releaseTrackingCampaign(release.version, await campaigns.sourcegraphCLIConfig())

            // default values
            const notPatchRelease = release.patch === 0
            const versionRegex = '[0-9]+\\.[0-9]+\\.[0-9]+'
            const campaignURL = campaigns.campaignURL(campaign)
            const trackingIssue = await getTrackingIssue(await getAuthenticatedGitHubClient(), release)
            if (!trackingIssue) {
                // Do not block release staging on lack of tracking issue
                console.error(`Tracking issue for version ${release.version} not found - has it been created yet?`)
            }

            // default PR content
            const defaultPRMessage = `release: sourcegraph@${release.version}`
            const prBodyAndDraftState = (
                actionItems: string[],
                customMessage?: string
            ): { draft: boolean; body: string } => {
                const defaultBody = `This pull request is part of the Sourcegraph ${release.version} release.
${customMessage || ''}

* [Release campaign](${campaignURL})
* ${trackingIssue ? `[Tracking issue](${trackingIssue.url})` : 'No tracking issue exists for this release'}`
                if (!actionItems || actionItems.length === 0) {
                    return { draft: false, body: defaultBody }
                }
                return {
                    draft: true, // further actions required before merge
                    body: `${defaultBody}

### :warning: Additional changes required

${actionItems.map(item => `- [ ] ${item}`).join('\n')}

cc @${config.captainGitHubUsername}
`,
                }
            }

            // Render changes
            const createdChanges = await createChangesets({
                requiredCommands: ['comby', sed, 'find', 'go'],
                changes: [
                    {
                        owner: 'sourcegraph',
                        repo: 'sourcegraph',
                        base: 'main',
                        head: `publish-${release.version}`,
                        commitMessage: notPatchRelease
                            ? `draft sourcegraph@${release.version} release`
                            : defaultPRMessage,
                        title: defaultPRMessage,
                        edits: [
                            // Update references to Sourcegraph versions in docs
                            `find . -type f -name '*.md' ! -name 'CHANGELOG.md' -exec ${sed} -i -E 's/sourcegraph\\/server:${versionRegex}/sourcegraph\\/server:${release.version}/g' {} +`,
                            `${sed} -i -E 's/version \`${versionRegex}\`/version \`${release.version}\`/g' doc/index.md`,
                            `${sed} -i -E 's/SOURCEGRAPH_VERSION="v${versionRegex}"/SOURCEGRAPH_VERSION="v${release.version}"/g' doc/admin/install/docker-compose/index.md`,
                            notPatchRelease
                                ? `comby -in-place '{{$previousReleaseRevspec := ":[1]"}} {{$previousReleaseVersion := ":[2]"}} {{$currentReleaseRevspec := ":[3]"}} {{$currentReleaseVersion := ":[4]"}}' '{{$previousReleaseRevspec := ":[3]"}} {{$previousReleaseVersion := ":[4]"}} {{$currentReleaseRevspec := "v${release.version}"}} {{$currentReleaseVersion := "${release.major}.${release.minor}"}}' doc/_resources/templates/document.html`
                                : `comby -in-place 'currentReleaseRevspec := ":[1]"' 'currentReleaseRevspec := "v${release.version}"' doc/_resources/templates/document.html`,

                            // Update references to Sourcegraph deployment versions
                            `comby -in-place 'latestReleaseKubernetesBuild = newBuild(":[1]")' "latestReleaseKubernetesBuild = newBuild(\\"${release.version}\\")" cmd/frontend/internal/app/updatecheck/handler.go`,
                            `comby -in-place 'latestReleaseDockerServerImageBuild = newBuild(":[1]")' "latestReleaseDockerServerImageBuild = newBuild(\\"${release.version}\\")" cmd/frontend/internal/app/updatecheck/handler.go`,
                            `comby -in-place 'latestReleaseDockerComposeOrPureDocker = newBuild(":[1]")' "latestReleaseDockerComposeOrPureDocker = newBuild(\\"${release.version}\\")" cmd/frontend/internal/app/updatecheck/handler.go`,

                            // Add a stub to add upgrade guide entries
                            notPatchRelease
                                ? `${sed} -i -E '/GENERATE UPGRADE GUIDE ON RELEASE/a \\\n\\n## ${previous.major}.${previous.minor} -> ${release.major}.${release.minor}\\n\\nTODO' doc/admin/updates/*.md`
                                : 'echo "Skipping upgrade guide entries"',
                        ],
                        ...prBodyAndDraftState(
                            ((): string[] => {
                                const items: string[] = []
                                if (notPatchRelease) {
                                    items.push('Update the upgrade guides in `doc/admin/updates`')
                                } else {
                                    items.push(
                                        'Update the [CHANGELOG](https://github.com/sourcegraph/sourcegraph/blob/main/CHANGELOG.md) to include all the changes included in this patch',
                                        'If any specific upgrade steps are required, update the upgrade guides in `doc/admin/updates`'
                                    )
                                }
                                items.push(
                                    'Ensure all other pull requests in the campaign have been merged - then run `yarn run release release:finalize` to generate the tags required, re-run Buildkite on this branch, and ensure the build passes before merging this pull request'
                                )
                                return items
                            })()
                        ),
                    },
                    {
                        owner: 'sourcegraph',
                        repo: 'about',
                        base: 'main',
                        head: `publish-${release.version}`,
                        commitMessage: defaultPRMessage,
                        title: defaultPRMessage,
                        edits: [
                            `${sed} -i -E 's/sourcegraph\\/server:${versionRegex}/sourcegraph\\/server:${release.version}/g' 'website/src/components/GetStarted.tsx'`,
                        ],
                        ...prBodyAndDraftState([], 'Note that this PR does *not* include the release blog post.'),
                    },
                    {
                        owner: 'sourcegraph',
                        repo: 'deploy-sourcegraph',
                        base: `${release.major}.${release.minor}`,
                        head: `publish-${release.version}`,
                        commitMessage: defaultPRMessage,
                        title: defaultPRMessage,
                        edits: [`tools/update-docker-tags.sh ${release.version}`],
                        ...prBodyAndDraftState([]),
                    },
                    {
                        owner: 'sourcegraph',
                        repo: 'deploy-sourcegraph-docker',
                        base: `${release.major}.${release.minor}`,
                        head: `publish-${release.version}`,
                        commitMessage: defaultPRMessage,
                        title: defaultPRMessage,
                        edits: [`tools/update-docker-tags.sh ${release.version}`],
                        ...prBodyAndDraftState([
                            `Follow the [release guide](https://github.com/sourcegraph/deploy-sourcegraph-docker/blob/master/RELEASING.md) to complete this PR ${
                                notPatchRelease ? '' : '(note: `pure-docker` release is optional for patch releases)'
                            }`,
                        ]),
                    },
                    {
                        owner: 'sourcegraph',
                        repo: 'deploy-sourcegraph-aws',
                        base: 'master',
                        head: `publish-${release.version}`,
                        commitMessage: defaultPRMessage,
                        title: defaultPRMessage,
                        edits: [
                            `${sed} -i -E 's/export SOURCEGRAPH_VERSION=${versionRegex}/export SOURCEGRAPH_VERSION=${release.version}/g' resources/amazon-linux2.sh`,
                        ],
                        ...prBodyAndDraftState([]),
                    },
                    {
                        owner: 'sourcegraph',
                        repo: 'deploy-sourcegraph-digitalocean',
                        base: 'master',
                        head: `publish-${release.version}`,
                        commitMessage: defaultPRMessage,
                        title: defaultPRMessage,
                        edits: [
                            `${sed} -i -E 's/export SOURCEGRAPH_VERSION=${versionRegex}/export SOURCEGRAPH_VERSION=${release.version}/g' resources/user-data.sh`,
                        ],
                        ...prBodyAndDraftState([]),
                    },
                ],
                dryRun: dryRun.changesets,
            })

            // if changesets were actually published, set up a campaign and post in Slack
            if (!dryRun.changesets) {
                // Create campaign to track changes
                try {
                    console.log(`Creating campaign in ${campaign.cliConfig.SRC_ENDPOINT}`)
                    await campaigns.createCampaign(createdChanges, campaign)
                } catch (error) {
                    console.error(error)
                    console.error('Failed to create campaign for this release, continuing with announcement')
                }

                // Announce release update in Slack
                await postMessage(
                    `:captain: *Sourcegraph ${release.version} release has been staged*

Campaign: ${campaignURL}`,
                    slackAnnounceChannel
                )
            }
        },
    },
    {
        id: 'release:add-to-campaign',
        description: 'Manually add a change to a release campaign',
        argNames: ['changeRepo', 'changeID'],
        // Example: yarn run release release:add-to-campaign sourcegraph/about 1797
        run: async (config, changeRepo, changeID) => {
            const { upcoming: release } = await releaseVersions(config)
            if (!changeRepo || !changeID) {
                throw new Error('Missing parameters (required: version, repo, change ID)')
            }

            const campaign = campaigns.releaseTrackingCampaign(release.version, await campaigns.sourcegraphCLIConfig())
            await campaigns.addToCampaign(
                [
                    {
                        repository: changeRepo,
                        pullRequestNumber: parseInt(changeID, 10),
                    },
                ],
                campaign
            )
            console.log(`Added ${changeRepo}#${changeID} to campaign ${campaigns.campaignURL(campaign)}`)
        },
    },
    {
        id: 'release:finalize',
        description: 'Run final tasks for the sourcegraph/sourcegraph release pull request',
        run: async config => {
            const { upcoming: release } = await releaseVersions(config)

            // Push final tags
            const branch = `${release.major}.${release.minor}`
            const tag = `v${release.version}`
            for (const repo of ['deploy-sourcegraph', 'deploy-sourcegraph-docker']) {
                await createTag(
                    await getAuthenticatedGitHubClient(),
                    {
                        owner: 'sourcegraph',
                        repo,
                        branch,
                        tag,
                    },
                    config.dryRun.tags || false
                )
            }
        },
    },
    {
        id: 'release:close',
        description: 'Mark a release as closed',
        run: async config => {
            const { slackAnnounceChannel } = config
            const { upcoming: release } = await releaseVersions(config)
            const githubClient = await getAuthenticatedGitHubClient()

            // Set up announcement message
            const versionAnchor = release.version.replaceAll('.', '-')
            const campaignURL = campaigns.campaignURL(
                campaigns.releaseTrackingCampaign(release.version, await campaigns.sourcegraphCLIConfig())
            )
            const releaseMessage = `*Sourcegraph ${release.version} has been published*

* Changelog: https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/CHANGELOG.md#${versionAnchor}
* Release campaign: ${campaignURL}`

            // Slack
            await postMessage(`:captain: ${releaseMessage}`, slackAnnounceChannel)
            console.log(`Posted to Slack channel ${slackAnnounceChannel}`)

            // GitHub
            const trackingIssue = await getTrackingIssue(githubClient, release)
            if (!trackingIssue) {
                console.warn(`Could not find tracking issue for release ${release.version} - skipping`)
            } else {
                await githubClient.issues.createComment({
                    owner: trackingIssue.owner,
                    repo: trackingIssue.repo,
                    issue_number: trackingIssue.number,
                    body: `${releaseMessage}

@${config.captainGitHubUsername}: Please complete the post-release steps before closing this issue.`,
                })
            }
        },
    },
    {
        id: '_test:google-calendar',
        description: 'Test Google Calendar integration',
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
        description: 'Test Slack integration',
        argNames: ['channel', 'message'],
        run: async (_config, channel, message) => {
            await postMessage(message, channel)
        },
    },
    {
        id: '_test:campaign-create-from-changes',
        description: 'Test campaigns integration',
        argNames: ['campaignConfigJSON'],
        // Example: yarn run release _test:campaign-create-from-changes "$(cat ./.secrets/import.json)"
        run: async (_config, campaignConfigJSON) => {
            const campaignConfig = JSON.parse(campaignConfigJSON) as {
                changes: CreatedChangeset[]
                name: string
                description: string
            }

            // set up src-cli
            await commandExists('src')
            const campaign = {
                name: campaignConfig.name,
                description: campaignConfig.description,
                namespace: 'sourcegraph',
                cliConfig: await campaigns.sourcegraphCLIConfig(),
            }

            await campaigns.createCampaign(campaignConfig.changes, campaign)
            console.log(`Created campaign ${campaigns.campaignURL(campaign)}`)
        },
    },
    {
        id: '_test:config',
        description: 'Test release configuration loading',
        run: config => {
            console.log(JSON.stringify(config, null, '  '))
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
    const config = loadConfig()
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
