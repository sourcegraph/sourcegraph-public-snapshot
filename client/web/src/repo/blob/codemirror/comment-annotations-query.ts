import { CompletionRequest, getCodyCompletionOneShot } from '../../../enterprise/cody/api'

export async function translateToQuery(input: string): Promise<string | null> {
    const messages: CompletionRequest['messages'] = [
        {
            speaker: 'human',
            text:
                'You are going to generate some comments for a file. ' +
                'Each comment should be useful and informative and only consist of a single line of text. ' +
                'You should not generate a comment if the code is self-explanatory. ' +
                'You should not generate a comment if the code is already commented. ' +
                'You should generate a comment if the code is not self-explanatory and is not already commented. ' +
                'You should output your comments as a JavaScript object, where the keys are the line numbers and the values are the comments.',
        },
        { speaker: 'assistant', text: 'Understood. I will follow these rules.' },
        { speaker: 'human', text: input },
    ]

    const result = await getCodyCompletionOneShot(messages)

    console.log(result)

    return ''
}
