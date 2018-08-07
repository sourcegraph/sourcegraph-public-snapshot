import { first } from 'lodash'
import * as OmniCLI from 'omnicli'
import score from 'string-score'

import * as runtime from '../../extension/runtime'
import storage from '../../extension/storage'
import * as tabs from '../../extension/tabs'

import { buildSearchURLQuery, toBlobURL } from '../util/url'

function executeEnter(props: chrome.tabs.CreateProperties, disposition?: string): void {
    switch (disposition) {
        case 'newForegroundTab':
            tabs.create(props)
            break
        case 'newBackgroundTab':
            tabs.create({ ...props, active: false })
            break
        case 'currentTab':
        default:
            tabs.update(props)
            break
    }
}

interface Repo {
    name: string
    rev: string
    files: string[]
}

interface FileSuggestion extends OmniCLI.Suggestion {
    score: number
}

class FileCommand implements OmniCLI.Command {
    public name = 'file'
    public alias = ['f']

    private repos = new Map<string, Repo>()
    private openOnSourcegraph = true

    public setRepo(repo: Repo): void {
        this.repos.set(`${repo.name}@${repo.rev}`, repo)
    }

    public setOpenOnSourcegraph(val: boolean): void {
        this.openOnSourcegraph = val
    }

    public removeRepo(repo: Repo): void {
        this.repos.delete(repo.name)
    }

    public getSuggestions = ([queryPath]: string[]): OmniCLI.Suggestion[] => {
        if (this.repos.size === 0) {
            return [
                {
                    content: 'set-tree true',
                    description: 'Open a tab to a GitHub repo and enable the file tree to use the fuzzy file finder',
                },
            ]
        }

        const suggestions: FileSuggestion[] = []

        for (const repo of this.repos.values()) {
            suggestions.push(
                ...repo.files.map(path => ({
                    content: this.buildUrl(path, repo),
                    description: `${repo.name}@${repo.rev} - ${path}`,
                    score: queryPath ? score(path.toLowerCase(), queryPath.toLowerCase()) : 0,
                }))
            )
        }

        return suggestions
            .sort((a, b) => b.score - a.score)
            .map(({ content, description }) => ({ content, description }))
    }

    public action = ([raw, ...rest]: string[], disposition?: string) => {
        try {
            const url = new URL(raw)
            const props = { url: url.toString() }

            executeEnter(props, disposition)
            return
        } catch (e) {
            // If it's not a valid url, fall through to just searching for the input
            storage.getSync(({ sourcegraphURL, serverUrls }) => {
                const url = sourcegraphURL || first(serverUrls)
                const query = `type:path ${[raw, ...rest].join(' ').trim()}`
                const props = { url: `${url}/search?${buildSearchURLQuery(query)}` }

                executeEnter(props, disposition)
            })
        }
    }

    private buildUrl(path: string, repo: Repo): string {
        if (this.openOnSourcegraph) {
            return toBlobURL({
                repoPath: repo.name,
                rev: repo.rev,
                filePath: path,
            })
        }

        return `https://${repo.name}/blob/${repo.rev}/${path}`
    }
}

const cmd = new FileCommand()

tabs.query({ url: 'https://github.com/*' }, res => {
    for (const tab of res) {
        if (tab.id) {
            // TODO: Figure out a better way to get files from the already open repos and for non-github.com repos.
            //
            // Rather obtrusivly make sure we have files for tabs that could be a repo.
            // This will only ever happen when the background script first runs, so on install
            // or after re-enabling the extension.
            tabs.reload(tab.id)
        }
    }
})

runtime.onMessage(message => {
    const repo = message.payload as Repo
    if (message.type === 'fetched-files') {
        cmd.setRepo(repo)
    } else if (message.type === 'repo-closed') {
        cmd.removeRepo(repo)
    }
})

storage.getSync(({ openFileOnSourcegraph }) => {
    cmd.setOpenOnSourcegraph(openFileOnSourcegraph)
})

storage.onChanged(({ openFileOnSourcegraph }) => {
    if (openFileOnSourcegraph && openFileOnSourcegraph.newValue !== undefined) {
        cmd.setOpenOnSourcegraph(openFileOnSourcegraph.newValue)
    }
})

export default cmd
