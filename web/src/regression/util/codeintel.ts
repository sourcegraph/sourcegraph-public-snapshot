import * as child_process from 'mz/child_process'
import * as path from 'path'
import { Config } from '../../../../shared/src/e2e/config'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import { Driver } from '../../../../shared/src/e2e/driver'
import { map } from 'rxjs/operators'
import { GraphQLClient } from './GraphQLClient'
import { range } from 'lodash'
import { applyEdits } from '@sqs/jsonc-parser'
import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import { getGlobalSettings } from './helpers'
import { overwriteSettings } from '../../../../shared/src/settings/edit'

async function setGlobalLSIFSetting(
    driver: Driver,
    gqlClient: GraphQLClient,
    enabled: boolean
): Promise<() => Promise<void>> {
    const { subjectID, settingsID, contents: oldContents } = await getGlobalSettings(gqlClient)
    const newContents = applyEdits(
        oldContents,
        setProperty(oldContents, ['codeIntel.lsif'], [enabled], {
            eol: '\n',
            insertSpaces: true,
            tabSize: 2,
        })
    )

    await overwriteSettings(gqlClient, subjectID, settingsID, newContents)
    return async () => {
        const { subjectID: currentSubjectID, settingsID: currentSettingsID } = await getGlobalSettings(gqlClient)
        await overwriteSettings(gqlClient, currentSubjectID, currentSettingsID, oldContents)
    }
}

export const enableLSIF = (driver: Driver, gqlClient: GraphQLClient): Promise<() => Promise<void>> =>
    setGlobalLSIFSetting(driver, gqlClient, true)

export const disableLSIF = (driver: Driver, gqlClient: GraphQLClient): Promise<() => Promise<void>> =>
    setGlobalLSIFSetting(driver, gqlClient, false)

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

export async function uploadAndEnsureDump(
    driver: Driver,
    config: Pick<Config, 'sourcegraphBaseUrl'>,
    gqlClient: GraphQLClient,
    repoBase: string,
    repository: string,
    commit: string,
    root: string
): Promise<() => Promise<void>> {
    // First, remove all existing dumps for the repository
    await clearDumps(gqlClient, `${repoBase}/${repository}`)

    // Upload each dump in parallel and get back the job status URLs
    const jobUrl = await uploadDump(config, {
        repository: `${repoBase}/${repository}`,
        commit,
        root,
        filename: `lsif-data/${repository}@${commit.substring(0, 12)}.lsif`,
    })

    // Check the job status URLs to ensure that they succeed, then ensure
    // that they are all listed as one of the "active" dumps for that repo
    await ensureDump(driver, config, {
        repository: `${repoBase}/${repository}`,
        commit,
        root,
        jobUrl,
    })

    return (): Promise<void> => clearDumps(gqlClient, `${repoBase}/${repository}`)
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
        try {
            // See if the error is due to a missing utility
            await child_process.exec('which src')
        } catch (error) {
            throw new Error('src-cli is not available on PATH')
        }

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
