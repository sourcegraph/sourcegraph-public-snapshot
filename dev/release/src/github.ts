import Octokit from '@octokit/rest'
import { readLine, formatDate, timezoneLink } from './util'
import { promisify } from 'util'
import * as semver from 'semver'
import { mkdtemp as original_mkdtemp } from 'fs'
import * as os from 'os'
import * as path from 'path'
import execa from 'execa'
import commandExists from 'command-exists'
const mkdtemp = promisify(original_mkdtemp)

export async function getAuthenticatedGitHubClient(): Promise<Octokit> {
    const githubPAT = await readLine(
        'Enter a GitHub personal access token with "repo" scope (https://github.com/settings/tokens/new): ',
        '.secrets/github.txt'
    )
    const trimmedGithubPAT = githubPAT.trim()
    return new Octokit({ auth: trimmedGithubPAT })
}

function dateMarkdown(date: Date, name: string): string {
    return `[${formatDate(date)}](${timezoneLink(date, name)})`
}

// Ensure these templates are up to date with the state of the tooling and release processes.
const templates = {
    releaseIssue: {
        owner: 'sourcegraph',
        repo: 'about',
        path: 'handbook/engineering/releases/release_issue_template.md',
    },
    patchReleaseIssue: {
        owner: 'sourcegraph',
        repo: 'about',
        path: 'handbook/engineering/releases/patch_release_issue_template.md',
    },
}

/**
 * Ensures a release ($MAJOR.$MINOR) tracking issue has been created with the given
 * parameters using `templates.releaseIssue`.
 */
export async function ensureReleaseTrackingIssue({
    version,
    assignees,
    releaseDateTime,
    oneWorkingDayBeforeRelease,
    fourWorkingDaysBeforeRelease,
    fiveWorkingDaysBeforeRelease,
    dryRun,
}: {
    version: semver.SemVer
    assignees: string[]
    releaseDateTime: Date
    oneWorkingDayBeforeRelease: Date
    fourWorkingDaysBeforeRelease: Date
    fiveWorkingDaysBeforeRelease: Date
    dryRun: boolean
}): Promise<{ url: string; created: boolean }> {
    const octokit = await getAuthenticatedGitHubClient()
    console.log(`Preparing issue from ${JSON.stringify(templates.releaseIssue)}`)
    const releaseIssueTemplate = await getContent(octokit, templates.releaseIssue)
    const majorMinor = `${version.major}.${version.minor}`
    const releaseIssueBody = releaseIssueTemplate
        .replace(/\$MAJOR/g, version.major.toString())
        .replace(/\$MINOR/g, version.minor.toString())
        .replace(/\$PATCH/g, version.patch.toString())
        .replace(/\$RELEASE_DATE/g, dateMarkdown(releaseDateTime, `${majorMinor} release date`))
        .replace(
            /\$FIVE_WORKING_DAYS_BEFORE_RELEASE/g,
            dateMarkdown(fiveWorkingDaysBeforeRelease, `Five working days before ${majorMinor} release`)
        )
        .replace(
            /\$FOUR_WORKING_DAYS_BEFORE_RELEASE/g,
            dateMarkdown(fourWorkingDaysBeforeRelease, `Four working days before ${majorMinor} release`)
        )
        .replace(
            /\$ONE_WORKING_DAY_BEFORE_RELEASE/g,
            dateMarkdown(oneWorkingDayBeforeRelease, `One working day before ${majorMinor} release`)
        )

    // Release milestones are not as emphasised now as they used to be, since most teams
    // use sprints shorter than releases to track their work. For reference, if one is
    // available we apply it to this tracking issue, otherwise just leave it without a
    // milestone.
    let milestoneNumber: number | undefined
    const milestone = await getReleaseMilestone(octokit, version)
    if (!milestone) {
        console.log(
            `Milestone ${JSON.stringify(releaseMilestoneName(version))} is closed or not found — omitting from issue.`
        )
    } else {
        milestoneNumber = milestone ? milestone.number : undefined
    }

    return ensureIssue(
        octokit,
        {
            title: trackingIssueTitle(version),
            owner: 'sourcegraph',
            repo: 'sourcegraph',
            assignees,
            body: releaseIssueBody,
            milestone: milestoneNumber,
            labels: ['release-tracking'],
        },
        dryRun
    )
}

/**
 * Ensures a patch release ($MAJOR.$MINOR.PATCH) tracking issue has been created with the
 * given parameters using `templates.releaseIssue`.
 */
export async function ensurePatchReleaseIssue({
    version,
    assignees,
    dryRun,
}: {
    version: semver.SemVer
    assignees: string[]
    dryRun: boolean
}): Promise<{ url: string; created: boolean }> {
    const octokit = await getAuthenticatedGitHubClient()
    console.log(`Preparing issue from ${JSON.stringify(templates.patchReleaseIssue)}`)
    const issueTemplate = await getContent(octokit, templates.patchReleaseIssue)
    const issueBody = issueTemplate
        .replace(/\$MAJOR/g, version.major.toString())
        .replace(/\$MINOR/g, version.minor.toString())
        .replace(/\$PATCH/g, version.patch.toString())
    return ensureIssue(
        octokit,
        {
            title: trackingIssueTitle(version),
            owner: 'sourcegraph',
            repo: 'sourcegraph',
            assignees,
            body: issueBody,
            labels: ['release-tracking'],
        },
        dryRun
    )
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
        labels?: string[]
    },
    dryRun: boolean
): Promise<{ url: string; created: boolean }> {
    const issueData = {
        title,
        owner,
        repo,
        assignees,
        milestone,
        labels,
    }
    if (dryRun) {
        console.log('Dry run enabled, skipping issue creation')
        console.log(`Issue that would have been created:\n${JSON.stringify(issueData, null, 1)}`)
        console.log(`With body: ${body}`)
        return { url: '', created: false }
    }
    const issue = await getIssueByTitle(octokit, title)
    if (issue) {
        return { url: issue.url, created: false }
    }
    const createdIssue = await octokit.issues.create({ body, ...issueData })
    return { url: createdIssue.data.html_url, created: true }
}

export async function listIssues(
    octokit: Octokit,
    query: string
): Promise<Octokit.SearchIssuesAndPullRequestsResponseItemsItem[]> {
    return (await octokit.search.issuesAndPullRequests({ per_page: 100, q: query })).data.items
}

export interface Issue {
    number: number
    url: string

    // Repository
    owner: string
    repo: string
}

export async function getTrackingIssue(client: Octokit, release: semver.SemVer): Promise<Issue | null> {
    return getIssueByTitle(client, trackingIssueTitle(release))
}

function releaseMilestoneName(release: semver.SemVer): string {
    return `${release.major}.${release.minor}${release.patch !== 0 ? `.${release.patch}` : ''}`
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
    const milestoneTitle = releaseMilestoneName(release)
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

function trackingIssueTitle(version: semver.SemVer): string {
    if (!version.patch) {
        return `${version.major}.${version.minor} release tracking issue`
    }
    return `${version.version} patch release tracking issue`
}

async function getIssueByTitle(octokit: Octokit, title: string): Promise<Issue | null> {
    const owner = 'sourcegraph'
    const repo = 'sourcegraph'
    const response = await octokit.search.issuesAndPullRequests({
        per_page: 100,
        q: `type:issue repo:${owner}/${repo} is:open ${JSON.stringify(title)}`,
    })

    const matchingIssues = response.data.items.filter(issue => issue.title === title)
    if (matchingIssues.length === 0) {
        return null
    }
    if (matchingIssues.length > 1) {
        throw new Error(`Multiple issues matched issue title ${JSON.stringify(title)}`)
    }
    return { number: matchingIssues[0].number, url: matchingIssues[0].html_url, owner, repo }
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
    changes: (Octokit.PullsCreateParams & CreateBranchWithChangesOptions)[]
    dryRun?: boolean
}

export interface CreatedChangeset {
    repository: string
    branch: string
    pullRequestURL: string
    pullRequestNumber: number
}

export async function createChangesets(options: ChangesetsOptions): Promise<CreatedChangeset[]> {
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

    // Generate and push changes
    for (const change of options.changes) {
        const repository = `${change.owner}/${change.repo}`
        console.log(`${repository}: Preparing change for on '${change.base}' to '${change.head}'`)
        await createBranchWithChanges(octokit, { ...change, dryRun: options.dryRun })
    }

    // Publish changes as pull requests only if all changes are successfully created
    const results: CreatedChangeset[] = []
    for (const change of options.changes) {
        const repository = `${change.owner}/${change.repo}`
        console.log(`${repository}: Preparing pull request for change from '${change.base}' to '${change.head}':

Title: ${change.title}
Body: ${change.body || 'none'}`)
        let pullRequest: { url: string; number: number } = { url: '', number: -1 }
        if (!options.dryRun) {
            pullRequest = await createPR(octokit, change)
        }

        results.push({
            repository,
            branch: change.base,
            pullRequestURL: pullRequest.url,
            pullRequestNumber: pullRequest.number,
        })
    }

    // Log results
    for (const result of results) {
        console.log(`${result.repository} (${result.branch}): created pull request ${result.pullRequestURL}`)
    }

    return results
}

async function cloneRepo(
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
    const fetchFlags = '--depth 10'

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

    // Set up repository
    const setupScript = `set -ex

git clone ${fetchFlags} git@github.com:${owner}/${repo} || git clone ${fetchFlags} https://github.com/${owner}/${repo};
cd ${repo};
${checkoutCommand};`
    await execa('bash', ['-c', setupScript], { stdio: 'inherit', cwd: tmpdir })
    return {
        workdir: path.join(tmpdir, repo),
    }
}

async function createBranchWithChanges(
    octokit: Octokit,
    { owner, repo, base: baseRevision, head: headBranch, commitMessage, edits, dryRun }: CreateBranchWithChangesOptions
): Promise<void> {
    // Set up repository
    const { workdir } = await cloneRepo(octokit, owner, repo, { revision: baseRevision })

    // Apply edits
    for (const edit of edits) {
        switch (typeof edit) {
            case 'function':
                edit(workdir)
                break
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
    { owner, repo, branch: rawBranch, tag: rawTag }: TagOptions,
    dryRun: boolean
): Promise<void> {
    const { workdir } = await cloneRepo(octokit, owner, repo, { revision: rawBranch, revisionMustExist: true })
    const branch = JSON.stringify(rawBranch)
    const tag = JSON.stringify(rawTag)
    const finalizeTag = dryRun ? `git --no-pager show ${tag} --no-patch` : `git push origin ${tag}`
    console.log(
        dryRun
            ? `Dry-run enabled - creating and printing tag ${tag} on ${owner}/${repo}@${branch}`
            : `Creating and pushing tag ${tag} on ${owner}/${repo}@${branch}`
    )
    await execa('bash', ['-c', `git tag -a ${tag} -m ${tag} && ${finalizeTag}`], { stdio: 'inherit', cwd: workdir })
}
