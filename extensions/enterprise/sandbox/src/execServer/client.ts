import * as sourcegraph from 'sourcegraph'
import { memoizeAsync } from '../util'

export type ExecServerClient = ({
    commands,
    context,
    label,
}: Pick<Params, 'commands' | 'dir'> &
    Pick<Payload, 'files'> & { context?: RepositoryContext; label: string }) => Promise<Result>

export interface RepositoryContext {
    repository: string
    commit: string
}

interface Request {
    params: Params
    payload?: Payload
}

interface Params {
    archiveURL?: string
    dir?: string
    commands: string[][]
    includeFiles: string[]
}

interface Payload {
    files?: { [path: string]: string }
}

export interface Result {
    commands: {
        combinedOutput: string
        ok: boolean
        error?: string
    }[]
    files?: { [path: string]: string }
    fileDiffs?: { [path: string]: string }
}

export const createExecServerClient = (
    containerName: string,
    includeFiles: Params['includeFiles'] = [],
    cache = true
): ExecServerClient => {
    const baseUrl = new URL(`/.api/extension-containers/${containerName}`, sourcegraph.internal.sourcegraphURL)

    const do2: ExecServerClient = async ({ commands, dir, files, context, label }) => {
        const request: Request = {
            params: {
                archiveURL: context ? getPublicRepoArchiveUrl(context.repository, context.commit) : undefined,
                commands,
                dir,
                includeFiles,
            },
            payload: { files },
        }
        const hasPayload = Boolean(
            request.payload && request.payload.files && Object.keys(request.payload.files).length > 0
        )

        const url = new URL('', baseUrl)
        url.searchParams.set('params', JSON.stringify(request.params))
        url.hash = `#${label}`

        // console.debug('%cexec%c', 'background-color:blue;color:white', 'background-color:transparent;color:unset')
        const resp = await fetch(url.toString(), {
            headers: {
                'Content-Type': 'application/json; charset=utf-8',
            },
            ...(hasPayload
                ? {
                      method: 'POST',
                      body: JSON.stringify(request.payload),
                  }
                : {}),
        })
        if (!resp.ok) {
            throw new Error(`${label}: error executing commands on ${containerName}: HTTP ${resp.status}`)
        }
        const result: Result = await resp.json()
        for (const [i, command] of result.commands.entries()) {
            if (!command.ok) {
                throw new Error(
                    `${label}: error executing command ${JSON.stringify(commands[i])} on ${containerName}: ${
                        command.error
                    }\n${command.combinedOutput}`
                )
            }
        }
        return result
    }
    return cache ? memoizeAsync<Parameters<typeof do2>[0], Result>(do2, arg => JSON.stringify(arg)) : do2
}

function getPublicRepoArchiveUrl(repo: string, commit: string): string {
    const MAP: { [repo: string]: string } = {
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
        'github.com/sd9/openapi-generator': 'github.com/OpenAPITools/openapi-generator',
        'github.com/sd9/redbird': 'github.com/OptimalBits/redbird',
        'github.com/sd9org/react-router': 'github.com/ReactTraining/react-router',
        'github.com/sd9/react-loading-spinner': 'github.com/sourcegraph/react-loading-spinner',
        'github.com/sd9/graphql-dotnet': 'github.com/graphql-dotnet/graphql-dotnet',
        'github.com/sd9/taskbotjs': 'github.com/eropple/taskbotjs',
        'github.com/sd9/ReactStateMuseum': 'github.com/GantMan/ReactStateMuseum',
    }
    if (repo in MAP || sourcegraph.internal.sourcegraphURL.hostname === 'localhost') {
        return `https://sourcegraph.com/${MAP[repo] || repo}@${commit}/-/raw/`
    }
    // TODO!(sqs): requires token
    return new URL(`/${repo}@${commit}/-/raw/`, sourcegraph.internal.sourcegraphURL).toString()
}
