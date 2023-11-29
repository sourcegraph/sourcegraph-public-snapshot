import { readFileSync, rmdirSync, writeFileSync } from 'fs'
import * as path from 'path'
import { exit } from 'process'

import chalk from 'chalk'
import commandExists from 'command-exists'
import { addMinutes } from 'date-fns'
import execa from 'execa'
import { DateTime } from 'luxon'
import { SemVer } from 'semver'
import semver from 'semver/preload'

import * as batchChanges from './batchChanges'
import * as changelog from './changelog'
import {
    activateRelease,
    addScheduledRelease,
    loadReleaseConfig,
    newReleaseFromInput,
    type ReleaseConfig,
    getActiveRelease,
    removeScheduledRelease,
    saveReleaseConfig,
    getReleaseDefinition,
    deactivateAllReleases,
    setSrcCliVersion,
    newRelease,
    setGoogleExecutorVersion,
    setAWSExecutorVersion,
} from './config'
import { getCandidateTags, getPreviousVersion } from './git'
import {
    cloneRepo,
    closeTrackingIssue,
    commentOnIssue,
    createChangesets,
    type CreatedChangeset,
    createLatestRelease,
    createTag,
    type Edit,
    ensureTrackingIssues,
    getAuthenticatedGitHubClient,
    getTrackingIssue,
    IssueLabel,
    localSourcegraphRepo,
    queryIssues,
    releaseBlockerLabel,
    releaseName,
} from './github'
import { calendarTime, ensureEvent, type EventOptions, getClient } from './google-calendar'
import { postMessage, slackURL } from './slack'
import {
    bakeAWSExecutorsSteps,
    bakeGoogleExecutorsSteps,
    bakeSrcCliSteps,
    batchChangesInAppChangelog,
    combyReplace,
    indexerUpdate,
} from './static-updates'
import {
    backportStatus,
    cacheFolder,
    changelogURL,
    ensureDocker,
    ensureReleaseBranchUpToDate,
    ensureSrcCliEndpoint,
    ensureSrcCliUpToDate,
    formatDate,
    getAllUpgradeGuides,
    getLatestSrcCliGithubRelease,
    getLatestTag,
    getReleaseBlockers,
    nextSrcCliVersionInputWithAutodetect,
    nextGoogleExecutorVersionInputWithAutodetect,
    nextAWSExecutorVersionInputWithAutodetect,
    pullRequestBody,
    releaseBlockerUri,
    retryInput,
    timezoneLink,
    updateUpgradeGuides,
    validateNoOpenBackports,
    validateNoReleaseBlockers,
    verifyWithInput,
    type ReleaseTag,
    updateMigratorBazelOuts,
    getContainerRegistryCredential,
} from './util'

const sed = process.platform === 'linux' ? 'sed' : 'gsed'

export type StepID =
    | 'help'
    // release tracking
    | 'tracking:timeline'
    | 'tracking:issues'
    // branch cut
    | 'changelog:cut'
    | 'release:branch-cut'
    // release
    | 'release:status'
    | 'release:backport-status'
    | 'release:create-candidate'
    | 'release:promote-candidate'
    | 'release:check-candidate'
    | 'release:stage'
    | 'release:add-to-batch-change'
    | 'release:finalize'
    | 'release:announce'
    | 'release:close'
    | 'release:bake-content'
    | 'release:prepare'
    | 'release:remove'
    | 'release:activate-release'
    | 'release:deactivate-release'
    // src-cli and executors
    | 'release:create-tags'
    | 'release:verify-releases'
    // util
    | 'util:clear-cache'
    | 'util:previous-version'
    // testing
    | '_test:google-calendar'
    | '_test:slack'
    | '_test:batchchange-create-from-changes'
    | '_test:config'
    | '_test:dockerensure'
    | '_test:srccliensure'
    | '_test:patch-dates'
    | '_test:release-guide-content'
    | '_test:release-guide-update'

/**
 * Runs given release step with the provided configuration and arguments.
 */
export async function runStep(config: ReleaseConfig, step: StepID, ...args: string[]): Promise<void> {
    if (!steps.map(({ id }) => id as string).includes(step)) {
        throw new Error(`Unrecognized step ${JSON.stringify(step)}`)
    }
    await Promise.all(
        steps
            .filter(({ id }) => id === step)
            .map(async step => {
                if (step.run) {
                    await step.run(config, ...args)
                }
            })
    )
}

interface Step {
    id: StepID
    description: string
    run?:
        | ((config: ReleaseConfig, ...args: string[]) => Promise<void>)
        | ((config: ReleaseConfig, ...args: string[]) => void)
    argNames?: string[]
}

const steps: Step[] = [
    {
        id: 'help',
        description: 'Output help text about this tool',
        argNames: ['all'],
        run: (_config, all) => {
            console.error('Sourcegraph release tool - https://handbook.sourcegraph.com/engineering/releases')
            console.error('\nUSAGE\n')
            console.error('\tpnpm run release <step>')
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
        id: 'tracking:timeline',
        description: 'Generate a set of Google Calendar events for a MAJOR.MINOR release',
        run: async config => {
            const next = await getReleaseDefinition(config)
            const name = releaseName(new SemVer(next.current))
            const events: EventOptions[] = [
                {
                    title: `Security Team to Review Release Container Image Scans ${name}`,
                    description: '(This is not an actual event to attend, just a calendar marker.)',
                    anyoneCanAddSelf: true,
                    attendees: [config.metadata.teamEmail],
                    transparency: 'transparent',
                    ...calendarTime(next.securityApprovalDate),
                },
                {
                    title: `Cut Sourcegraph ${name}`,
                    description: '(This is not an actual event to attend, just a calendar marker.)',
                    anyoneCanAddSelf: true,
                    attendees: [config.metadata.teamEmail],
                    transparency: 'transparent',
                    ...calendarTime(next.codeFreezeDate),
                },
                {
                    title: `Release Sourcegraph ${name}`,
                    description: '(This is not an actual event to attend, just a calendar marker.)',
                    anyoneCanAddSelf: true,
                    attendees: [config.metadata.teamEmail],
                    transparency: 'transparent',
                    ...calendarTime(next.releaseDate),
                },
            ]

            if (next.patches) {
                for (let i = 0; i < next.patches.length; i++) {
                    events.push({
                        title: `Scheduled Patch #${i + 1} Sourcegraph ${name}`,
                        description: '(This is not an actual event to attend, just a calendar marker.)',
                        anyoneCanAddSelf: true,
                        attendees: [config.metadata.teamEmail],
                        transparency: 'transparent',
                        ...calendarTime(next.patches[i]),
                    })
                }
            }

            if (!config.dryRun.calendar) {
                const googleCalendar = await getClient()
                for (const event of events) {
                    console.log(`Create calendar event: ${event.title}: ${event.startDateTime || 'undefined'}`)
                    await ensureEvent(event, googleCalendar)
                }
            } else {
                console.log('dryRun.calendar=true, skipping calendar event creation', events)
            }
        },
    },
    {
        id: 'tracking:issues',
        description: 'Generate GitHub tracking issues for a release',
        run: async (config: ReleaseConfig) => {
            const next = await getReleaseDefinition(config)
            const version = new SemVer(next.current)
            const date = new Date(next.releaseDate)

            // Create issue
            const trackingIssues = await ensureTrackingIssues({
                version,
                assignees: [next.captainGitHubUsername],
                releaseDate: date,
                securityReviewDate: new Date(next.securityApprovalDate),
                codeFreezeDate: new Date(next.codeFreezeDate),
                dryRun: config.dryRun.trackingIssues || false,
            })
            console.log('Rendered tracking issues', trackingIssues)

            // If at least one issue was created, post to Slack
            if (trackingIssues.find(({ created }) => created)) {
                const name = releaseName(version)
                const releaseDateString = slackURL(formatDate(date), timezoneLink(date, `${name} release`))
                let annoncement = `:mega: *${name} release*

:captain: Release captain: @${next.captainSlackUsername}
:spiral_calendar_pad: Scheduled for: ${releaseDateString}
:pencil: Tracking issues:
${trackingIssues.map(index => `- ${slackURL(index.title, index.url)}`).join('\n')}`
                if (version.patch !== 0) {
                    const patchRequestTemplate = `https://github.com/sourcegraph/sourcegraph/issues/new?assignees=&labels=team%2Fdistribution&template=request_patch_release.md&title=${version.version}%3A+`
                    annoncement += `\n\nIf you have changes that should go into this patch release, ${slackURL(
                        'please *file a patch request issue*',
                        patchRequestTemplate
                    )}, or it will not be included.`
                }
                if (!config.dryRun.slack) {
                    await postMessage(annoncement, config.metadata.slackAnnounceChannel)
                    console.log(`Posted to Slack channel ${config.metadata.slackAnnounceChannel}`)
                }
            } else {
                console.log('No tracking issues were created, skipping Slack announcement')
            }
        },
    },
    {
        id: 'changelog:cut',
        description: 'Generate pull requests to perform a changelog cut for branch cut',
        argNames: ['changelogFile'],
        run: async (config, changelogFile = 'CHANGELOG.md') => {
            const upcoming = await getActiveRelease(config)
            const srcCliNext = await nextSrcCliVersionInputWithAutodetect(config)

            const commitMessage = `changelog: cut sourcegraph@${upcoming.version.version}`
            const prBody = commitMessage + '\n\n ## Test plan\n\nN/A'
            const pullRequest = await createChangesets({
                requiredCommands: [],
                changes: [
                    {
                        owner: 'sourcegraph',
                        repo: 'sourcegraph',
                        base: 'main',
                        head: `changelog-${upcoming.version.version}`,
                        title: commitMessage,
                        commitMessage,
                        body: prBody,
                        edits: [
                            (directory: string) => {
                                console.log(`Updating '${changelogFile} for ${upcoming.version.format()}'`)
                                const changelogPath = path.join(directory, changelogFile)
                                let changelogContents = readFileSync(changelogPath).toString()

                                // Convert 'unreleased' to a release
                                const releaseHeader = `## ${upcoming.version.format()}`
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
                        ],
                    },
                    {
                        owner: 'sourcegraph',
                        repo: 'deploy-sourcegraph-helm',
                        base: 'main',
                        head: `changelog-${upcoming.version.version}`,
                        title: commitMessage,
                        commitMessage,
                        body: prBody,
                        edits: [
                            (directory: string) => {
                                console.log(`Updating '${changelogFile} for ${upcoming.version.format()}'`)
                                const changelogPath = path.join(directory, 'charts', 'sourcegraph', changelogFile)
                                let changelogContents = readFileSync(changelogPath).toString()

                                // Convert 'unreleased' to a release
                                const releaseHeader = `## ${upcoming.version.format()}`
                                const releaseUpdate =
                                    releaseHeader + `\n\n- Sourcegraph ${upcoming.version.format()} is now available\n`
                                const unreleasedHeader = '## Unreleased\n'
                                changelogContents = changelogContents.replace(unreleasedHeader, releaseUpdate)

                                // Add a blank changelog template for the next release
                                changelogContents = changelogContents.replace(
                                    changelog.divider,
                                    changelog.simpleReleaseTemplate
                                )

                                // Update changelog
                                writeFileSync(changelogPath, changelogContents)
                            },
                        ],
                    },
                    {
                        owner: 'sourcegraph',
                        repo: 'src-cli',
                        base: 'main',
                        head: `changelog-${srcCliNext.version}`,
                        title: commitMessage,
                        body: prBody,
                        commitMessage,
                        edits: [
                            (directory: string) => {
                                console.log(`Updating '${changelogFile} for ${srcCliNext.format()}'`)
                                const changelogPath = path.join(directory, changelogFile)
                                let changelogContents = readFileSync(changelogPath).toString()

                                // Convert 'unreleased' to a release
                                const unreleasedHeader = '## Unreleased'
                                const unreleasedSection = `${unreleasedHeader}\n\n### Added\n\n### Changed\n\n### Fixed\n\n### Removed\n\n`
                                const newSection = `${unreleasedSection}## ${srcCliNext.format()}`
                                changelogContents = changelogContents.replace(unreleasedHeader, newSection)

                                // Update changelog
                                writeFileSync(changelogPath, changelogContents)
                            },
                        ],
                    },
                ],
                dryRun: config.dryRun.changesets,
            })
            const changeLogPrUrl = pullRequest[0].pullRequestURL
            console.log(
                `\nPlease review the changelog PR at ${changeLogPrUrl}, and merge manually when checks have passed.`
            )
        },
    },
    {
        id: 'release:branch-cut',
        description: 'Create release branch',
        run: async config => {
            const release = await getActiveRelease(config)
            const client = await getAuthenticatedGitHubClient()
            let message: string
            // notify cs team on patch release cut
            if (release.version.patch !== 0) {
                message = `:mega: *${release.version.version}* branch has been cut cc: @cs\nIf you need to include changes in the release, see instructions on how to backport: https://handbook.sourcegraph.com/departments/engineering/dev/tools/backport/#how-should-i-use-the-backporting-tool.`
            } else {
                message = `:mega: *${release.version.version}* branch has been cut.`
            }
            try {
                // Create and push new release branch from changelog commit
                await execa('git', ['branch', release.branch])
                await execa('git', ['push', 'origin', release.branch])
                await postMessage(message, config.metadata.slackAnnounceChannel)
                console.log(`To check the status of the branch, run:\nsg ci status -branch ${release.branch} --wait\n`)
            } catch (error) {
                console.error('Failed to create release branch', error)
            }

            if (release.version.patch === 0) {
                // create backport label for major / minor versions
                const params = {
                    owner: 'sourcegraph',
                    repo: 'sourcegraph',
                }
                const labelName = `backport ${release.version.major}.${release.version.minor}`
                const labelExists = await client.issues
                    .getLabel({ name: labelName, ...params })
                    .then(resp => resp.status === 200)
                    .catch(() => false)
                if (!labelExists) {
                    console.log(await client.issues.createLabel({ name: labelName, color: 'e69138', ...params }))
                    console.log(`Label ${labelName} created`)
                } else {
                    console.log(`label ${labelName} already exists`)
                }
            }
        },
    },
    {
        id: 'release:status',
        description: 'Post a message in Slack summarizing the progress of a release',
        run: async config => {
            const githubClient = await getAuthenticatedGitHubClient()
            const release = await getActiveRelease(config)

            const trackingIssue = await getTrackingIssue(githubClient, release.version)
            if (!trackingIssue) {
                throw new Error(
                    `Tracking issue for version ${release.version.version} not found - has it been created yet?`
                )
            }
            const latestTag = (await getLatestTag('sourcegraph', 'sourcegraph')).toString()
            const latestBuildURL = `https://buildkite.com/sourcegraph/sourcegraph/builds?branch=${latestTag}`
            const latestBuildMessage = `Latest release build: ${latestTag}. See the build status on <${latestBuildURL}|Buildkite>`

            const blockingIssues = await getReleaseBlockers(githubClient)
            const blockingMessage =
                blockingIssues.length === 0
                    ? 'There are no release-blocking issues'
                    : `There ${
                          blockingIssues.length === 1
                              ? 'is 1 release-blocking issue'
                              : `are ${blockingIssues.length} release-blocking issues`
                      }`

            const message = `:mega: *${release.version.version} Release Status Update*

* Tracking issue: ${trackingIssue.url}
* ${blockingMessage}: ${releaseBlockerUri()}
* ${latestBuildMessage}`
            if (!config.dryRun.slack) {
                await postMessage(message, config.metadata.slackAnnounceChannel)
            } else {
                console.log(chalk.green('Dry run: ' + message))
            }
        },
    },
    {
        id: 'release:backport-status',
        description: 'Check for backport issues on the currently active release',
        run: async config => {
            const release = await getActiveRelease(config)
            getAuthenticatedGitHubClient()
                .then(client => backportStatus(client, release.version))
                .then(str => console.log(str))
                .catch(error => error)
        },
    },
    {
        id: 'release:create-candidate',
        description: 'Generate the Nth release candidate. Set <candidate> to "final" to generate a final release',
        run: async config => {
            const release = await getActiveRelease(config)
            ensureReleaseBranchUpToDate(release.branch)

            const owner = 'sourcegraph'
            const repo = 'sourcegraph'

            try {
                const client = await getAuthenticatedGitHubClient()
                const { workdir } = await cloneRepo(client, owner, repo, {
                    revision: release.branch,
                    revisionMustExist: true,
                })

                const tags = getCandidateTags(workdir, release.version.version)
                let nextCandidate = 1
                for (const tag of tags) {
                    const lastNum = tag.match('.*-rc\\.(\\d+)')
                    if (!lastNum || lastNum.length === 0) {
                        break
                    }
                    const num = parseInt(lastNum[1], 10)
                    if (num >= nextCandidate) {
                        nextCandidate = num + 1
                    }
                }
                const tag = `v${release.version.version}-rc.${nextCandidate}`

                console.log(`Detected next candidate: ${nextCandidate}, attempting to create tag: ${tag}`)
                await createTag(
                    client,
                    workdir,
                    {
                        owner,
                        repo,
                        branch: release.branch,
                        tag,
                    },
                    config.dryRun.tags || false
                )
                console.log(`To check the status of the build, run:\nsg ci status -branch ${tag} --wait\n`)
            } catch (error) {
                console.error('Failed to create tag', error)
            }
        },
    },
    {
        id: 'release:promote-candidate',
        description:
            'Promote a release candidate to release build. Specify the full candidate tag to promote the tagged commit to release.',
        argNames: ['candidate'],
        run: async (config, candidate) => {
            const release = await getActiveRelease(config)
            ensureReleaseBranchUpToDate(release.branch)

            const client = await getAuthenticatedGitHubClient()
            await validateNoReleaseBlockers(client)
            await validateNoOpenBackports(client, release.version)

            const warnMsg =
                'Verify the provided tag is correct to promote to release. Note: it is very unusual to require a non-standard tag to promote to release, proceed with caution.'
            const exampleTag = `v${release.version.version}-rc.1`
            if (!candidate) {
                throw new Error(
                    `Candidate tag is a required argument. This should be the git tag of the commit to promote to release (ex.${exampleTag}`
                )
            } else if (!candidate.match('v\\d\\.\\d(?:\\.\\d)?-rc\\.\\d')) {
                await verifyWithInput(
                    `Warning!\nCandidate tag: ${candidate} does not match the standard convention (ex. ${exampleTag}). ${warnMsg}`
                )
            } else if (!candidate.match(`${release.version.version}-rc\\.\\d`)) {
                await verifyWithInput(
                    `Warning!\nCandidate tag: ${candidate} does not match the expected version ${release.version.version} (ex. ${exampleTag}). ${warnMsg}`
                )
            }

            const owner = 'sourcegraph'
            const repo = 'sourcegraph'

            const releaseTag = `v${release.version.version}`

            try {
                // passing the tag as branch so that only the specified tag is shallow cloned
                const { workdir } = await cloneRepo(client, owner, repo, {
                    revision: candidate,
                    revisionMustExist: true,
                })

                execa.sync('git', ['fetch', '--tags'], { stdio: 'inherit', cwd: workdir })
                await createTag(
                    client,
                    workdir,
                    {
                        owner,
                        repo,
                        branch: candidate,
                        tag: releaseTag,
                    },
                    config.dryRun.tags || false
                )
                console.log(`To check the status of the build, run:\nsg ci status -branch ${releaseTag} --wait\n`)
            } catch (error) {
                console.error('Failed to create tag', error)
            }
        },
    },
    {
        id: 'release:check-candidate',
        description: 'Check release candidates.',
        argNames: ['version'],
        run: async (config, version) => {
            if (!version) {
                const release = await getActiveRelease(config)
                version = release.version.version
            }
            const tags = getCandidateTags(localSourcegraphRepo, version)
            if (tags.length > 0) {
                console.log(`Release candidate tags for version: ${chalk.blue(version)}`)
                for (const tag of tags) {
                    console.log(tag)
                }
                console.log('To check the status of the build, run:\nsg ci status -branch tag\n')
            } else {
                console.log(chalk.yellow('No candidates found!'))
            }
        },
    },
    {
        id: 'release:stage',
        description: 'Open pull requests and a batch change staging a release',
        run: async config => {
            const release = await getActiveRelease(config)
            // ensure docker is running for 'batch changes'
            try {
                await ensureDocker()
            } catch (error) {
                console.log(error)
                console.log('Docker required for batch changes')
                process.exit(1)
            }
            // ensure $SRC_ENDPOINT is set
            ensureSrcCliEndpoint()
            // ensure src-cli is up to date
            await ensureSrcCliUpToDate()
            // set up batch change config
            const batchChange = batchChanges.releaseTrackingBatchChange(
                release.version.version,
                await batchChanges.sourcegraphCLIConfig()
            )

            await validateNoReleaseBlockers(await getAuthenticatedGitHubClient())

            // default values
            const notPatchRelease = release.version.patch === 0
            const versionRegex = '[0-9]+\\.[0-9]+\\.[0-9]+'
            const batchChangeURL = batchChanges.batchChangeURL(batchChange)
            const trackingIssue = await getTrackingIssue(await getAuthenticatedGitHubClient(), release.version)
            if (!trackingIssue) {
                throw new Error(
                    `Tracking issue for version ${release.version.version} not found - has it been created yet?`
                )
            }

            // default PR content
            const defaultPRMessage = `release: sourcegraph@${release.version.version}`
            const prBodyAndDraftState = (
                actionItems: string[],
                customMessage?: string
            ): { draft: boolean; body: string } => {
                const defaultBody = `This pull request is part of the Sourcegraph ${release.version.version} release.
${customMessage || ''}

* [Release batch change](${batchChangeURL})
* ${trackingIssue ? `[Tracking issue](${trackingIssue.url})` : 'No tracking issue exists for this release'}

### Test plan

CI checks in this repository should pass, and a manual review should confirm if the generated changes are correct.`

                if (!actionItems || actionItems.length === 0) {
                    return { draft: false, body: defaultBody }
                }
                return {
                    draft: true, // further actions required before merge
                    body: `${defaultBody}

### :warning: Additional changes required

These steps must be completed before this PR can be merged, unless otherwise stated. Push any required changes directly to this PR branch.

${actionItems.map(item => `- [ ] ${item}`).join('\n')}

cc @${release.captainGitHubUsername}

`,
                }
            }

            const { username: dockerUsername, password: dockerPassword } = await getContainerRegistryCredential(
                'index.docker.io'
            )

            // Render changes
            const createdChanges = await createChangesets({
                requiredCommands: ['comby', sed, 'find', 'go', 'src', 'sg'],
                changes: [
                    {
                        owner: 'sourcegraph',
                        repo: 'sourcegraph',
                        base: 'main',
                        head: `publish-${release.version.version}`,
                        commitMessage: notPatchRelease
                            ? `draft sourcegraph@${release.version.version} release`
                            : defaultPRMessage,
                        title: defaultPRMessage,
                        edits: [
                            // Update references to Sourcegraph versions in docs
                            `${sed} -i -E 's/version \`${versionRegex}\`/version \`${release.version.version}\`/g' doc/index.md`,
                            // Update sourcegraph/server:VERSION everywhere except changelog
                            `find . -type f -name '*.md' ! -name 'CHANGELOG.md' -exec ${sed} -i -E 's/sourcegraph\\/server:${versionRegex}/sourcegraph\\/server:${release.version.version}/g' {} +`,
                            // Update Sourcegraph versions in installation guides
                            `find ./doc/admin/deploy/ -type f -name '*.md' -exec ${sed} -i -E 's/SOURCEGRAPH_VERSION="v${versionRegex}"/SOURCEGRAPH_VERSION="v${release.version.version}"/g' {} +`,
                            `find ./doc/admin/deploy/ -type f -name '*.md' -exec ${sed} -i -E 's/--version ${versionRegex}/--version ${release.version.version}/g' {} +`,
                            `${sed} -i -E 's/${versionRegex}/${release.version.version}/g' ./doc/admin/executors/deploy_executors_kubernetes.md`,
                            // Update fork variables in installation guides
                            `find ./doc/admin/deploy/ -type f -name '*.md' -exec ${sed} -i -E "s/DEPLOY_SOURCEGRAPH_DOCKER_FORK_REVISION='v${versionRegex}'/DEPLOY_SOURCEGRAPH_DOCKER_FORK_REVISION='v${release.version.version}'/g" {} +`,

                            notPatchRelease
                                ? `comby -in-place '{{$previousReleaseRevspec := ":[1]"}} {{$previousReleaseVersion := ":[2]"}} {{$currentReleaseRevspec := ":[3]"}} {{$currentReleaseVersion := ":[4]"}}' '{{$previousReleaseRevspec := ":[3]"}} {{$previousReleaseVersion := ":[4]"}} {{$currentReleaseRevspec := "v${release.version.version}"}} {{$currentReleaseVersion := "${release.version.major}.${release.version.minor}"}}' doc/_resources/templates/document.html`
                                : `comby -in-place 'currentReleaseRevspec := ":[1]"' 'currentReleaseRevspec := "v${release.version.version}"' doc/_resources/templates/document.html`,

                            // Update references to Sourcegraph deployment versions
                            `comby -in-place 'latestReleaseKubernetesBuild = newPingResponse(":[1]")' "latestReleaseKubernetesBuild = newPingResponse(\\"${release.version.version}\\")" internal/updatecheck/handler.go`,
                            `comby -in-place 'latestReleaseDockerServerImageBuild = newPingResponse(":[1]")' "latestReleaseDockerServerImageBuild = newPingResponse(\\"${release.version.version}\\")" internal/updatecheck/handler.go`,
                            `comby -in-place 'latestReleaseDockerComposeOrPureDocker = newPingResponse(":[1]")' "latestReleaseDockerComposeOrPureDocker = newPingResponse(\\"${release.version.version}\\")" internal/updatecheck/handler.go`,

                            // Support current release as the "previous release" going forward
                            notPatchRelease
                                ? `comby -in-place 'const minimumUpgradeableVersion = ":[1]"' 'const minimumUpgradeableVersion = "${release.version.version}"' dev/ci/internal/ci/*.go`
                                : 'echo "Skipping minimumUpgradeableVersion bump on patch release"',
                            updateUpgradeGuides(release.previous.version, release.version.version),
                            updateMigratorBazelOuts(release.version.version),
                        ],
                        ...prBodyAndDraftState(
                            ((): string[] => {
                                const items: string[] = []
                                items.push(
                                    'Ensure all other pull requests in the batch change have been merged',
                                    'Run `pnpm run release release:finalize` to generate the tags required. CI will not pass until this command is run.',
                                    'Re-run the build on this branch (using either the command `sg ci build` or the Buildkite UI) and merge when the build passes.'
                                )
                                return items
                            })()
                        ),
                    },
                    {
                        owner: 'sourcegraph',
                        repo: 'about',
                        base: 'main',
                        head: `publish-${release.version.version}`,
                        commitMessage: defaultPRMessage,
                        title: defaultPRMessage,
                        edits: [
                            // Update sourcegraph/server:VERSION in all tsx files
                            `find . -type f -name '*.tsx' -exec ${sed} -i -E 's/sourcegraph\\/server:${versionRegex}/sourcegraph\\/server:${release.version.version}/g' {} +`,
                        ],
                        ...prBodyAndDraftState(
                            [],
                            notPatchRelease ? 'Note that this PR does *not* include the release blog post.' : undefined
                        ),
                    },
                    {
                        owner: 'sourcegraph',
                        repo: 'deploy-sourcegraph',
                        base: release.branch,
                        head: `publish-${release.version.version}`,
                        commitMessage: defaultPRMessage,
                        title: defaultPRMessage,
                        edits: [`tools/update-docker-tags.sh ${release.version.version}`],
                        ...prBodyAndDraftState([]),
                    },
                    {
                        owner: 'sourcegraph',
                        repo: 'deploy-sourcegraph-k8s',
                        base: release.branch,
                        head: `publish-${release.version.version}`,
                        commitMessage: defaultPRMessage,
                        title: defaultPRMessage,
                        edits: [
                            `sg ops update-images -cr-username ${dockerUsername} -cr-password ${dockerPassword} -pin-tag ${release.version.version} base/`,
                            `sg ops update-images -cr-username ${dockerUsername} -cr-password ${dockerPassword} -pin-tag ${release.version.version} components/executors/`,
                            `sg ops update-images -cr-username ${dockerUsername} -cr-password ${dockerPassword} -pin-tag ${release.version.version} components/utils/migrator`,
                        ],
                        ...prBodyAndDraftState([]),
                    },
                    {
                        owner: 'sourcegraph',
                        repo: 'deploy-sourcegraph-docker',
                        base: release.branch,
                        head: `publish-${release.version.version}`,
                        commitMessage: defaultPRMessage,
                        title: defaultPRMessage,
                        edits: [`tools/update-docker-tags.sh ${release.version.version}`],
                        ...prBodyAndDraftState([]),
                    },
                    {
                        owner: 'sourcegraph',
                        repo: 'deploy-sourcegraph-docker-customer-replica-1',
                        base: release.branch,
                        head: `publish-${release.version.version}`,
                        commitMessage: defaultPRMessage,
                        title: defaultPRMessage,
                        edits: [`tools/update-docker-tags.sh ${release.version.version}`],
                        ...prBodyAndDraftState([]),
                    },
                    {
                        owner: 'sourcegraph',
                        repo: 'deploy-sourcegraph-aws',
                        base: 'master',
                        head: `publish-${release.version.version}`,
                        commitMessage: defaultPRMessage,
                        title: defaultPRMessage,
                        edits: [
                            `${sed} -i -E 's/export SOURCEGRAPH_VERSION=${versionRegex}/export SOURCEGRAPH_VERSION=${release.version.version}/g' resources/amazon-linux2.sh`,
                        ],
                        ...prBodyAndDraftState([]),
                    },
                    {
                        owner: 'sourcegraph',
                        repo: 'deploy-sourcegraph-digitalocean',
                        base: 'master',
                        head: `publish-${release.version.version}`,
                        commitMessage: defaultPRMessage,
                        title: defaultPRMessage,
                        edits: [
                            `${sed} -i -E 's/export SOURCEGRAPH_VERSION=${versionRegex}/export SOURCEGRAPH_VERSION=${release.version.version}/g' resources/user-data.sh`,
                        ],
                        ...prBodyAndDraftState([]),
                    },
                    {
                        owner: 'sourcegraph',
                        repo: 'deploy-sourcegraph-helm',
                        base: `release/${release.branch}`,
                        head: `publish-${release.version.version}`,
                        commitMessage: defaultPRMessage,
                        title: defaultPRMessage,
                        edits: [
                            `for i in charts/{sourcegraph,sourcegraph-executor/{dind,k8s},sourcegraph-migrator}; do sg ops update-images -cr-username ${dockerUsername} -cr-password ${dockerPassword} -kind helm -pin-tag ${release.version.version} $i/.; done`,
                            `find charts -name Chart.yaml | xargs ${sed} -i 's/appVersion:.*/appVersion: "${release.version.version}"/g'`,
                            `find charts -name Chart.yaml | xargs ${sed} -i 's/version:.*/version: "${release.version.version}"/g'`,
                            './scripts/helm-docs.sh',
                        ],
                        ...prBodyAndDraftState([]),
                    },
                ],
                dryRun: config.dryRun.changesets,
            })

            // if changesets were actually published, set up a batch change and post in Slack
            if (!config.dryRun.changesets) {
                // Create batch change to track changes
                try {
                    console.log(`Creating batch change in ${batchChange.cliConfig.SRC_ENDPOINT}`)
                    await batchChanges.createBatchChange(
                        createdChanges,
                        batchChange,
                        `Track publishing of sourcegraph v${release.version.version}: ${trackingIssue?.url}`
                    )
                } catch (error) {
                    console.error(error)
                    console.error('Failed to create batch change for this release, continuing with announcement')
                }

                // Announce release update in Slack
                if (!config.dryRun.slack) {
                    await postMessage(
                        `:captain: *Sourcegraph ${release.version.version} has been staged.*

Batch change: ${batchChangeURL}`,
                        config.metadata.slackAnnounceChannel
                    )
                }
            }
        },
    },
    {
        id: 'release:add-to-batch-change',
        description: 'Manually add a change to a release batch change',
        argNames: ['changeRepo', 'changeID'],
        // Example: pnpm run release release:add-to-batch-change sourcegraph/about 1797
        run: async (config, changeRepo, changeID) => {
            const release = await getActiveRelease(config)
            if (!changeRepo || !changeID) {
                throw new Error('Missing parameters (required: version, repo, change ID)')
            }

            const batchChange = batchChanges.releaseTrackingBatchChange(
                release.version.version,
                await batchChanges.sourcegraphCLIConfig()
            )
            await batchChanges.addToBatchChange(
                [
                    {
                        repository: changeRepo,
                        pullRequestNumber: parseInt(changeID, 10),
                    },
                ],
                batchChange
            )
            console.log(`Added ${changeRepo}#${changeID} to batch change ${batchChanges.batchChangeURL(batchChange)}`)
        },
    },
    {
        id: 'release:finalize',
        description: 'Run final tasks for sourcegraph/sourcegraph release pull requests',
        run: async config => {
            const release = await getActiveRelease(config)
            let failed = false

            const defaultBranchPattern = `${release.branch}`
            const defaultTagPattern = `v${release.version.version}`
            const defaults = { branchPattern: defaultBranchPattern, tagPattern: defaultTagPattern }
            const repoConfigs = [
                { repo: 'deploy-sourcegraph', ...defaults },
                { repo: 'deploy-sourcegraph-docker', ...defaults },
                { repo: 'deploy-sourcegraph-docker-customer-replica-1', ...defaults },
                { repo: 'deploy-sourcegraph-k8s', ...defaults },
                {
                    repo: 'deploy-sourcegraph-helm',
                    branchPattern: `release/${release.branch}`,
                    tagPattern: `sourcegraph-${release.version.version}`,
                },
            ]

            const owner = 'sourcegraph'
            // Push final tags
            for (const repoConfig of repoConfigs) {
                try {
                    const client = await getAuthenticatedGitHubClient()
                    const { workdir } = await cloneRepo(client, owner, repoConfig.repo, {
                        revision: repoConfig.branchPattern,
                        revisionMustExist: true,
                    })
                    await createTag(
                        client,
                        workdir,
                        {
                            owner,
                            repo: repoConfig.repo,
                            branch: repoConfig.branchPattern,
                            tag: repoConfig.tagPattern,
                        },
                        config.dryRun.tags || false
                    )
                } catch (error) {
                    console.error(error)
                    console.error(
                        `Failed to create tag ${repoConfig.tagPattern} on ${repoConfig.repo}@${release.branch}`
                    )
                    failed = true
                }
            }

            if (failed) {
                throw new Error('Error occured applying some changes - please check log output')
            }
        },
    },
    {
        id: 'release:announce',
        description: 'Announce a release as live',
        run: async config => {
            const release = await getActiveRelease(config)
            const githubClient = await getAuthenticatedGitHubClient()

            // Create final GitHub release
            let githubRelease = ''
            try {
                githubRelease = await createLatestRelease(
                    githubClient,
                    {
                        owner: 'sourcegraph',
                        repo: 'sourcegraph',
                        release: release.version,
                    },
                    config.dryRun.tags
                )
            } catch (error) {
                console.error('Failed to generate GitHub release:', error)
                // Do not block process
            }

            // Set up announcement message
            const batchChangeURL = batchChanges.batchChangeURL(
                batchChanges.releaseTrackingBatchChange(
                    release.version.version,
                    await batchChanges.sourcegraphCLIConfig()
                )
            )
            const releaseMessage = `*Sourcegraph ${release.version.version} has been published*

* Changelog: ${changelogURL(release.version.format())}
* GitHub release: ${githubRelease || 'No release generated'}
* Release batch change: ${batchChangeURL}`

            // Slack
            const slackMessage = `:captain: ${releaseMessage}`
            if (!config.dryRun.slack) {
                await postMessage(slackMessage, config.metadata.slackAnnounceChannel)
                console.log(`Posted to Slack channel ${config.metadata.slackAnnounceChannel}`)
            } else {
                console.log(
                    `dryRun enabled, skipping Slack post to ${config.metadata.slackAnnounceChannel}: ${slackMessage}`
                )
            }

            // GitHub tracking issues
            const trackingIssue = await getTrackingIssue(githubClient, release.version)
            if (!trackingIssue) {
                console.warn(`Could not find tracking issue for release ${release.version.version} - skipping`)
            } else {
                // Note patch release requests if there are any outstanding
                let comment = `${releaseMessage}

@${release.captainGitHubUsername}: Please complete the post-release steps before closing this issue.`
                const patchRequestIssues = await queryIssues(githubClient, '', [IssueLabel.PATCH_REQUEST])
                if (patchRequestIssues && patchRequestIssues.length > 0) {
                    comment += `
Please also update outstanding patch requests, if relevant:

${patchRequestIssues.map(issue => `* #${issue.number}`).join('\n')}`
                }
                if (!config.dryRun.trackingIssues) {
                    const commentURL = await commentOnIssue(githubClient, trackingIssue, comment)
                    console.log(`Please make sure to follow up on the release issue: ${commentURL}`)
                } else {
                    console.log(`dryRun enabled, skipping GitHub comment to ${trackingIssue.url}: ${comment}`)
                }
            }
        },
    },
    {
        id: 'release:close',
        description: 'Close tracking issues for current release',
        run: async config => {
            const active = await getActiveRelease(config)
            // close tracking issue
            await closeTrackingIssue(active.version)
            console.log(chalk.blue('Deactivating release...'))
            removeScheduledRelease(config, active.version.version)
            deactivateAllReleases(config)
            console.log(chalk.green(`Release ${active.version.format()} closed!`))
        },
    },
    {
        id: 'release:prepare',
        description: 'Schedule a release',
        run: async config => {
            const rel = await newReleaseFromInput()
            addScheduledRelease(config, rel)
            saveReleaseConfig(config)
        },
    },
    {
        id: 'release:remove',
        description: 'Remove a release from the config',
        argNames: ['version'],
        run: async (config, version) => {
            await verifyWithInput(`Confirm you want to remove release: ${version} from the release config?`)
            const rconfig = loadReleaseConfig()
            removeScheduledRelease(rconfig, version)
            saveReleaseConfig(rconfig)
        },
    },
    {
        id: 'release:activate-release',
        description: 'Activate a feature release',
        run: async config => {
            await activateRelease(config)
        },
    },
    {
        id: 'release:deactivate-release',
        description: 'De-activate a feature release',
        run: async config => {
            await verifyWithInput('Are you sure you want to deactivate all releases?')
            deactivateAllReleases(config)
        },
    },
    {
        id: 'release:bake-content',
        description:
            'Bake constants and other static content into the release. Only required for minor / major versions.',
        run: async config => {
            const release = await getActiveRelease(config)
            if (release.version.patch !== 0) {
                console.log('content bake is only required for major / minor versions')
                exit(1)
            }

            // Creates PR's for executor release steps
            await bakeGoogleExecutorsSteps(config)
            await bakeAWSExecutorsSteps(config)

            const releaseBranch = release.branch
            const version = release.version.version
            ensureReleaseBranchUpToDate(releaseBranch)

            const multiVersionSteps: Edit[] = [
                `git remote set-branches --add origin '${releaseBranch}'`,
                `git fetch --depth 1 origin ${releaseBranch}`,
                combyReplace(
                    'const maxVersionString = ":[1]"',
                    version,
                    'internal/database/migration/shared/data/cmd/generator/consts.go'
                ),
                'cd internal/database/migration/shared && go run ./data/cmd/generator --write-frozen=false',
            ]
            const srcCliSteps = await bakeSrcCliSteps(config)

            const mainBranchEdits: Edit[] = [
                ...multiVersionSteps,
                ...srcCliSteps,
                ...batchChangesInAppChangelog(new SemVer(release.version.version).inc('minor'), true), // in the next main branch this will reflect the guessed next version
                indexerUpdate(),
            ]

            const releaseBranchEdits: Edit[] = [
                ...multiVersionSteps,
                ...srcCliSteps,
                ...batchChangesInAppChangelog(release.version, false),
                indexerUpdate(),
            ]

            const prDetails = {
                body: pullRequestBody(`Bake constants and static content into version v${version}.`),
                title: `v${version} bake constants and static content`,
                commitMessage: `bake constants and static content for version v${version}`,
            }

            const sets = await createChangesets({
                requiredCommands: ['comby', 'go'],
                changes: [
                    {
                        ...prDetails,
                        repo: 'sourcegraph',
                        owner: 'sourcegraph',
                        base: 'main',
                        head: `${version}-bake`,
                        edits: mainBranchEdits,
                        labels: [releaseBlockerLabel],
                    },
                    {
                        ...prDetails,
                        repo: 'sourcegraph',
                        owner: 'sourcegraph',
                        base: releaseBranch,
                        head: `${version}-bake-rb`,
                        edits: releaseBranchEdits,
                        labels: [releaseBlockerLabel],
                    },
                ],
                dryRun: config.dryRun.changesets,
            })
            console.log('Merge the following pull requests:\n')
            for (const set of sets) {
                console.log(set.pullRequestURL)
            }
        },
    },
    {
        id: 'release:create-tags',
        description: 'Release a new version of src-cli and executors. Only required for minor and major versions',
        run: async config => {
            const release = await getActiveRelease(config)
            if (release.version.patch !== 0) {
                console.log('src-cli and executors releases are only supported in this tool for major / minor releases')
                exit(1)
            }

            const repos = ['src-cli', 'terraform-google-executors', 'terraform-aws-executors']
            const tags: ReleaseTag[] = []

            const client = await getAuthenticatedGitHubClient()

            for (const repo of repos) {
                const { workdir } = await cloneRepo(client, 'sourcegraph', repo, {
                    revision: repo === 'src-cli' ? 'main' : 'master',
                    revisionMustExist: true,
                })

                switch (repo) {
                    case 'src-cli': {
                        const next = await nextSrcCliVersionInputWithAutodetect(config, workdir)
                        setSrcCliVersion(config, next.version)
                        tags.push({
                            repo,
                            nextTag: next.version,
                            workDir: workdir,
                        })
                        break
                    }
                    case 'terraform-google-executors': {
                        const nextGoogle = await nextGoogleExecutorVersionInputWithAutodetect(config, workdir)
                        setGoogleExecutorVersion(config, nextGoogle.version)
                        tags.push({
                            repo,
                            nextTag: nextGoogle.version,
                            workDir: workdir,
                        })
                        break
                    }
                    case 'terraform-aws-executors': {
                        const nextAWS = await nextAWSExecutorVersionInputWithAutodetect(config, workdir)
                        setAWSExecutorVersion(config, nextAWS.version)
                        tags.push({
                            repo,
                            nextTag: nextAWS.version,
                            workDir: workdir,
                        })
                        break
                    }
                }
            }

            for (const tag of tags) {
                const { repo, nextTag, workDir } = tag
                if (!config.dryRun.changesets) {
                    // actually execute the release
                    if (repo === 'src-cli') {
                        await execa('bash', ['-c', 'yes | ./release.sh'], {
                            stdio: 'inherit',
                            cwd: workDir,
                            env: { ...process.env, VERSION: nextTag },
                        })
                    } else {
                        await execa('bash', ['-c', `yes | ./release.sh ${nextTag}`], {
                            stdio: 'inherit',
                            cwd: workDir,
                        })
                    }
                } else {
                    console.log(chalk.blue(`Skipping ${repo} release for dry run`))
                }
            }
        },
    },
    {
        id: 'release:verify-releases',
        description: 'Verify src-cli version is available in brew and npm and executors tags are available',
        run: async config => {
            let passed = true
            let expected = config.in_progress?.srcCliVersion
            let expectedAWSExecutor = config.in_progress?.awsExecutorVersion
            let expectedGoogleExecutor = config.in_progress?.googleExecutorVersion

            const formatVersion = function (val: string): string {
                if (val === expected) {
                    return chalk.green(val)
                }
                passed = false
                return chalk.red(val)
            }
            if (!config.in_progress?.srcCliVersion) {
                expected = await retryInput(
                    'Enter the expected version of src-cli: ',
                    val => !!semver.parse(val),
                    'Expected semver format'
                )
            } else {
                console.log(`Expecting src-cli version ${expected} from release config`)
            }
            if (!config.in_progress?.googleExecutorVersion) {
                expectedGoogleExecutor = await retryInput(
                    'Enter the expected version of the Google executor: ',
                    val => !!semver.parse(val),
                    'Expected semver format'
                )
            } else {
                console.log(`Expecting Google executor version v${expectedGoogleExecutor} from release config`)
            }

            if (!config.in_progress?.awsExecutorVersion) {
                expectedAWSExecutor = await retryInput(
                    'Enter the expected version of the AWS executor: ',
                    val => !!semver.parse(val),
                    'Expected semver format'
                )
            } else {
                console.log(`Expecting AWS executor version v${expectedAWSExecutor} from release config`)
            }

            const latestGoogleVersion = await getLatestTag('sourcegraph', 'terraform-google-executors')
            const latestAWSVersion = await getLatestTag('sourcegraph', 'terraform-aws-executors')

            if (latestGoogleVersion !== `v${expectedGoogleExecutor}`) {
                passed = false
            }
            console.log(`terraform-google-executors:\t${formatVersion(latestGoogleVersion)}`)

            if (latestAWSVersion !== `v${expectedAWSExecutor}`) {
                passed = false
            }

            console.log(`terraform-aws-executors:\t${formatVersion(latestAWSVersion)}`)

            const githubRelease = await getLatestSrcCliGithubRelease()
            console.log(`github:\t${formatVersion(githubRelease)}`)

            const brewVersion = execa.sync('bash', [
                '-c',
                "brew info sourcegraph/src-cli/src-cli -q | sed -n 's/.*stable \\([0-9]\\.[0-9]\\.[0-9]\\)/\\1/p'",
            ]).stdout
            console.log(`brew:\t${formatVersion(brewVersion)}`)

            const npmVersion = execa.sync('bash', ['-c', 'npm show @sourcegraph/src version']).stdout
            console.log(`npm:\t${formatVersion(npmVersion)}`)

            if (passed === true) {
                console.log(chalk.green('All versions matched expected version!'))
            } else {
                console.log(chalk.red('Failed to verify versions'))
                exit(1)
            }
        },
    },
    {
        id: 'util:clear-cache',
        description: 'Clear release tool cache',
        run: () => {
            rmdirSync(cacheFolder, { recursive: true })
        },
    },
    {
        id: 'util:previous-version',
        description: 'Calculate the previous version based on repo tags',
        argNames: ['version'],
        run: (config: ReleaseConfig, version?: string) => {
            let ver: SemVer | undefined
            if (version) {
                ver = new SemVer(version)
                console.log(`Getting previous version from: ${version}...`)
            } else {
                console.log('Getting previous version...')
            }
            const prev = getPreviousVersion(ver)
            console.log(chalk.green(`${prev.format()}`))
        },
    },
    {
        id: '_test:release-guide-content',
        description: 'Generate upgrade guides',
        argNames: ['previous', 'next'],
        run: (config, previous, next) => {
            for (const content of getAllUpgradeGuides(previous, next)) {
                console.log(content)
            }
        },
    },
    {
        id: '_test:release-guide-update',
        description: 'Test update the upgrade guides',
        argNames: ['previous', 'next', 'dir'],
        run: (config, previous, next, dir) => {
            updateUpgradeGuides(previous, next)(dir)
        },
    },
    {
        id: '_test:google-calendar',
        description: 'Test Google Calendar integration',
        run: async config => {
            const googleCalendar = await getClient()
            const release = await getActiveRelease(config)
            await ensureEvent(
                {
                    title: 'TEST EVENT',
                    startDateTime: new Date(release.releaseDate).toISOString(),
                    endDateTime: addMinutes(new Date(release.releaseDate), 1).toISOString(),
                    transparency: 'transparent',
                },
                googleCalendar
            )
        },
    },
    {
        id: '_test:slack',
        description: 'Test Slack integration',
        argNames: ['channel', 'message'],
        run: async ({ dryRun }, channel, message) => {
            if (!dryRun.slack) {
                await postMessage(message, channel)
            }
        },
    },
    {
        id: '_test:batchchange-create-from-changes',
        description: 'Test batch changes integration',
        argNames: ['batchchangeConfigJSON'],
        // Example: pnpm run release _test:batchchange-create-from-changes "$(cat ./.secrets/test-batch-change-import.json)"
        run: async (_config, batchchangeConfigJSON) => {
            const batchChangeConfig = JSON.parse(batchchangeConfigJSON) as {
                changes: CreatedChangeset[]
                name: string
                description: string
            }

            // set up src-cli
            await commandExists('src')
            const batchChange = {
                name: batchChangeConfig.name,
                description: batchChangeConfig.description,
                namespace: 'sourcegraph',
                cliConfig: await batchChanges.sourcegraphCLIConfig(),
            }

            await batchChanges.createBatchChange(
                batchChangeConfig.changes,
                batchChange,
                'release tool testing batch change'
            )
            console.log(`Created batch change ${batchChanges.batchChangeURL(batchChange)}`)
        },
    },
    {
        id: '_test:config',
        description: 'Test release configuration loading',
        run: config => {
            console.log(JSON.stringify(config, null, 2))
        },
    },
    {
        id: '_test:dockerensure',
        description: 'test docker ensure function',
        run: async () => {
            try {
                await ensureDocker()
            } catch (error) {
                console.log(error)
                console.log('Docker required for batch changes')
                process.exit(1)
            }
        },
    },
    {
        id: '_test:srccliensure',
        description: 'test srccli version',
        run: async () => {
            ensureSrcCliEndpoint()
            await ensureSrcCliUpToDate()
        },
    },
    {
        id: '_test:patch-dates',
        description: 'test patch dates',
        run: () => {
            console.log(newRelease(new SemVer('1.0.0'), DateTime.fromISO('2023-03-22'), 'test', 'test'))
        },
    },
]
