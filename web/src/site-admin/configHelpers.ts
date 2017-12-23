import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import { Edit, FormattingOptions } from '@sqs/jsonc-parser/lib/format'
import { GitHubConnection, Repository } from '../schema/site.schema'

/**
 * A helper function that modifies site configuration to configure specific
 * common things, such as syncing GitHub repositories.
 */
type ConfigHelper = (configJSON: string) => { edits: Edit[]; selectText?: string }

const defaultFormattingOptions: FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

const addGitHubDotCom: ConfigHelper = config => {
    const tokenPlaceholder = '<personal access token with repo scope>'
    const value: GitHubConnection = {
        token: tokenPlaceholder,
        url: 'https://github.com',
    }
    const edits = setProperty(config, ['github', -1], value, defaultFormattingOptions)
    return { edits, selectText: tokenPlaceholder }
}

const addGitHubEnterprise: ConfigHelper = config => {
    const tokenPlaceholder = '<personal access token with repo scope>'
    const value: GitHubConnection = {
        token: tokenPlaceholder,
        url: 'https://github-enterprise-hostname.example.com',
    }
    const edits = setProperty(config, ['github', -1], value, defaultFormattingOptions)
    return { edits, selectText: tokenPlaceholder }
}

const addOtherRepository: ConfigHelper = config => {
    const urlPlaceholder = '<git clone URL>'
    const value: Repository = {
        url: urlPlaceholder,
        path: '<desired name of repository on Sourcegraph (example: my/repo)>',
    }
    const edits = setProperty(config, ['repos.list', -1], value, defaultFormattingOptions)
    return { edits, selectText: urlPlaceholder }
}

export interface EditorAction {
    id: string
    label: string
    run: ConfigHelper
}

export const editorActions: EditorAction[] = [
    { id: 'sourcegraph.site.githubDotCom', label: 'Add GitHub.com repositories', run: addGitHubDotCom },
    {
        id: 'sourcegraph.site.githubEnterprise',
        label: 'Add GitHub.com Enterprise repositories',
        run: addGitHubEnterprise,
    },
    { id: 'sourcegraph.site.otherRepository', label: 'Add other repository', run: addOtherRepository },
]
