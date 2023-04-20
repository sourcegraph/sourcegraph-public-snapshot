/* eslint-disable no-console */

import { promises as fs } from 'fs'
import path from 'path'

import { Octokit } from 'octokit'
import { exec } from 'shelljs'

const COMMENT_HEADING = '## Bundle size report 📦'
const MERGE_BASE = exec('git merge-base HEAD origin/main').toString().trim()
let COMPARE_REV = ''

const ROOT_PATH = path.join(__dirname, '../../../')
const STATIC_ASSETS_PATH = process.env.WEB_BUNDLE_PATH || path.join(ROOT_PATH, 'ui/assets')
const statoscopeBinPath = path.join(ROOT_PATH, 'node_modules/.bin/statoscope')

const {
    GOOGLE_APPLICATION_CREDENTIALS,
    BRANCH,
    BUILDKITE_PULL_REQUEST_REPO,
    BUILDKITE_PULL_REQUEST,
    COMMIT,
    GITHUB_TOKEN,
} = process.env

console.log({ GOOGLE_APPLICATION_CREDENTIALS })

async function getCompareRev(): Promise<string | undefined> {
    const revisions = exec(`git --no-pager log "${MERGE_BASE}" --pretty=format:"%H" -n 20`).toString().split('\n')

    console.log('mergeBase', MERGE_BASE)
    console.log('revisions', revisions)

    for (const revision of revisions) {
        try {
            const tarPath = path.join(STATIC_ASSETS_PATH, `bundle_size_cache-${revision}.tar.gz`)
            const bucket = 'sourcegraph_buildkite_cache'
            const file = `sourcegraph/sourcegraph/bundle_size_cache-${revision}.tar.gz`

            exec(`gsutil -q cp -r "gs://${bucket}/${file}" "${tarPath}"`)

            console.log(`Found cached archive for ${revision}`, tarPath)
            COMPARE_REV = revision

            return tarPath
        } catch (error) {
            console.log(error)
            console.log(`Cached archive for ${revision} not found, skipping.`)
        }
    }

    return undefined
}

async function prepareStats(): Promise<{ commitFile: string; compareFile: string } | undefined> {
    const tarPath = await getCompareRev()

    if (tarPath) {
        console.log('---------')
        exec(`tar tvf ${tarPath}`)
        console.log('---------')

        const cmd = `tar -xf ${tarPath} --strip-components=2 -C ${STATIC_ASSETS_PATH}`
        console.log(cmd)
        exec(cmd)
        console.log('---------')

        exec(`ls -la ${STATIC_ASSETS_PATH}`)
        console.log('---------')

        const commitFile = path.join(STATIC_ASSETS_PATH, `stats-${COMMIT}.json`)
        const compareFile = path.join(STATIC_ASSETS_PATH, `stats-${COMPARE_REV}.json`)
        console.log({ commitFile, compareFile })

        try {
            await fs.access(commitFile)
            await fs.access(compareFile)

            const compareReportPath = path.join(STATIC_ASSETS_PATH, 'compare-report.html')

            exec(`${statoscopeBinPath} generate -i "${commitFile}" -r "${compareFile}" -t ${compareReportPath}`)

            const bucket = 'sourcegraph_reports'
            const file = `statoscope-reports/${BRANCH}/compare-report.html`
            exec(`gsutil cp ${compareReportPath} "gs://${bucket}/${file}"`)

            return { commitFile, compareFile }
        } catch (error) {
            console.log('No stats file found, skipping.', error)
            console.log(commitFile)
            console.log(compareFile)
        }
    }

    return undefined
}

async function main(): Promise<void> {
    try {
        const stats = await prepareStats()

        if (!stats) {
            return
        }

        const report = parseReport(stats.commitFile, stats.compareFile)

        if (hasZeroChanges(report)) {
            console.log('No changes detected in the bundle size, skip posting the comment.')
            process.exit(0)
        }

        const body = reportToMarkdown(report)
        await createOrUpdateComment(body)

        console.log(body)
    } catch (error) {
        console.error(error)
        process.exit(1)
    }
}

main().catch(error => {
    console.error(error)
    process.exit(1)
})

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

function parseReport(commitFile: string, compareFile: string): Report {
    const queryFile = path.join(__dirname, 'report-bundle-jora-query')
    const rawReport = exec(`cat "${queryFile}" | ${statoscopeBinPath} query -i "${compareFile}" -i "${commitFile}"`)

    return JSON.parse(rawReport) as Report
}

function reportToMarkdown(report: Report): string {
    const initialSizeMetric = report[1]
    const totalSizeMetric = report[2]
    const asyncSizeMetric = report[3]
    const modulesMetric = report[9]

    const initialSize = describeMetric(initialSizeMetric, 5000) // 5kb
    const totalSize = describeMetric(totalSizeMetric, 10000) // 10kb
    const asyncSize = describeMetric(asyncSizeMetric, 10000) // 10kb
    const modules = describeMetric(modulesMetric, 0)

    const url = `https://console.cloud.google.com/storage/browser/_details/sourcegraph_reports/statoscope-reports/${BRANCH}/compare-report.html;tab=live_object`

    let noExactDataWarning = ''
    if (MERGE_BASE !== COMPARE_REV) {
        noExactDataWarning = `
**Note:** We do not have exact data for ${shortRev(MERGE_BASE)}. So we have used data from: ${shortRev(COMPARE_REV)}.
The intended commit has no frontend pipeline, so we chose the last commit with one before it.`
    }

    return `${COMMENT_HEADING}

| Initial size | Total size | Async size | Modules |
| --- | --- | --- | --- |
| ${initialSize} | ${totalSize} | ${asyncSize} | ${modules} |

Look at the [Statoscope report](${url}) for a full comparison between the commits ${shortRev(COMMIT)} and ${shortRev(
        COMPARE_REV
    )} or [learn more](https://docs.sourcegraph.com/dev/how-to/testing#bundlesize).
${noExactDataWarning}

<details>
<summary>Open explanation</summary>

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
    const pullRequest = parseInt(BUILDKITE_PULL_REQUEST ?? '', 10)
    const [owner, _repo] =
        BUILDKITE_PULL_REQUEST_REPO?.replace('https://github.com/', '').replace('.git', '').split('/') ?? []
    const repo = { owner, repo: _repo }
    const octokit = new Octokit({ auth: GITHUB_TOKEN })
    if (!pullRequest || !owner || !_repo) {
        console.log(
            'No BUILDKITE_PULL_REQUEST or BUILDKITE_PULL_REQUEST_REPO env vars set, skip posting the following comment:'
        )
        console.log()
        console.log(body)
        return
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

function shortRev(rev: string | null | undefined): string {
    return rev ? rev.slice(0, 7) : 'unknown'
}

function hasZeroChanges(report: Report): boolean {
    for (const metric of report.slice(1) as Metric[]) {
        if (metric.value !== 0) {
            return false
        }
    }
    return true
}
