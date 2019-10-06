import { memoizeAsync } from '../util'

export interface BundlerExecutionContext {
    repository: string
    commit: string
    path?: string
}

export interface BundlerExecutionResult {
    commands: {
        combinedOutput: string
        ok: boolean
        error?: string
    }[]
    files: { [path: string]: string }
}

const RUBY_BUNDLER_EXEC_SERVICE_URL = 'http://localhost:5151'
// const RUBY_BUNDLER_EXEC_SERVICE_URL= 'http://ruby-bundler-exec.default.knative.sqs-sandbox.sgdev.org'

/**
 * Executes a Ruby `bundle` command in a repository tree and returns the resulting contents of
 * `Gemfile` and `Gemfile.lock`.
 */
const executeBundlerCommandUncached = async ({
    commands,
    context,
}: {
    commands: string[][]
    context: BundlerExecutionContext
}): Promise<BundlerExecutionResult> => {
    const resp = await fetch(RUBY_BUNDLER_EXEC_SERVICE_URL, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json; charset=utf-8',
        },
        body: JSON.stringify({
            archiveURL: getPublicRepoArchiveUrl(context.repository, context.commit),
            commands,
            // TODO!(sqs): support changes to all Gemfile.* and anything else that `bundler remove` might touch
            includeFiles: ['Gemfile', 'Gemfile.lock', 'Gemfile.common', 'Gemfile.modules'],
        } as {
            archiveURL: string
            dir?: string
            commands: string[][]
            includeFiles: string[]
        }),
    })
    if (!resp.ok) {
        throw new Error(`error executing bundler command in ${context.repository}: HTTP ${resp.status}`)
    }
    const result: BundlerExecutionResult = await resp.json()
    return result
}

export const executeBundlerCommand = memoizeAsync<
    Parameters<typeof executeBundlerCommandUncached>[0],
    BundlerExecutionResult
>(arg => executeBundlerCommandUncached(arg), arg => JSON.stringify(arg))

/**
 * Returns the result of running `bundle remove <gemName>` in a repository tree.
 */
export const bundlerRemove = async (
    gemName: string,
    context: BundlerExecutionContext
): Promise<BundlerExecutionResult> => {
    // TODO!(sqs): run `bundle install --deployment` after to get updates to lockfile?
    const result = await executeBundlerCommand({
        commands: [['bundler', 'remove', '--', gemName], ['bundler', 'lock', '--conservative', '--local']],
        context,
    })
    for (const commandResult of result.commands) {
        if (!commandResult.ok && !commandResult.combinedOutput.includes('is not specified in')) {
            throw new Error(`error in bundlerRemove: ${commandResult.error}\n${commandResult.combinedOutput}`)
        }
    }
    return result
}

function getPublicRepoArchiveUrl(repo: string, commit: string): string {
    const MAP = {
        'AC/activeadmin': 'github.com/activeadmin/activeadmin',
        'ACTG/acts-as-taggable-on': 'github.com/mbleigh/acts-as-taggable-on',
        'AD/administrate': 'github.com/thoughtbot/administrate',
        'CAN/cancancan': 'github.com/CanCanCommunity/cancancan',
        'DEV/devise': 'github.com/plataformatec/devise',
        'DIS/discourse': 'github.com/discourse/discourse',
        'FAK/faker': 'github.com/faker-ruby/faker',
        'LIQ/liquid': 'github.com/Shopify/liquid',
        'LOG/logstash': 'github.com/elastic/logstash',
        'OP/openproject': 'github.com/opf/openproject',
        'SID/sidekiq': 'github.com/mperham/sidekiq',
        'SOL/solidus': 'github.com/solidusio/solidus',
        'SPREE/spree': 'github.com/spree/spree',
    }
    return `https://sourcegraph.com/${MAP[repo] || repo}@${commit}/-/raw/`
}
