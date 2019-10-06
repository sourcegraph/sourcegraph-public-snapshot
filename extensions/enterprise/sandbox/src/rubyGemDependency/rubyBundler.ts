import { RepositoryContext, Result, createExecServerClient } from '../execServer/client'

export const rubyBundlerExecClient = createExecServerClient('a8n-ruby-bundler-exec', [
    'Gemfile',
    'Gemfile.lock',
    'Gemfile.common',
    'Gemfile.modules',
])

/**
 * Returns the result of running `bundle remove <gemName>` in a repository tree.
 */
export const bundlerRemove = async (gemName: string, context: RepositoryContext): Promise<Result> => {
    // TODO!(sqs): run `bundle install --deployment` after to get updates to lockfile?
    const result = await rubyBundlerExecClient({
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
