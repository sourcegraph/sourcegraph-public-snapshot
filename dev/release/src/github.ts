import Octokit from '@octokit/rest'
import { readLine } from './util'
import { readFile } from 'mz/fs'

export async function ensureTrackingIssue({
    majorVersion,
    minorVersion,
    assignees,
    releaseDateTime,
    oneWorkingDayBeforeRelease,
    threeWorkingDaysBeforeRelease,
    fourWorkingDaysBeforeRelease,
    fiveWorkingDaysBeforeRelease,
}: {
    majorVersion: string
    minorVersion: string
    assignees: string[]
    releaseDateTime: Date
    oneWorkingDayBeforeRelease: Date
    threeWorkingDaysBeforeRelease: Date
    fourWorkingDaysBeforeRelease: Date
    fiveWorkingDaysBeforeRelease: Date
}): Promise<{ url: string; created: boolean }> {
    const octokit = await getAuthenticatedGitHubClient()
    const url = await getTrackingIssueURL(octokit, majorVersion, minorVersion)
    if (url) {
        return { url, created: false }
    }

    const formatDate = (d: Date): string => `${d.getFullYear()}-${d.getMonth()}-${d.getDate()}`

    const releaseIssueTemplate = await readFile(
        '../../../about/handbook/engineering/releases/release_issue_template.md',
        { encoding: 'utf8' }
    )
    const releaseIssueBody = releaseIssueTemplate
        .replace(/\$MAJOR/g, majorVersion)
        .replace(/\$MINOR/g, minorVersion)
        .replace(/\$RELEASE_DATE/g, formatDate(releaseDateTime))
        .replace(/\$FIVE_WORKING_DAYS_BEFORE_RELEASE/g, formatDate(fiveWorkingDaysBeforeRelease))
        .replace(/\$FOUR_WORKING_DAYS_BEFORE_RELEASE/g, formatDate(fourWorkingDaysBeforeRelease))
        .replace(/\$THREE_WORKING_DAYS_BEFORE_RELEASE/g, formatDate(threeWorkingDaysBeforeRelease))
        .replace(/\$ONE_WORKING_DAY_BEFORE_RELEASE/g, formatDate(oneWorkingDayBeforeRelease))

    const createdIssue = await octokit.issues.create({
        title: issueTitle(majorVersion, minorVersion),
        owner: 'sourcegraph',
        repo: 'sourcegraph',
        assignees,
        body: releaseIssueBody,
    })
    return { url: createdIssue.data.html_url, created: true }
}

export async function listIssues(
    octokit: Octokit,
    query: string
): Promise<Octokit.SearchIssuesAndPullRequestsResponseItemsItem[]> {
    return (await octokit.search.issuesAndPullRequests({ per_page: 100, q: query })).data.items
}

export async function getTrackingIssueURL(
    octokit: Octokit,
    majorVersion: string,
    minorVersion: string
): Promise<string | null> {
    const title = issueTitle(majorVersion, minorVersion)
    const resp = await octokit.search.issuesAndPullRequests({
        per_page: 100,
        q: `type:issue repo:sourcegraph/sourcegraph is:open ${JSON.stringify(title)}`,
    })

    const matchingIssues = resp.data.items.filter(issue => issue.title === title)
    if (matchingIssues.length === 0) {
        return null
    }
    if (matchingIssues.length > 1) {
        throw new Error(`Multiple issues matched issue title ${JSON.stringify(title)}`)
    }
    return matchingIssues[0].html_url
}

function issueTitle(major: string, minor: string): string {
    return `${major}.${minor} release tracking issue`
}

export async function getAuthenticatedGitHubClient(): Promise<Octokit> {
    const githubPAT = await readLine(
        'Enter a GitHub personal access token with "repo" scope (https://github.com/settings/tokens/new): ',
        '.secrets/github.txt'
    )
    return new Octokit({ auth: githubPAT })
}
