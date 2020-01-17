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

export interface Upload {
    repository: string
    commit: string
    root: string
}

export async function performUploads(
    driver: Driver,
    config: Pick<Config, 'sourcegraphBaseUrl'>,
    gqlClient: GraphQLClient,
    repoBase: string,
    uploads: Upload[]
): Promise<void> {
    // First, remove all existing uploads for the repository
    for (const repository of new Set(uploads.map(u => u.repository))) {
        await clearUploads(gqlClient, `${repoBase}/${repository}`)
    }

    // Upload each upload in parallel and get back the upload status URLs
    const uploadUrls = await Promise.all(
        uploads.map(({ repository, commit, root }) =>
            performUpload(config, {
                repository: `${repoBase}/${repository}`,
                commit,
                root,
                filename: `lsif-data/${repository}@${commit.substring(0, 12)}.lsif`,
            })
        )
    )

    // Check the upload status URLs to ensure that they succeed, then ensure
    // that they are all listed as one of the "active" uploads for that repo
    for (const [i, { repository, ...rest }] of uploads.entries()) {
        await ensureUpload(driver, config, {
            ...rest,
            repository: `${repoBase}/${repository}`,
            uploadUrl: uploadUrls[i],
        })
    }
}

export async function uploadAndEnsure(
    driver: Driver,
    config: Pick<Config, 'sourcegraphBaseUrl'>,
    gqlClient: GraphQLClient,
    repoBase: string,
    repository: string,
    commit: string,
    root: string
): Promise<() => Promise<void>> {
    // First, remove all existing uploads for the repository
    await clearUploads(gqlClient, `${repoBase}/${repository}`)

    // Upload each upload in parallel and get back the upload status URLs
    const uploadUrl = await performUpload(config, {
        repository: `${repoBase}/${repository}`,
        commit,
        root,
        filename: `lsif-data/${repository}@${commit.substring(0, 12)}.lsif`,
    })

    // Check the upload status URLs to ensure that they succeed, then ensure
    // that they are all listed as one of the "active" uploads for that repo
    await ensureUpload(driver, config, {
        repository: `${repoBase}/${repository}`,
        commit,
        root,
        uploadUrl,
    })

    return (): Promise<void> => clearUploads(gqlClient, `${repoBase}/${repository}`)
}

//
// Helpers

async function clearUploads(gqlClient: GraphQLClient, repoName: string): Promise<void> {
    const { nodes, hasNextPage } = await gqlClient
        .queryGraphQL(
            gql`
                query ResolveRev($repoName: String!) {
                    repository(name: $repoName) {
                        lsifUploads {
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
                    : { nodes: repository.lsifUploads.nodes, hasNextPage: repository.lsifUploads.pageInfo.hasNextPage }
            )
        )
        .toPromise()

    const indices = range(nodes.length)
    const args: { [k: string]: string } = {}
    indices.forEach(i => (args[`upload${i}`] = nodes[i].id))

    await gqlClient
        .mutateGraphQL(
            gql`
                mutation(${indices.map(i => `$upload${i}: ID!`).join(', ')}) {
                    ${indices.map(i => gql`delete${i}: deleteLSIFUpload(id: $upload${i}) { alwaysNil }`).join('\n')}
                }
            `,
            args
        )
        .pipe(map(dataOrThrowErrors))
        .toPromise()

    if (hasNextPage) {
        // If we have more upload, clear the next page
        return clearUploads(gqlClient, repoName)
    }
}

async function performUpload(
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
    let out!: string
    try {
        // Untar the lsif data for this upload
        const tarCommand = ['tar', '-xzf', `${filename}.gz`, '-C', 'lsif-data'].join(' ')
        await child_process.exec(tarCommand, { cwd: path.join(__dirname, '..') })

        // Upload data
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

        throw new Error(`Failed to upload LSIF data: ${error.stderr || error.stdout || '(no output)'}`)
    }

    // Extract the status URL
    const match = out.match(/To check the status, visit (.+).\n$/)
    if (!match) {
        throw new Error(`Unexpected output from Sourcegraph cli: ${out}`)
    }

    return match[1]
}

async function ensureUpload(
    driver: Driver,
    config: Pick<Config, 'sourcegraphBaseUrl'>,
    { repository, commit, root, uploadUrl }: { repository: string; commit: string; root: string; uploadUrl: string }
): Promise<void> {
    const pendingUploadStateMessages = ['Upload is queued.', 'Upload is currently being processed...']

    await driver.page.goto(uploadUrl)
    while (true) {
        // Keep reloading upload page until the upload is terminal (not queued, not processed)
        const text = await (await driver.page.waitForSelector('.e2e-upload-state')).evaluate(elem => elem.textContent)
        if (!pendingUploadStateMessages.includes(text || '')) {
            break
        }

        await driver.page.reload()
    }

    // Ensure upload is successful
    const text = await (await driver.page.waitForSelector('.e2e-upload-state')).evaluate(elem => elem.textContent)
    expect(text).toEqual('Upload completed successfully.')

    await driver.page.goto(`${config.sourcegraphBaseUrl}/${repository}/-/settings/code-intelligence`)

    const commitElem = await driver.page.waitForSelector('.e2e-upload-commit')
    const actualCommit = await commitElem.evaluate(elem => elem.textContent)
    expect(actualCommit).toEqual(commit.substr(0, 7))

    const rootElem = await driver.page.waitForSelector('.e2e-upload-root')
    const actualRoot = await rootElem.evaluate(elem => elem.textContent)
    expect(actualRoot).toEqual(root)
}
