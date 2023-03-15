import assert from 'assert'

import { Transcript } from '../../chat/prompt'

import { defaultEditor, defaultEmbeddingsClient, defaultIntentDetector, defaultKeywordContextFetcher } from './mocks'

describe('Prompt', () => {
    it('generates empty prompt with no messages', () => {
        const transcript = new Transcript(
            'none',
            defaultEmbeddingsClient,
            defaultIntentDetector,
            defaultKeywordContextFetcher,
            defaultEditor
        )

        const prompt = transcript.getPrompt()
        assert.deepStrictEqual(prompt, [])
    })
})
