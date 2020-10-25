import { CreatedChangeset } from './github'
import { readLine } from './util'
import YAML from 'yaml'
import * as path from 'path'
import * as os from 'os'
import execa from 'execa'
import { promisify } from 'util'
import { writeFile as original_writeFile, mkdtemp as original_mkdtemp } from 'fs'

const mkdtemp = promisify(original_mkdtemp)
const writeFile = promisify(original_writeFile)

async function sourcegraphAuth(): Promise<NodeJS.ProcessEnv> {
    return {
        SRC_ENDPOINT: 'https://k8s.sgdev.org',
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
    // create a campaign
    const importChangesets: { repository: string; externalIDs: number[] }[] = changes.map(change => ({
        repository: `github.com/${change.repository}`,
        externalIDs: [change.pullRequestNumber],
    }))
    const campaignSpec = {
        name: options.name,
        description: options.description,
        importChangesets,
    }

    const tmpdir = await mkdtemp(path.join(os.tmpdir(), 'sg-campaigns'))
    const campaignYAML = YAML.stringify(campaignSpec)
    console.log(campaignYAML)
    await writeFile(path.join(tmpdir, 'campaign.yaml'), campaignYAML)
    const campaignScript = `set -ex

    src campaign ${options.preview ? 'preview' : 'apply'} \\
        -namespace ${options.namespace} \\
        -f campaign.yaml`
    await execa('bash', ['-c', campaignScript], { stdio: 'inherit', cwd: tmpdir, env: await sourcegraphAuth() })
    return ''
}
