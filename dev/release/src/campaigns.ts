import { CreatedChangeset } from './github'
import { readLine } from './util'
import YAML from 'yaml'
import execa from 'execa'
import fetch from 'node-fetch'

// https://about.sourcegraph.com/handbook/engineering/deployments/instances#k8s-sgdev-org
const DEFAULT_SRC_ENDPOINT = 'https://k8s.sgdev.org'

export interface SourcegraphAuth {
    SRC_ENDPOINT: string
    SRC_ACCESS_TOKEN: string
    [index: string]: string
}

export async function sourcegraphAuth(): Promise<SourcegraphAuth> {
    return {
        SRC_ENDPOINT: DEFAULT_SRC_ENDPOINT,
        SRC_ACCESS_TOKEN: await readLine('k8s.sgdev.org src-cli token: ', '.secrets/src-cli.txt'),
    }
}

export interface CampaignOptions {
    name: string
    description: string
    namespace: string
    auth: SourcegraphAuth
}

export function releaseTrackingCampaign(version: string, auth: SourcegraphAuth): CampaignOptions {
    return {
        name: `release-sourcegraph-${version}`,
        description: `Track publishing of sourcegraph@${version}`,
        namespace: 'sourcegraph',
        auth,
    }
}

interface CampaignSpec {
    name: string
    description: string
    importChangesets: { repository: string; externalIDs: number[] }[]
}

async function applyCampaign(campaign: CampaignSpec, options: CampaignOptions): Promise<string> {
    const campaignYAML = YAML.stringify(campaign)
    console.log(`Rendered campaign spec:\n\n${campaignYAML}`)

    // apply the campaign
    await execa('src', ['campaign', 'apply', '-namespace', options.namespace, '-f', '-'], {
        stdout: 'inherit',
        input: campaignYAML,
        env: options.auth,
    })

    // return the campaign URL
    return `${options.auth.SRC_ENDPOINT}/organizations/${options.namespace}/campaigns/${options.name}`
}

export async function createCampaign(changes: CreatedChangeset[], options: CampaignOptions): Promise<string> {
    // create a campaign spec
    const importChangesets = changes.map(change => ({
        repository: `github.com/${change.repository}`,
        externalIDs: [change.pullRequestNumber],
    }))
    // apply campaign
    return await applyCampaign(
        {
            name: options.name,
            description: options.description,
            importChangesets,
        },
        options
    )
}

export async function addToCampaign(
    changes: { repository: string; pullRequestNumber: number }[],
    options: CampaignOptions
): Promise<string> {
    const response = await fetch(`${options.auth.SRC_ENDPOINT}/.api/graphql`, {
        method: 'POST',
        headers: {
            Authorization: `token ${options.auth.SRC_ACCESS_TOKEN}`,
        },
        body: JSON.stringify({
            query: `query getCampaigns($namespace:String!) {
                organization(name:$namespace) {
                  campaigns(first:99) {
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
                campaigns: { nodes: results },
            },
        },
    } = (await response.json()) as {
        data: { organization: { campaigns: { nodes: { name: string; currentSpec: { originalInput: string } }[] } } }
    }
    const campaign = results.find(result => result.name === options.name)
    if (!campaign) {
        throw new Error(`Cannot find campaign ${options.name}`)
    }

    const importChangesets = changes.map(change => ({
        repository: `github.com/${change.repository}`,
        externalIDs: [change.pullRequestNumber],
    }))
    const newSpec = YAML.parse(campaign.currentSpec.originalInput) as CampaignSpec
    newSpec.importChangesets.push(...importChangesets)
    return await applyCampaign(newSpec, options)
}
