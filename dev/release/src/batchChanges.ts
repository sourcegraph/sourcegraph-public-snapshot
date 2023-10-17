import commandExists from 'command-exists'
import execa from 'execa'
import fetch from 'node-fetch'
import YAML from 'yaml'

import type { CreatedChangeset } from './github'
import { readLine, cacheFolder } from './util'

// https://handbook.sourcegraph.com/departments/engineering/dev/process/deployments/instances/#sourcegraphsourcegraphcom-s2
const DEFAULT_SRC_ENDPOINT = 'https://sourcegraph.sourcegraph.com'

interface SourcegraphCLIConfig {
    SRC_ENDPOINT: string
    SRC_ACCESS_TOKEN: string
    [index: string]: string
}

/**
 * Retrieves src-cli configuration and ensures src-cli exists.
 */
export async function sourcegraphCLIConfig(): Promise<SourcegraphCLIConfig> {
    await commandExists('src') // CLI must be present for batch change interactions
    return {
        SRC_ENDPOINT: DEFAULT_SRC_ENDPOINT,
        // I updated the file name here to avoid a situation where folks with existing s2 token
        // cached will get a 403 because it's not valid for S2.
        SRC_ACCESS_TOKEN: await readLine('s2 src-cli token: ', `${cacheFolder}/src-cli-s2.txt`),
    }
}

/**
 * Parameters defining a batch change to interact with.
 *
 * Generate `cliConfig` using `sourcegraphCLIConfig()`.
 */
export interface BatchChangeOptions {
    name: string
    namespace: string
    cliConfig: SourcegraphCLIConfig
}

/**
 * Generate batch change configuration for a given release.
 */
export function releaseTrackingBatchChange(version: string, cliConfig: SourcegraphCLIConfig): BatchChangeOptions {
    return {
        name: `release-sourcegraph-${version}`,
        namespace: 'sourcegraph',
        cliConfig,
    }
}

/**
 * Generates a URL for a batch change that would be created under the given camapign options.
 *
 * Does not ensure the batch change exists.
 */
export function batchChangeURL(options: BatchChangeOptions): string {
    return `${options.cliConfig.SRC_ENDPOINT}/organizations/${options.namespace}/batch-changes/${options.name}`
}

/**
 * Create a new batch change from a set of changes.
 */
export async function createBatchChange(
    changes: CreatedChangeset[],
    options: BatchChangeOptions,
    description: string
): Promise<void> {
    // create a batch change spec
    const importChangesets = changes.map(change => ({
        repository: `github.com/${change.repository}`,
        externalIDs: [change.pullRequestNumber],
    }))
    // apply batch change
    // eslint-disable-next-line @typescript-eslint/return-await
    return await applyBatchChange(
        {
            name: options.name,
            description,
            importChangesets,
        },
        options
    )
}

/**
 * Append changes to an existing batch change.
 */
export async function addToBatchChange(
    changes: { repository: string; pullRequestNumber: number }[],
    options: BatchChangeOptions
): Promise<void> {
    const response = await fetch(`${options.cliConfig.SRC_ENDPOINT}/.api/graphql`, {
        method: 'POST',
        headers: {
            Authorization: `token ${options.cliConfig.SRC_ACCESS_TOKEN}`,
        },
        body: JSON.stringify({
            query: `query getBatchChanges($namespace:String!) {
                organization(name:$namespace) {
                  batchChanges(first:99) {
                    nodes { name currentSpec { originalInput } }
                  }
                }
              }`,
            variables: {
                namespace: options.namespace,
            },
        }),
    })
    const {
        data: {
            organization: {
                batchChanges: { nodes: results },
            },
        },
    } = (await response.json()) as {
        data: { organization: { batchChanges: { nodes: { name: string; currentSpec: { originalInput: string } }[] } } }
    }
    const batchChange = results.find(result => result.name === options.name)
    if (!batchChange) {
        throw new Error(`Cannot find batch change ${options.name}`)
    }

    const importChangesets = changes.map(change => ({
        repository: `github.com/${change.repository}`,
        externalIDs: [change.pullRequestNumber],
    }))
    const newSpec = YAML.parse(batchChange.currentSpec.originalInput) as BatchChangeSpec
    newSpec.importChangesets.push(...importChangesets)
    await applyBatchChange(newSpec, options)
}

/**
 * Subset of batch change spec: https://docs.sourcegraph.com/batch_changes/references/batch_spec_yaml_reference
 */
interface BatchChangeSpec {
    name: string
    description: string
    importChangesets: { repository: string; externalIDs: number[] }[]
}

async function applyBatchChange(batchChange: BatchChangeSpec, options: BatchChangeOptions): Promise<void> {
    const batchChangeYAML = YAML.stringify(batchChange)
    console.log(`Rendered batch change spec:\n\n${batchChangeYAML}`)

    // apply the batch change
    await execa('src', ['batch', 'apply', '-namespace', options.namespace, '-f', '-'], {
        stdout: 'inherit',
        input: batchChangeYAML,
        env: options.cliConfig,
    })
}
