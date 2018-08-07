import * as OmniCLI from 'omnicli'

import { upsertSourcegraphUrl, URLError } from '../../extension/helpers/storage'
import storage from '../../extension/storage'

const upserUrl = (command: string) => ([url]: string[]) => {
    const err = upsertSourcegraphUrl(url)
    if (!err) {
        return
    }

    if (err === URLError.Empty || err === URLError.Invalid) {
        console.error(`src :${command} - invalid url entered`)
    } else if (err === URLError.HTTPNotSupported) {
        console.error(
            'Safari extensions do not support communication via `http:`. We suggest using https://ngrok.io for local testing.'
        )
    }
}

const addUrlCommand: OmniCLI.Command = {
    name: 'add-url',
    action: upserUrl('add-url'),
    description: 'Add a Sourcegraph Server URL',
}

function getSetURLSuggestions([cmd, ...args]: string[]): Promise<OmniCLI.Suggestion[]> {
    return new Promise(resolve => {
        storage.getSync(({ sourcegraphURL, serverUrls }) => {
            const suggestions: OmniCLI.Suggestion[] = serverUrls.map(url => ({
                content: url,
                description: `${url}${url === sourcegraphURL ? ' (current)' : ''}`,
            }))

            resolve(suggestions)
        })
    })
}

const setUrlCommand: OmniCLI.Command = {
    name: 'set-url',
    action: upserUrl('set-url'),
    getSuggestions: getSetURLSuggestions,
    description: 'Set your primary Sourcegraph Server URL',
}

function setFileTree([to]: string[]): void {
    if ((to && to === 'true') || to === 'false') {
        storage.setSync({
            repositoryFileTreeEnabled: to === 'true',
        })
        return
    }

    storage.getSync(({ repositoryFileTreeEnabled }) =>
        storage.setSync({ repositoryFileTreeEnabled: !repositoryFileTreeEnabled })
    )
}

function getSetFileTreeSuggestions(): Promise<OmniCLI.Suggestion[]> {
    return new Promise(resolve => {
        storage.getSync(({ repositoryFileTreeEnabled }) =>
            resolve([
                {
                    content: repositoryFileTreeEnabled ? 'false' : 'true',
                    description: `${repositoryFileTreeEnabled ? 'Disable' : 'Enable'} File Tree`,
                },
            ])
        )
    })
}

const setFileTreeCommand: OmniCLI.Command = {
    name: 'set-tree',
    action: setFileTree,
    getSuggestions: getSetFileTreeSuggestions,
    description: 'Set or toggle the File Tree',
}

function setOpenFileOn([to]: string[]): void {
    if ((to && to === 'true') || to === 'false') {
        storage.setSync({
            openFileOnSourcegraph: to === 'true',
        })
        return
    }

    storage.getSync(({ openFileOnSourcegraph }) => storage.setSync({ openFileOnSourcegraph: !openFileOnSourcegraph }))
}

function getSetOpenFileOnSuggestions(): Promise<OmniCLI.Suggestion[]> {
    return new Promise(resolve => {
        storage.getSync(({ openFileOnSourcegraph }) =>
            resolve([
                {
                    content: openFileOnSourcegraph ? 'false' : 'true',
                    description: `Open files from the fuzzy finder on ${
                        openFileOnSourcegraph ? 'your code host' : 'Sourcegraph'
                    }`,
                },
            ])
        )
    })
}

const setOpenFileOnCommand: OmniCLI.Command = {
    name: 'set-open-on-sg',
    alias: ['sof'],
    action: setOpenFileOn,
    getSuggestions: getSetOpenFileOnSuggestions,
    description: `Set whether you would like files to open on Sourcegraph of the given repo's code host`,
}

export default [addUrlCommand, setUrlCommand, setFileTreeCommand, setOpenFileOnCommand]
