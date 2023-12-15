/* eslint-disable no-console, no-sync */

import { execSync } from 'child_process'
import fs from 'fs'
import path from 'path'

import { Octokit } from 'octokit'

const COMMENT_HEADING = '## Bundle size report 📦'

const { BUILDKITE_COMMIT, BUILDKITE_BRANCH, BUILDKITE_PULL_REQUEST_REPO, BUILDKITE_PULL_REQUEST, GH_TOKEN } =
    process.env

console.log('report-bundle-diff env:', {
    BUILDKITE_COMMIT,
    BUILDKITE_BRANCH,
    BUILDKITE_PULL_REQUEST_REPO,
    BUILDKITE_PULL_REQUEST,
    GH_TOKEN,
})

const ROOT_PATH = path.join(__dirname, '../../../')
const STATIC_ASSETS_PATH = path.join(ROOT_PATH, process.env.WEB_BUNDLE_PATH || 'client/web/dist')
const STATOSCOPE_BIN = path.join(ROOT_PATH, 'node_modules/@statoscope/cli/bin/cli.js')

const MERGE_BASE = execSync('git merge-base HEAD origin/main').toString().trim()
let COMPARE_REV = ''

async function findFile(root: string, filename: string): Promise<string> {
    // file can be in one of 2 base paths
    const parts: string[] = ['enterprise', '']
    const files = await Promise.all(
        parts.flatMap(async (dir: string) => {
            const filePath = path.join(root, dir, filename)
            try {
                await fs.promises.access(filePath)
                return filePath
            } catch {
                return ''
            }
        })
    )

    const foundFile = files.reduce((accumulator: string, possibleFile: string): string => {
        if (possibleFile) {
            return possibleFile
        }
        return accumulator
    })

    if (!foundFile) {
        throw new Error(`"${filename} not found under root ${root}`)
    }

    return foundFile
}

/**
 * We may not have a stats.json file for the merge base commit as these are only
 * created for commits that touch frontend files. Instead, we scan for 20 commits
 * before the merge base and use the latest stats.json file we find.
 */
function getTarPath(): string | undefined {
    console.log('--- Find a commit to compare the bundle size against')
    const revisions = execSync(`git --no-pager log "${MERGE_BASE}" --pretty=format:"%H" -n 20`).toString().split('\n')

    for (const revision of revisions) {
        try {
            const tarPath = path.join(STATIC_ASSETS_PATH, `bundle_size_cache-${revision}.tar.gz`)
            const bucket = 'sourcegraph_buildkite_cache'
            const file = `sourcegraph/sourcegraph/bundle_size_cache-${revision}.tar.gz`

            execSync(`gsutil -q cp -r "gs://${bucket}/${file}" "${tarPath}"`)

            // gsutil doesn't exit with a non-zero exit code when the file is not found.
            if (!fs.existsSync(tarPath)) {
                throw new Error('gsutil failed to copy the file.')
            }

            console.log(`Found cached archive for ${revision}:`, tarPath)
            // TODO: remove mutable global variable
            COMPARE_REV = revision

            return tarPath
        } catch (error) {
            console.log(`Cached archive for ${revision} not found:`, error)
        }
    }

    return undefined
}

async function prepareStats(): Promise<{ commitFile: string; compareFile: string } | undefined> {
    const tarPath = getTarPath()

    if (tarPath) {
        execSync(`tar -xf ${tarPath} --strip-components=2 -C ${STATIC_ASSETS_PATH}`)
        execSync(`ls -la ${STATIC_ASSETS_PATH}`)

        try {
            const commitFile = await findFile(STATIC_ASSETS_PATH, `stats-${BUILDKITE_COMMIT}.json`)
            const compareFile = await findFile(STATIC_ASSETS_PATH, `stats-${COMPARE_REV}.json`)
            console.log({ commitFile, compareFile })

            const compareReportPath = path.join(STATIC_ASSETS_PATH, 'compare-report.html')

            execSync(`${STATOSCOPE_BIN} generate -i "${commitFile}" -r "${compareFile}" -t ${compareReportPath}`)

            const bucket = 'sourcegraph_reports'
            const file = `statoscope-reports/${BUILDKITE_BRANCH}/compare-report.html`
            execSync(`gsutil cp ${compareReportPath} "gs://${bucket}/${file}"`)

            return { commitFile, compareFile }
        } catch (error) {
            console.error('Failed to prepare stats:', error)
            process.exit(1)
        }
    }

    return undefined
}

async function main(): Promise<void> {
    try {
        const stats = await prepareStats()

        if (!stats) {
            console.log('Failed to find stats to compare the bundle size against.')
            process.exit(0)
        }

        console.log('--- Report bundle diff')
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
    const rawReport = execSync(`cat "${queryFile}" | ${STATOSCOPE_BIN} query -i "${compareFile}" -i "${commitFile}"`, {
        encoding: 'utf8',
    })

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

    const url = `https://console.cloud.google.com/storage/browser/_details/sourcegraph_reports/statoscope-reports/${BUILDKITE_BRANCH}/compare-report.html;tab=live_object`

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

Look at the [Statoscope report](${url}) for a full comparison between the commits ${shortRev(
        BUILDKITE_COMMIT
    )} and ${shortRev(COMPARE_REV)} or [learn more](https://docs.sourcegraph.com/dev/how-to/testing#bundlesize).
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
    const octokit = new Octokit({ auth: GH_TOKEN })

    if (!pullRequest || !owner || !_repo) {
        console.log(
            'No BUILDKITE_PULL_REQUEST or BUILDKITE_PULL_REQUEST_REPO env vars set, skip posting the following comment:'
        )
        console.log()
        console.log(body)
        return
    }

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
