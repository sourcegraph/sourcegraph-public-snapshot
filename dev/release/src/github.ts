import { existsSync, mkdtemp as original_mkdtemp, readFileSync } from 'fs'
import * as os from 'os'
import * as path from 'path'
import { promisify } from 'util'

import Octokit, { type IssuesAddLabelsParams } from '@octokit/rest'
import commandExists from 'command-exists'
import execa from 'execa'
import fetch from 'node-fetch'
import * as semver from 'semver'

import type { ActiveRelease } from './config'
import { cacheFolder, changelogURL, formatDate, getContainerRegistryCredential, readLine, timezoneLink } from './util'

const mkdtemp = promisify(original_mkdtemp)
let githubPAT: string

export async function getAuthenticatedGitHubClient(): Promise<Octokit> {
    const cacheFile = `${cacheFolder}/github.txt`
    if (existsSync(cacheFile) && (await validateToken()) === true) {
        githubPAT = readFileSync(`${cacheFolder}/github.txt`, 'utf-8')
    } else {
        githubPAT = await readLine(
            'Enter a GitHub personal access token with "repo" scope (https://github.com/settings/tokens/new): ',
            cacheFile
        )
    }

    const trimmedGithubPAT = githubPAT.trim()
    return new Octokit({ auth: trimmedGithubPAT })
}

/**
 * releaseName generates a standardized format for referring to releases.
 */
export function releaseName(release: semver.SemVer): string {
    return `${release.major}.${release.minor}${release.patch !== 0 ? `.${release.patch}` : ''}`
}

export enum IssueLabel {
    // https://github.com/sourcegraph/sourcegraph/labels/release-tracking
    RELEASE_TRACKING = 'release-tracking',
    // https://github.com/sourcegraph/sourcegraph/labels/patch-release-request
    PATCH_REQUEST = 'patch-release-request',

    // New labels to better distinguish release-tracking issues
    RELEASE = 'release',
    PATCH = 'patch',
    MANAGED = 'managed-instances',
    DEVOPS_TEAM = 'team/devops',
    SECURITY_TEAM = 'team/security',
    RELEASE_BLOCKER = 'release-blocker',
}

enum IssueTitleSuffix {
    RELEASE_TRACKING = 'release tracking issue',
    PATCH_TRACKING = 'patch release tracking issue',
    MANAGED_TRACKING = 'upgrade managed instances tracking issue',
    SECURITY_TRACKING = 'container image vulnerability assessment tracking issue',
}

/**
 * Template used to generate tracking issue
 */
interface IssueTemplate {
    owner: string
    repo: string
    /**
     * Relative path to markdown file containing template body.
     *
     * Template bodies can leverage arguments as described in `IssueTemplateArguments` docstrings.
     */
    path: string
    /**
     * Title for issue.
     */
    titleSuffix: IssueTitleSuffix
    /**
     * Labels to apply on issues.
     */
    labels: string[]
}

/**
 * Arguments available for rendering IssueTemplate
 */
interface IssueTemplateArguments {
    /**
     * Available as `$MAJOR`, `$MINOR`, and `$PATCH`
     */
    version: semver.SemVer
    /**
     * Available as `$SECURITY_REVIEW_DATE`
     */
    securityReviewDate: Date
    /**
     * Available as `$CODE_FREEZE_DATE`
     */
    codeFreezeDate: Date
    /**
     * Available as `$RELEASE_DATE`
     */
    releaseDate: Date
}

/**
 * Configure templates for the release tool to generate issues with.
 *
 * Ensure these templates are up to date with the state of the tooling and release processes.
 */
// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
const getTemplates = () => {
    const releaseIssue: IssueTemplate = {
        owner: 'sourcegraph',
        repo: 'sourcegraph',
        path: 'dev/release/templates/release_issue_template.md',
        titleSuffix: IssueTitleSuffix.RELEASE_TRACKING,
        labels: [IssueLabel.RELEASE_TRACKING, IssueLabel.RELEASE],
    }
    const patchReleaseIssue: IssueTemplate = {
        owner: 'sourcegraph',
        repo: 'sourcegraph',
        path: 'dev/release/templates/patch_release_issue_template.md',
        titleSuffix: IssueTitleSuffix.PATCH_TRACKING,
        labels: [IssueLabel.RELEASE_TRACKING, IssueLabel.PATCH],
    }
    const securityAssessmentIssue: IssueTemplate = {
        owner: 'sourcegraph',
        repo: 'sourcegraph',
        path: 'dev/release/templates/security_assessment.md',
        titleSuffix: IssueTitleSuffix.SECURITY_TRACKING,
        labels: [IssueLabel.RELEASE_TRACKING, IssueLabel.SECURITY_TEAM, IssueLabel.RELEASE_BLOCKER],
    }
    return { releaseIssue, patchReleaseIssue, securityAssessmentIssue }
}

function dateMarkdown(date: Date, name: string): string {
    return `[${formatDate(date)}](${timezoneLink(date, name)})`
}

async function execTemplate(
    octokit: Octokit,
    template: IssueTemplate,
    { version, securityReviewDate, codeFreezeDate, releaseDate }: IssueTemplateArguments
): Promise<string> {
    console.log(`Preparing issue from ${JSON.stringify(template)}`)
    const name = releaseName(version)
    const content = await getContent(octokit, template)
    return content
        .replaceAll('$MAJOR', version.major.toString())
        .replaceAll('$MINOR', version.minor.toString())
        .replaceAll('$PATCH', version.patch.toString())
        .replaceAll(
            '$SECURITY_REVIEW_DATE',
            dateMarkdown(securityReviewDate, `One working week before ${name} release`)
        )
        .replaceAll('$CODE_FREEZE_DATE', dateMarkdown(codeFreezeDate, `Three working days before ${name} release`))
        .replaceAll('$RELEASE_DATE', dateMarkdown(releaseDate, `${name} release date`))
}

interface MaybeIssue {
    title: string
    url: string
    number: number
    created: boolean
}

/**
 * Ensures tracking issues for the given release.
 *
 * The first returned issue is considered the parent issue.
 */
export async function ensureTrackingIssues({
    version,
    assignees,
    releaseDate,
    securityReviewDate,
    codeFreezeDate,
    dryRun,
}: {
    version: semver.SemVer
    assignees: string[]
    releaseDate: Date
    securityReviewDate: Date
    codeFreezeDate: Date
    dryRun: boolean
}): Promise<MaybeIssue[]> {
    const octokit = await getAuthenticatedGitHubClient()
    const templates = getTemplates()
    const release = releaseName(version)

    // Determine what issues to generate. The first issue is considered the "main"
    // tracking issue, and subsequent issues will contain references to it.
    let issueTemplates: IssueTemplate[]
    if (version.patch === 0) {
        issueTemplates = [templates.releaseIssue]
    } else {
        issueTemplates = [templates.patchReleaseIssue]
    }

    // Release milestones are not as emphasised now as they used to be, since most teams
    // use sprints shorter than releases to track their work. For reference, if one is
    // available we apply it to this tracking issue, otherwise just leave it without a
    // milestone.
    let milestoneNumber: number | undefined
    const milestone = await getReleaseMilestone(octokit, version)
    if (!milestone) {
        console.log(`Milestone ${release} is closed or not found — omitting from issue.`)
    } else {
        milestoneNumber = milestone ? milestone.number : undefined
    }

    // Create issues
    let parentIssue: MaybeIssue | undefined
    const created: MaybeIssue[] = []
    for (const template of issueTemplates) {
        const body = await execTemplate(octokit, template, {
            version,
            releaseDate,
            securityReviewDate,
            codeFreezeDate,
        })
        const issue = await ensureIssue(
            octokit,
            {
                title: trackingIssueTitle(version, template),
                labels: template.labels,
                body: parentIssue ? `${body}\n\n---\n\nAlso see [${parentIssue.title}](${parentIssue.url})` : body,
                assignees,
                owner: 'sourcegraph',
                repo: 'sourcegraph',
                milestone: milestoneNumber,
            },
            dryRun
        )
        // if this is the first issue, we treat it as the parent issue
        if (!parentIssue) {
            parentIssue = { ...issue }
        }
        created.push({ ...issue })
    }
    return created
}

async function getContent(
    octokit: Octokit,
    parameters: {
        owner: string
        repo: string
        path: string
    }
): Promise<string> {
    const response = await octokit.repos.getContents(parameters)
    if (Array.isArray(response.data)) {
        throw new TypeError(`${parameters.path} is a directory`)
    }
    return Buffer.from(response.data.content as string, 'base64').toString()
}

async function ensureIssue(
    octokit: Octokit,
    {
        title,
        owner,
        repo,
        assignees,
        body,
        milestone,
        labels,
    }: {
        title: string
        owner: string
        repo: string
        assignees: string[]
        body: string
        milestone?: number
        labels: string[]
    },
    dryRun: boolean
): Promise<MaybeIssue> {
    const issueData = {
        title,
        owner,
        repo,
        assignees,
        milestone,
        labels,
    }
    const issue = await getIssueByTitle(octokit, title, labels)
    if (issue) {
        return { title, url: issue.url, number: issue.number, created: false }
    }
    if (dryRun) {
        console.log('Dry run enabled, skipping issue creation')
        console.log(`Issue that would have been created:\n${JSON.stringify(issueData, null, 1)}`)
        console.log(`With body: ${body}`)
        return { title, url: '', number: 0, created: false }
    }
    const createdIssue = await octokit.issues.create({ body, ...issueData })
    return { title, url: createdIssue.data.html_url, number: createdIssue.data.number, created: true }
}

export async function listIssues(
    octokit: Octokit,
    query: string
): Promise<Octokit.SearchIssuesAndPullRequestsResponseItemsItem[]> {
    return (await octokit.search.issuesAndPullRequests({ per_page: 100, q: query })).data.items
}

export interface Issue {
    title: string
    number: number
    url: string

    // Repository
    owner: string
    repo: string
}

export async function getTrackingIssue(client: Octokit, release: semver.SemVer): Promise<Issue | null> {
    const templates = getTemplates()
    const template = release.patch ? templates.patchReleaseIssue : templates.releaseIssue
    return getIssueByTitle(client, trackingIssueTitle(release, template), template.labels)
}

function trackingIssueTitle(release: semver.SemVer, template: IssueTemplate): string {
    return `${release.version} ${template.titleSuffix}`
}

export async function commentOnIssue(client: Octokit, issue: Issue, body: string): Promise<string> {
    const comment = await client.issues.createComment({
        body,
        issue_number: issue.number,
        owner: issue.owner,
        repo: issue.repo,
    })
    return comment.data.html_url
}

async function closeIssue(client: Octokit, issue: Issue): Promise<void> {
    await client.issues.update({
        state: 'closed',
        issue_number: issue.number,
        owner: issue.owner,
        repo: issue.repo,
    })
}

interface Milestone {
    number: number
    url: string

    // Repository
    owner: string
    repo: string
}

async function getReleaseMilestone(client: Octokit, release: semver.SemVer): Promise<Milestone | null> {
    const owner = 'sourcegraph'
    const repo = 'sourcegraph'
    const milestoneTitle = releaseName(release)
    const milestones = await client.issues.listMilestonesForRepo({
        owner,
        repo,
        per_page: 100,
        direction: 'desc',
    })
    const milestone = milestones.data.filter(milestone => milestone.title === milestoneTitle)
    return milestone.length > 0
        ? {
              number: milestone[0].number,
              url: milestone[0].html_url,
              owner,
              repo,
          }
        : null
}

export async function queryIssues(octokit: Octokit, titleQuery: string, labels: string[]): Promise<Issue[]> {
    const owner = 'sourcegraph'
    const repo = 'sourcegraph'
    const response = await octokit.search.issuesAndPullRequests({
        per_page: 100,
        q: `type:issue repo:${owner}/${repo} is:open ${labels
            .map(label => `label:${label}`)
            .join(' ')} ${JSON.stringify(titleQuery)}`,
    })
    return response.data.items.map(item => ({
        title: item.title,
        number: item.number,
        url: item.html_url,
        owner,
        repo,
    }))
}

async function getIssueByTitle(octokit: Octokit, title: string, labels: string[]): Promise<Issue | null> {
    const matchingIssues = (await queryIssues(octokit, title, labels)).filter(issue => issue.title === title)
    if (matchingIssues.length === 0) {
        return null
    }
    if (matchingIssues.length > 1) {
        throw new Error(`Multiple issues matched issue title ${JSON.stringify(title)}`)
    }
    return matchingIssues[0]
}

export type EditFunc = (d: string) => void

export type Edit = string | EditFunc

export interface CreateBranchWithChangesOptions {
    owner: string
    repo: string
    base: string
    head: string
    commitMessage: string
    edits: Edit[]
    dryRun?: boolean
}

export interface ChangesetsOptions {
    requiredCommands: string[]
    changes: (Octokit.PullsCreateParams & CreateBranchWithChangesOptions & { labels?: string[] })[]
    dryRun?: boolean
}

export interface CreatedChangeset {
    repository: string
    branch: string
    pullRequestURL: string
    pullRequestNumber: number
}

export async function createChangesets(options: ChangesetsOptions): Promise<CreatedChangeset[]> {
    // Overwriting `process.env` may not be a good practice,
    // but it's the easiest way to avoid making changes all over the place
    const dockerHubCredential = await getContainerRegistryCredential('index.docker.io')
    process.env.CR_USERNAME = dockerHubCredential.username
    process.env.CR_PASSWORD = dockerHubCredential.password
    for (const command of options.requiredCommands) {
        try {
            await commandExists(command)
        } catch {
            throw new Error(`Required command ${command} does not exist`)
        }
    }
    const octokit = await getAuthenticatedGitHubClient()
    if (options.dryRun) {
        console.log('Changesets dry run enabled - diffs and pull requests will be printed instead')
    } else {
        console.log('Generating changes and publishing as pull requests')
    }

    // Generate and push changes. We abort here if a repo fails because it should be safe
    // to re-run changesets, which force push changes to a PR branch.
    for (const change of options.changes) {
        const repository = `${change.owner}/${change.repo}`
        console.log(`${repository}: Preparing change for on '${change.base}' to '${change.head}'`)
        await createBranchWithChanges(octokit, { ...change, dryRun: options.dryRun })
    }

    // Publish changes as pull requests only if all changes are successfully created. We
    // continue on error for errors when publishing.
    const results: CreatedChangeset[] = []
    let publishChangesFailed = false
    for (const change of options.changes) {
        const repository = `${change.owner}/${change.repo}`
        console.log(`${repository}: Preparing pull request for change from '${change.base}' to '${change.head}':

Title: ${change.title}
Body: ${change.body || 'none'}`)
        let pullRequest: { url: string; number: number } = { url: '', number: -1 }
        try {
            if (!options.dryRun) {
                pullRequest = await createPR(octokit, change)
                if (change.labels) {
                    await octokit.issues.addLabels({
                        issue_number: pullRequest.number,
                        repo: change.repo,
                        owner: change.owner,
                        labels: change.labels,
                    } as IssuesAddLabelsParams)
                }
            }

            results.push({
                repository,
                branch: change.base,
                pullRequestURL: pullRequest.url,
                pullRequestNumber: pullRequest.number,
            })
        } catch (error) {
            publishChangesFailed = true
            console.error(error)
            console.error(`Failed to create pull request for change in ${repository}`, { change })
        }
    }

    // Log results
    for (const result of results) {
        console.log(`${result.repository} (${result.branch}): created pull request ${result.pullRequestURL}`)
    }
    if (publishChangesFailed) {
        throw new Error('Error occured applying some changes - please check log output')
    }

    return results
}

export async function cloneRepo(
    octokit: Octokit,
    owner: string,
    repo: string,
    checkout: {
        revision: string
        revisionMustExist?: boolean
    }
): Promise<{
    workdir: string
}> {
    const tmpdir = await mkdtemp(path.join(os.tmpdir(), `sg-release-${owner}-${repo}-`))
    console.log(`Created temp directory ${tmpdir}`)
    const fetchFlags = '--depth 1'

    // Determine whether or not to create the base branch, or use the existing one
    let revisionExists = true
    if (!checkout.revisionMustExist) {
        try {
            await octokit.repos.getBranch({ branch: checkout.revision, owner, repo })
        } catch (error) {
            if (error.status === 404) {
                console.log(`Target revision ${checkout.revision} does not exist, this branch will be created`)
                revisionExists = false
            } else {
                throw error
            }
        }
    }
    const checkoutCommand =
        revisionExists === true
            ? // for an existing branch - fetch fails if we are already checked out, so ignore errors optimistically
              `git fetch ${fetchFlags} origin ${checkout.revision}:${checkout.revision} || true ; git checkout ${checkout.revision}`
            : // create from HEAD and publish base branch if it does not yet exist
              `git checkout -b ${checkout.revision} ; git push origin ${checkout.revision}:${checkout.revision}`

    // PERF: if we have a local clone using reference avoids needing to fetch
    // all the objects from the remote. We assume the local clone will exist
    // in the same directory as the current sourcegraph/sourcegraph clone.
    const cloneFlags = `${fetchFlags} --reference-if-able ${localSourcegraphRepo}/../${repo}`

    // Set up repository
    const setupScript = `set -ex

git clone ${cloneFlags} git@github.com:${owner}/${repo} || git clone ${cloneFlags} https://github.com/${owner}/${repo};
cd ${repo};
${checkoutCommand};`
    await execa('bash', ['-c', setupScript], { stdio: 'inherit', cwd: tmpdir })
    return {
        workdir: path.join(tmpdir, repo),
    }
}

export const localSourcegraphRepo = `${process.cwd()}/../..`

async function createBranchWithChanges(
    octokit: Octokit,
    { owner, repo, base: baseRevision, head: headBranch, commitMessage, edits, dryRun }: CreateBranchWithChangesOptions
): Promise<void> {
    // Set up repository
    const { workdir } = await cloneRepo(octokit, owner, repo, { revision: baseRevision })

    // Apply edits
    for (const edit of edits) {
        switch (typeof edit) {
            case 'function': {
                edit(workdir)
                break
            }
            case 'string': {
                const editScript = `set -ex

                ${edit};`
                await execa('bash', ['-c', editScript], { stdio: 'inherit', cwd: workdir })
            }
        }
    }

    if (dryRun) {
        console.warn('Dry run enabled - printing diff instead of publishing')
        const showChangesScript = `set -ex

        git --no-pager diff;`
        await execa('bash', ['-c', showChangesScript], { stdio: 'inherit', cwd: workdir })
    } else {
        // Publish changes. We force push to ensure that the generated changes are applied.
        const publishScript = `set -ex

        git add :/;
        git commit -a -m ${JSON.stringify(commitMessage)};
        git push --force origin HEAD:${headBranch};`
        await execa('bash', ['-c', publishScript], { stdio: 'inherit', cwd: workdir })
    }
}

async function createPR(
    octokit: Octokit,
    options: {
        owner: string
        repo: string
        head: string
        base: string
        title: string
        body?: string
    }
): Promise<{ url: string; number: number }> {
    const response = await octokit.pulls.create(options)
    return {
        url: response.data.html_url,
        number: response.data.number,
    }
}

export interface TagOptions {
    owner: string
    repo: string
    branch: string
    tag: string
}

/**
 * Creates a tag on a remote branch for the given repository.
 *
 * The target branch must exist on the remote.
 */
export async function createTag(
    octokit: Octokit,
    workdir: string,
    { owner, repo, branch: rawBranch, tag: rawTag }: TagOptions,
    dryRun: boolean
): Promise<void> {
    const branch = JSON.stringify(rawBranch)
    const tag = JSON.stringify(rawTag)
    const finalizeTag = dryRun ? `git --no-pager show ${tag} --no-patch` : `git push origin ${tag}`
    if (dryRun) {
        console.log(`Dry-run enabled - creating and printing tag ${tag} on ${owner}/${repo}@${branch}`)
        return
    }
    console.log(`Creating and pushing tag ${tag} on ${owner}/${repo}@${branch}`)
    await execa('bash', ['-c', `git tag -a ${tag} -m ${tag} && ${finalizeTag}`], { stdio: 'inherit', cwd: workdir })
}

// createLatestRelease generates a GitHub release iff this release is the latest and
// greatest, otherwise it is a no-op.
export async function createLatestRelease(
    octokit: Octokit,
    { owner, repo, release }: { owner: string; repo: string; release: semver.SemVer },
    dryRun?: boolean
): Promise<string> {
    const latest = await octokit.repos.getLatestRelease({
        owner,
        repo,
    })
    const latestTag = latest.data.tag_name
    if (semver.gt(latestTag.startsWith('v') ? latestTag.slice(1) : latestTag, release)) {
        // if latest is greater than release, do not generate a release
        console.log(`Latest release ${latestTag} is more recent than ${release.version}, skipping GitHub release`)
        return ''
    }

    const updateURL = 'https://docs.sourcegraph.com/admin/updates'
    const releasePostURL = `https://sourcegraph.com/blog/release/${release.major}.${release.minor}` // CI:URL_OK

    const request: Octokit.RequestOptions & Octokit.ReposCreateReleaseParams = {
        owner,
        repo,
        tag_name: `v${release.version}`,
        name: `Sourcegraph ${release.version}`,
        prerelease: false,
        draft: false,
        body: `Sourcegraph ${release.version} is now available!

- [Changelog](${changelogURL(release.format())})
- [Update](${updateURL})
- [Release post](${releasePostURL}) (might not be available immediately upon release)
`,
    }
    if (dryRun) {
        console.log('Skipping GitHub release, parameters:', request)
        return ''
    }
    const response = await octokit.repos.createRelease(request)
    return response.data.html_url
}

async function validateToken(): Promise<boolean> {
    const githubPAT: string = readFileSync(`${cacheFolder}/github.txt`, 'utf-8')
    const trimmedGithubPAT = githubPAT.trim()
    const response = await fetch('https://api.github.com/repos/sourcegraph/sourcegraph', {
        method: 'GET',
        headers: {
            Authorization: `token ${trimmedGithubPAT}`,
        },
    })

    if (response.status !== 200) {
        console.log(`Existing GitHub token is invalid, got status ${response.statusText}`)
        return false
    }
    return true
}

export async function closeTrackingIssue(version: semver.SemVer): Promise<void> {
    const octokit = await getAuthenticatedGitHubClient()
    const release = releaseName(version)
    const labels = [IssueLabel.RELEASE_TRACKING, IssueLabel.RELEASE]
    // close old tracking issue
    const previous = await queryIssues(octokit, release, labels)
    for (const previousIssue of previous) {
        const comment = await commentOnIssue(
            octokit,
            previousIssue,
            `Issue closed by release tool. #${previousIssue.number}`
        )
        console.log(`Closing #${previousIssue.number} '${previousIssue.title} with ${comment}`)
        await closeIssue(octokit, previousIssue)
    }
}

export const releaseBlockerLabel = 'release-blocker'

export function getBackportLabelForRelease(release: ActiveRelease): string {
    return `backport ${release.branch}`
}
