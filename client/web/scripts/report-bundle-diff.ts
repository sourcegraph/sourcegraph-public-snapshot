/* eslint-disable @typescript-eslint/no-var-requires */
/* eslint-disable @typescript-eslint/no-require-imports */
/* eslint-disable @typescript-eslint/no-unsafe-assignment */
const path = require('path')

const { Octokit } = require('octokit')
const shelljs = require('shelljs')

const COMMENT_HEADING = '## Bundle size report 📦'

async function main(): Promise<void> {
    try {
        const [commitFilename, mergeBaseFilename] = process.argv.slice(-2)

        const report = parseReport(commitFilename, mergeBaseFilename)
        const body = reportToMarkdown(report)
        await createOrUpdateComment(body)

        console.log(body)
    } catch (error) {
        console.error(error)
        process.exit(1)
    }
}
// eslint-disable-next-line @typescript-eslint/no-floating-promises
main()

interface Header {
    hash: string
    diffWith: string
}
interface Metric {
    value: number
    valueP: number
    valueText: string
    valueTextP: string
    label: string
    visible: number
}
type Report = [Header, Metric, Metric, Metric, Metric, Metric, Metric, Metric, Metric, Metric, Metric, Metric, Metric]

function parseReport(commitFilename: string, mergeBaseFilename: string): Report {
    const queryFile = path.join(__dirname, 'report-bundle-jora-query')

    const commitFile = path.join('..', '..', commitFilename)
    const mergeBaseFile = path.join('..', '..', mergeBaseFilename)

    const statoscope = path.join(__dirname, '..', '..', '..', 'node_modules', '.bin', 'statoscope')

    console.log({ queryFile, commitFile, mergeBaseFile, statoscope })

    const rawReport = shelljs.exec(`cat "${queryFile}" | ${statoscope} query -i "${commitFile}" -i "${mergeBaseFile}"`)

    return JSON.parse(rawReport) as Report
}

function reportToMarkdown(report: Report): string {
    return COMMENT_HEADING + '\n\n```\n' + JSON.stringify(report, null, 2) + '\n```'
}

async function createOrUpdateComment(body: string): Promise<void> {
    const pullRequest = parseInt(process.env.BUILDKITE_PULL_REQUEST ?? '', 10)
    const [owner, _repo] =
        process.env.BUILDKITE_PULL_REQUEST_REPO?.replace('https://github.com/', '').replace('.git', '').split('/') ?? []
    const repo = { owner, repo: _repo }
    const octokit = new Octokit({ auth: process.env.GITHUB_TOKEN })
    if (!pullRequest || !owner || !_repo) {
        console.log({ pullRequest, owner, _repo })
        throw new Error('No BUILDKITE_PULL_REQUEST or BUILDKITE_PULL_REQUEST_REPO env vars set')
    }

    const {
        data: { login },
    } = await octokit.rest.users.getAuthenticated()
    console.log('Hello, %s', login)

    const sizeLimitComment = await fetchPreviousComment(octokit, repo, pullRequest)

    if (!sizeLimitComment) {
        try {
            await octokit.issues.createComment({
                ...repo,
                issue_number: pullRequest,
                body,
            })
        } catch {
            console.log(
                "Error creating comment. This can happen for PR's originating from a fork without write permissions."
            )
        }
    } else {
        try {
            await octokit.issues.updateComment({
                ...repo,
                // eslint-disable-next-line camelcase
                comment_id: sizeLimitComment.id,
                body,
            })
        } catch {
            console.log(
                "Error updating comment. This can happen for PR's originating from a fork without write permissions."
            )
        }
    }
}

async function fetchPreviousComment(
    octokit: any,
    repo: { owner: string; repo: string },
    pullRequest: number
): Promise<any> {
    const commentList = await octokit.paginate('GET /repos/:owner/:repo/issues/:issue_number/comments', {
        ...repo,
        issue_number: pullRequest,
    })

    const sizeLimitComment = commentList.find((comment: any) => comment.body.startsWith(COMMENT_HEADING))
    return !sizeLimitComment ? null : sizeLimitComment
}
