import * as fs from 'fs'

const fetch = require('node-fetch')

interface Repo {
    name: string
    url: string
}

interface Embedding {
    id: string
    state: string
    repo: Repo
}

// const access_token = process.env.SOURCEGRAPH_DOCS_ACCESS_TOKEN
const access_token = 'sgp_3cd38e14441e074638882442f9762609624402e0'
const endpoint = 'https://sourcegraph.com/.api/graphql'

async function start() {
    try {
        let embeddedRepos = await gqlRequest(endpoint)
        embeddedRepos = filter(embeddedRepos)

        let markdown = embeddedReposToMarkdown(embeddedRepos)

        fs.writeFileSync('embedded-repos.md', markdown)
    } catch (err) {
        console.error(err)
    }
}

async function gqlRequest(endpoint: string): Promise<Embedding[]> {
    const embeddedRepos: Embedding[] = []
    try {
        let pagination = true
        let endCursor = ''

        while (pagination) {
            let query = `
                query RepoEmbeddingJobs {
                    repoEmbeddingJobs(first: 100, after: ${endCursor ? '"' + endCursor + '"' : null}) {
                    totalCount
                    pageInfo {
                        endCursor
                        hasNextPage
                    }
                    nodes {
                        id
                        state
                        repo {
                        name
                        url
                        }
                    }
                }
            }`
            const response = await fetch(endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `token ${access_token}`,
                },
                body: JSON.stringify({ query }),
            })
            const { data } = await response.json()
            embeddedRepos.push(...data.repoEmbeddingJobs.nodes)
            pagination = data.repoEmbeddingJobs.pageInfo.hasNextPage
            endCursor = data.repoEmbeddingJobs.pageInfo.endCursor
        }
    } catch (err) {
        console.error(err)
    }

    return embeddedRepos
}

export function filter(repos: Embedding[]): Embedding[] {
    const filtered = repos.filter(item => item.state === 'COMPLETED')
    const result = Array.from(new Set(filtered.map(x => x.repo?.name))).map(name =>
        filtered.find(x => x.repo?.name === name)
    ) as Embedding[]

    return result
}

export function sort(repos: Repo[]): Repo[] {
    // sort alphabetically
    repos.sort((a: Repo, b: Repo) => (a.name > b.name ? 1 : b.name > a.name ? -1 : 0))

    return repos
}

export function embeddedReposToMarkdown(repos: Embedding[] | undefined): string {
    const listOfRepos: Repo[] = []
    const today = new Date()

    let markdown = `# Embeddings for repositories with 5+ stars\n\n`
    markdown += `Last updated: ${today.toLocaleString('en-US', {
        month: '2-digit',
        day: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        timeZoneName: 'short',
    })} \n\n`

    repos?.forEach(repo => {
        let repoName: string | undefined = repo.repo?.name
        let repoUrl: string | undefined = repo.repo?.url
        if (repoName === undefined || repoUrl === undefined) {
            return
        }

        const r: Repo = {
            name: repo.repo?.name.replace('github.com/', ''),
            url: repoUrl.replace(/^\//, 'https://'),
        }
        listOfRepos.push(r)
    })

    if (listOfRepos.length === 0) {
        throw new Error('no embedded repos found!')
    }

    const sorted = sort(listOfRepos)
    sorted?.forEach(repo => {
        markdown += `1. [${repo?.name}](${repo?.url})\n`
    })

    return markdown
}

start()
