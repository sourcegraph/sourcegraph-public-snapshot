import { CreatedChangeset } from './github'
import { readLine } from './util'
import YAML from 'yaml'
import execa from 'execa'

// https://about.sourcegraph.com/handbook/engineering/deployments/instances#k8s-sgdev-org
const DEFAULT_SRC_ENDPOINT = 'https://k8s.sgdev.org'

async function sourcegraphAuth(): Promise<NodeJS.ProcessEnv> {
    return {
        SRC_ENDPOINT: DEFAULT_SRC_ENDPOINT,
        SRC_ACCESS_TOKEN: await readLine('k8s.sgdev.org src-cli token: ', '.secrets/src-cli.txt'),
    }
}

export interface CampaignOptions {
    name: string
    description: string
    namespace: string
    preview?: boolean
}

export async function importFromCreatedChanges(changes: CreatedChangeset[], options: CampaignOptions): Promise<string> {
    // create a campaign spec
    const importChangesets: { repository: string; externalIDs: number[] }[] = changes.map(change => ({
        repository: `github.com/${change.repository}`,
        externalIDs: [change.pullRequestNumber],
    }))
    const campaignSpec = {
        name: options.name,
        description: options.description,
        importChangesets,
    }

    // apply campaign
    const campaignYAML = YAML.stringify(campaignSpec)
    console.log(`Rendered campaign spec:\n\n${campaignYAML}`)
    const campaignScript = `set -ex

(
cat <<EOF
${campaignYAML}
EOF
) | src campaign ${options.preview ? 'preview' : 'apply'} \\
    -namespace ${options.namespace} \\
    -f -`
    await execa('bash', ['-c', campaignScript], { stdio: 'inherit', env: await sourcegraphAuth() })

    return ''
}
