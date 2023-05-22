import { Octokit } from 'octokit'

import { ENVIRONMENT_CONFIG } from '../constants'

const octokit = new Octokit({ auth: ENVIRONMENT_CONFIG.GITHUB_TOKEN })

interface FetchFileContentOptions {
    owner: string
    repo: string
    path: string
}

export async function fetchFileContent(options: FetchFileContentOptions) {
    const { owner, repo, path } = options

    try {
        const response = await octokit.rest.repos.getContent({ owner, repo, path })
        if ('type' in response.data && response.data.type === 'file') {
            const content = Buffer.from(response.data.content, 'base64').toString('utf8')
            return {
                content,
                url: response.data.html_url,
            }
        }

        console.error('Unexpected response fetching file from GitHub:', response)
    } catch (error) {
        console.error('Error fetching file from GitHub!', error)
    }

    return undefined
}
