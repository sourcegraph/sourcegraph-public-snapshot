import type { SemVer } from 'semver'

import { type ReleaseConfig, setAWSExecutorVersion, setGoogleExecutorVersion, setSrcCliVersion } from './config'
import { cloneRepo, createChangesets, type Edit, getAuthenticatedGitHubClient, releaseBlockerLabel } from './github'
import {
    nextAWSExecutorVersionInputWithAutodetect,
    nextGoogleExecutorVersionInputWithAutodetect,
    nextSrcCliVersionInputWithAutodetect,
    pullRequestBody,
} from './util'

export async function bakeSrcCliSteps(config: ReleaseConfig): Promise<Edit[]> {
    const client = await getAuthenticatedGitHubClient()

    // ok, this seems weird that we're cloning src-cli here, so read on -
    // We have docs that live in the main src/src repo about src-cli. Each version we update these docs for any changes
    // from the most recent version of src-cli. Cool, makes sense.
    // The thing is that these docs are generated from src-cli itself (a literal command, src docs).
    // So our options are either to release a new version of src-cli, wait for the github action to be complete and THEN update the src/src repo,
    // OR we can assume that main is going to be the new version (which it is). So we will clone it and execute the
    // commands against the binary directly, saving ourselves a lot of time.
    const { workdir } = await cloneRepo(client, 'sourcegraph', 'src-cli', {
        revision: 'main',
        revisionMustExist: true,
    })

    const next = await nextSrcCliVersionInputWithAutodetect(config, workdir)
    setSrcCliVersion(config, next.version)

    return [
        combyReplace('const MinimumVersion = ":[1]"', next.version, 'internal/src-cli/consts.go'),
        `cd ${workdir}/cmd/src && go build`,
        `cd doc/cli/references && go run ./doc.go --binaryPath="${workdir}/cmd/src/src"`,
    ]
}
export async function bakeAWSExecutorsSteps(config: ReleaseConfig): Promise<void> {
    const client = await getAuthenticatedGitHubClient()
    const { workdir } = await cloneRepo(client, 'sourcegraph', 'terraform-aws-executors', {
        revision: 'master',
        revisionMustExist: true,
    })

    const next = await nextAWSExecutorVersionInputWithAutodetect(config, workdir)
    setAWSExecutorVersion(config, next.version)
    console.log(next)

    const prDetails = {
        body: pullRequestBody(`Update files for ${next.version} release`),
        title: `executor: v${next.version}`,
        commitMessage: `executor: v${next.version}`,
    }
    const sets = await createChangesets({
        requiredCommands: [],
        changes: [
            {
                ...prDetails,
                repo: 'sourcegraph',
                owner: 'terraform-aws-executors',
                base: 'master',
                head: `release/prepare-${next.version}`,
                edits: [`./prepare-release.sh ${next.version}`],
                labels: [releaseBlockerLabel],
                draft: true,
            },
        ],
        dryRun: config.dryRun.changesets,
    })
    console.log('Merge the following pull requests:\n')
    for (const set of sets) {
        console.log(set.pullRequestURL)
    }
}

export async function bakeGoogleExecutorsSteps(config: ReleaseConfig): Promise<void> {
    const client = await getAuthenticatedGitHubClient()
    const { workdir } = await cloneRepo(client, 'sourcegraph', 'terraform-google-executors', {
        revision: 'master',
        revisionMustExist: true,
    })
    console.log(`Cloned sourcegraph/terraform-google-executors to ${workdir}`)

    const next = await nextGoogleExecutorVersionInputWithAutodetect(config, workdir)
    setGoogleExecutorVersion(config, next.version)

    const prDetails = {
        body: pullRequestBody(`Update files for ${next.version} release`),
        title: `executor: v${next.version}`,
        commitMessage: `executor: v${next.version}`,
    }
    const sets = await createChangesets({
        requiredCommands: [],
        changes: [
            {
                ...prDetails,
                repo: 'terraform-google-executors',
                owner: 'sourcegraph',
                base: 'master',
                head: `release/prepare-${next.version}`,
                edits: [`./prepare-release.sh ${next.version}`],
                labels: [releaseBlockerLabel],
                draft: true,
            },
        ],
        dryRun: config.dryRun.changesets,
    })
    console.log('Merge the following pull requests:\n')
    for (const set of sets) {
        console.log(set.pullRequestURL)
    }
}

export function batchChangesInAppChangelog(version: SemVer, resetShow: boolean): Edit[] {
    const path = 'client/web/src/enterprise/batches/list/BatchChangesChangelogAlert.tsx'
    const steps = [combyReplace("const CURRENT_VERSION = ':[1]'", `${version.major}.${version.minor}`, path)]
    if (resetShow) {
        steps.push(combyReplace('const SHOW_CHANGELOG = :[1]', 'false', path))
    }
    return steps
}

// given a comby pattern such as 'const MinimumVersion = ":[1]"' generate the comby expression to replace with provided substitution
export function combyReplace(pattern: string, replace: string, path: string): Edit {
    pattern = pattern.replaceAll('"', '\\"')
    const sub = pattern.replace(':[1]', replace)
    return `comby -in-place "${pattern}" "${sub}" ${path}`
}

export function indexerUpdate(): Edit {
    // eslint-disable-next-line no-template-curly-in-string
    return 'cd internal/codeintel/autoindexing/internal/inference/libs/ && DOCKER_USER=${CR_USERNAME} DOCKER_PASS=${CR_PASSWORD} ./update-shas.sh'
}
