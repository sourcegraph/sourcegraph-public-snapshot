import { SourcegraphCompletionsClient } from '../sourcegraph-api/completions/client'

import { IntentDetector } from '.'
import { getCurrentBranchName, getCurrentFilePath, getOpenFilePaths, getProjectName, getRepoName } from './helpers'

export class LlmBasedIntentDetector implements IntentDetector {
    private cache = new Map<string, string>()

    constructor(private completionsClient: SourcegraphCompletionsClient) {}

    public async isEditorContextRequired(input: string): Promise<boolean | Error> {
        if (!this.cache.has(input)) {
            const response = await this.determineContextNeeds(input)
            this.cache.set(input, response)
        }

        const response = this.cache.get(input)
        return !!response && response.includes('a')
    }

    public async isEditorBroaderFileContextRequired(input: string): Promise<boolean | Error> {
        if (!this.cache.has(input)) {
            const response = await this.determineContextNeeds(input)
            this.cache.set(input, response)
        }

        const response = this.cache.get(input)
        return !!response && response.includes('b')
    }

    public async isCodebaseContextRequired(input: string): Promise<boolean | Error> {
        if (!this.cache.has(input)) {
            const response = await this.determineContextNeeds(input)
            this.cache.set(input, response)
        }
        return this.cache.get(input) === 'c'
    }

    private async determineContextNeeds(input: string): Promise<string> {
        const projectName = getProjectName()
        const repoName = await getRepoName()
        const currentFilePath = getCurrentFilePath() // TODO: Use repo-relative path
        const branchName = await getCurrentBranchName()
        const openFilePathsList = getOpenFilePaths()
            .filter(p => p !== currentFilePath)
            .join('", "') // TODO: Use repo-relative paths

        const prompt = `\n\nHuman: I'm looking at my IDE. I have a project called "${projectName}" open, of repo "${repoName}", on branch ${branchName}.
The file I'm looking at is at "${currentFilePath}". I also have these files open: "${openFilePathsList}".
I have access to:
 - the full repo, and can provide context from all of these sources.
 - an amazing search engine to fetch some relevant file/snippets from the repo.

My question is always: What would be the most helpful from the following options?
 a.) the content of my currently open file
 b.) the contents of some of my other open files
 c.) search for code snippets with that cool search engine of mine
 d.) probably none of the above

Respond with one or more of the letters abcd, then any short explanation.

My first question is: What's in my current file?\n
Assistant: a - Only your current file is needed to answer this.\n
Human: What's in my open files?\n\nAssistant: ab - Would help to see your current and other open files to answer this question.\n
Human: Tell me where we do authentication\n\nAssistant: c - A smart search in the repo for auth-related files is needed.\n
Human: What day is it today?\n\nAssistant: d - This question seems unrelated to code altogether.\n
Human: ${input}\n\nAssistant: `

//        console.log('Asking for context needs...', prompt)

        const response = await this.completionsClient.complete({
            prompt,
            stopSequences: ['\n\nHuman:'],
            maxTokensToSample: 200,
            model: 'claude-instant-v1.0',
            temperature: 1, // default value (source: https://console.anthropic.com/docs/api/reference)
            topK: -1, // default value
            topP: -1, // default value
        })

        console.log(response)

        return response.completion.trim().split(' - ')[0]
    }
}
