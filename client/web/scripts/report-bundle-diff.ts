/* eslint-disable @typescript-eslint/no-var-requires */
/* eslint-disable @typescript-eslint/no-require-imports */

/* eslint-disable @typescript-eslint/no-unsafe-assignment */
const path = require('path')

const { Octokit } = require('octokit')
const shelljs = require('shelljs')

const COMMENT_HEADING = '## Bundle size report 📦'

async function main(): Promise<void> {
    try {
        const [commitFilename, compareFilename] = process.argv.slice(-2)

        const report = parseReport(commitFilename, compareFilename)
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

function parseReport(commitFilename: string, compareFilename: string): Report {
    const queryFile = path.join(__dirname, 'report-bundle-jora-query')

    const commitFile = path.join('..', '..', commitFilename)
    const compareFile = path.join('..', '..', compareFilename)

    const statoscope = path.join(__dirname, '..', '..', '..', 'node_modules', '.bin', 'statoscope')

    const rawReport = shelljs.exec(`cat "${queryFile}" | ${statoscope} query -i "${compareFile}" -i "${commitFile}"`)

    return JSON.parse(rawReport) as Report
}

function reportToMarkdown(report: Report): string {
    const { diffWith, hash } = report[0]
    const initialSizeMetric = report[1]
    const totalSizeMetric = report[2]
    const asyncSizeMetric = report[3]
    const modulesMetric = report[9]

    const initialSize = describeMetric(initialSizeMetric, 5000) // 5kb
    const totalSize = describeMetric(totalSizeMetric, 10000) // 10kb
    const asyncSize = describeMetric(asyncSizeMetric, 10000) // 10kb
    const modules = describeMetric(modulesMetric, 0)

    const url = `https://storage.cloud.google.com/sourcegraph_reports/statoscope-reports/${process.env.BRANCH}/#diff&diffWith=${diffWith}&hash=${hash}`

    return `${COMMENT_HEADING}

| Initial size | Total size | Async size | Modules |
| --- | --- | --- | --- |
| ${initialSize} | ${totalSize} | ${asyncSize} | ${modules} |

[View change in Statoscope report](${url})

<details>
<summary>See raw data and explaination</summary>

- \`Initial size\` is the size of the initial bundle (the one that is loaded when you open the page)
- \`Total size\` is the size of the initial bundle + all the async loaded chunks
- \`Async size\` is the size of all the async loaded chunks
- \`Modules\` is the number of modules in the initial bundle
</details>`
}

function describeMetric(metric: Metric, treshold: number): string {
    if (metric.value > treshold) {
        return `${metric.valueTextP} (+${metric.valueText}) 🔺`
    }
    if (metric.value < -treshold) {
        return `${metric.valueTextP} (${metric.valueText}) 🔽`
    }
    return `${metric.valueTextP} (${metric.value > 0 ? '+' : ''}${metric.valueText})`
}

async function createOrUpdateComment(body: string): Promise<void> {
    const pullRequest = parseInt(process.env.BUILDKITE_PULL_REQUEST ?? '', 10)
    const [owner, _repo] =
        process.env.BUILDKITE_PULL_REQUEST_REPO?.replace('https://github.com/', '').replace('.git', '').split('/') ?? []
    const repo = { owner, repo: _repo }
    const octokit = new Octokit({ auth: process.env.GITHUB_TOKEN })
    console.log({ pullRequest, owner, _repo })
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
            await octokit.rest.issues.createComment({
                ...repo,
                issue_number: pullRequest,
                body,
            })
        } catch (error) {
            console.error(error)
            console.log(
                "Error creating comment. This can happen for PR's originating from a fork without write permissions."
            )
        }
    } else {
        try {
            await octokit.rest.issues.updateComment({
                ...repo,
                // eslint-disable-next-line camelcase
                comment_id: sizeLimitComment.id,
                body,
            })
        } catch (error) {
            console.error(error)
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
