import * as child_process from 'mz/child_process'
import * as path from 'path'
import { Config } from '../../../../shared/src/e2e/config'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import { Driver } from '../../../../shared/src/e2e/driver'
import { map } from 'rxjs/operators'
import { GraphQLClient } from './GraphQLClient'
import { range } from 'lodash'

async function setGlobalLSIFSetting(
    driver: Driver,
    config: Pick<Config, 'sourcegraphBaseUrl'>,
    enabled: boolean
): Promise<void> {
    await driver.page.goto(`${config.sourcegraphBaseUrl}/site-admin/global-settings`)
    const globalSettings = `{"codeIntel.lsif": ${enabled}}`
    await driver.replaceText({
        selector: '.monaco-editor',
        newText: globalSettings,
        selectMethod: 'keyboard',
        enterTextMethod: 'type',
    })
    await (
        await driver.findElementWithText('Save changes', {
            selector: 'button',
            wait: { timeout: 500 },
        })
    ).click()

    await driver.page.waitForFunction(
        () => document.querySelectorAll('[title="No changes to save or discard"]').length > 0
    )
}

export const enableLSIF = (driver: Driver, config: Pick<Config, 'sourcegraphBaseUrl'>): Promise<void> =>
    setGlobalLSIFSetting(driver, config, true)

export const disableLSIF = (driver: Driver, config: Pick<Config, 'sourcegraphBaseUrl'>): Promise<void> =>
    setGlobalLSIFSetting(driver, config, false)

export interface Dump {
    repository: string
    commit: string
    root: string
}

export async function uploadDumps(
    driver: Driver,
    config: Pick<Config, 'sourcegraphBaseUrl'>,
    gqlClient: GraphQLClient,
    repoBase: string,
    dumps: Dump[]
): Promise<void> {
    // First, remove all existing dumps for the repository
    for (const repository of new Set(dumps.map(d => d.repository))) {
        await clearDumps(gqlClient, `${repoBase}/${repository}`)
    }

    // Upload each dump in parallel and get back the job status URLs
    const jobUrls = await Promise.all(
        dumps.map(({ repository, commit, root }) =>
            uploadDump(config, {
                repository: `${repoBase}/${repository}`,
                commit,
                root,
                filename: `lsif-data/${repository}@${commit.substring(0, 12)}.lsif`,
            })
        )
    )

    // Check the job status URLs to ensure that they succeed, then ensure
    // that they are all listed as one of the "active" dumps for that repo
    for (const [i, { repository, ...rest }] of dumps.entries()) {
        await ensureDump(driver, config, {
            ...rest,
            repository: `${repoBase}/${repository}`,
            jobUrl: jobUrls[i],
        })
    }
}

//
// Helpers

async function clearDumps(gqlClient: GraphQLClient, repoName: string): Promise<void> {
    const { nodes, hasNextPage } = await gqlClient
        .queryGraphQL(
            gql`
                query ResolveRev($repoName: String!) {
                    repository(name: $repoName) {
                        lsifDumps {
                            nodes {
                                id
                            }

                            pageInfo {
                                hasNextPage
                            }
                        }
                    }
                }
            `,
            { repoName }
        )
        .pipe(
            map(dataOrThrowErrors),
            map(({ repository }) =>
                repository === null
                    ? { nodes: [], hasNextPage: false }
                    : { nodes: repository.lsifDumps.nodes, hasNextPage: repository.lsifDumps.pageInfo.hasNextPage }
            )
        )
        .toPromise()

    const indices = range(nodes.length)
    const args: { [k: string]: string } = {}
    indices.forEach(i => (args[`dump${i}`] = nodes[i].id))

    await gqlClient
        .mutateGraphQL(
            gql`
                mutation(${indices.map(i => `$dump${i}: ID!`).join(', ')}) {
                    ${indices.map(i => gql`delete${i}: deleteLSIFDump(id: $dump${i}) { alwaysNil }`).join('\n')}
                }
            `,
            args
        )
        .pipe(map(dataOrThrowErrors))
        .toPromise()

    if (hasNextPage) {
        // If we have more dumps, clear the next page
        return clearDumps(gqlClient, repoName)
    }
}

async function uploadDump(
    config: Pick<Config, 'sourcegraphBaseUrl'>,
    {
        repository,
        commit,
        root,
        filename,
    }: {
        repository: string
        commit: string
        root: string
        filename: string
    }
): Promise<string> {
    let out!: Buffer
    try {
        // Untar the lsif data for this upload
        const tarCommand = ['tar', '-xzf', `${filename}.gz`, '-C', 'lsif-data'].join(' ')
        await child_process.exec(tarCommand, { cwd: path.join(__dirname, '..') })

        // Upload the dump
        const uploadCommand = [
            `src -endpoint ${config.sourcegraphBaseUrl}`,
            'lsif upload',
            `-repo ${repository}`,
            `-commit ${commit}`,
            `-root ${root}`,
            `-file ${filename}`,
        ].join(' ')
        ;[out] = await child_process.exec(uploadCommand, { cwd: path.join(__dirname, '..') })
    } catch (error) {
        throw new Error(`Failed to upload LSIF dump: ${error.stderr || error.stdout || '(no output)'}`)
    }

    // Extract the status URL
    const match = out.toString().match(/To check the status, visit (.+).\n$/)
    if (!match) {
        throw new Error(`Unexpected output from Sourcegraph cli: ${out.toString()}`)
    }

    return match[1]
}

async function ensureDump(
    driver: Driver,
    config: Pick<Config, 'sourcegraphBaseUrl'>,
    { repository, commit, root, jobUrl }: { repository: string; commit: string; root: string; jobUrl: string }
): Promise<void> {
    const pendingJobStateMessages = ['Job is queued.', 'Job is currently being processed...']

    await driver.page.goto(jobUrl)
    while (true) {
        // Keep reloading job page until the job is terminal (not queued, not processed)
        const text = await (await driver.page.waitForSelector('.e2e-job-state')).evaluate(elem => elem.textContent)
        if (!pendingJobStateMessages.includes(text || '')) {
            break
        }

        await driver.page.reload()
    }

    // Ensure job is successful
    const text = await (await driver.page.waitForSelector('.e2e-job-state')).evaluate(elem => elem.textContent)
    expect(text).toEqual('Job completed successfully.')

    await driver.page.goto(`${config.sourcegraphBaseUrl}/${repository}/-/settings/code-intelligence`)

    const commitElem = await driver.page.waitForSelector('.e2e-dump-commit')
    const actualCommit = await commitElem.evaluate(elem => elem.textContent)
    expect(actualCommit).toEqual(commit.substr(0, 7))

    const rootElem = await driver.page.waitForSelector('.e2e-dump-path')
    const actualRoot = await rootElem.evaluate(elem => elem.textContent)
    expect(actualRoot).toEqual(root)
}
