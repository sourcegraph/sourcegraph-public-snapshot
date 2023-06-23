// A PromptVersion that has a prompt
export interface PromptVersion {
    prompt: string
    provider: 'anthropic' | 'openai'
    temperature: number
    maxTokensToSample: number
}
