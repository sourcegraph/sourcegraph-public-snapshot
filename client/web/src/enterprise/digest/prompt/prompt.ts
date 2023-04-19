export interface CommitPromptInput {
    input: {
        heading: string
        description: string | null
        diff: string | null
    }
    granularity: 'Overview' | 'Detailed'
}

export const buildCommitPrompt = ({ input, granularity }: CommitPromptInput): string => `
${JSON.stringify(input)}

Generate a summary of this change in a readable, plaintext, bullet-point list.

Additional information to help you build your summary:
- The summary should have the following granularity: ${granularity}
`
