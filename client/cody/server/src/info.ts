import { QueryInfo } from '@sourcegraph/cody-common'

import { ClaudeBackend } from './prompts/claude'

export async function getInfo(backend: ClaudeBackend, query: string): Promise<QueryInfo> {
	const needsCodebaseContext = await new Promise<boolean>((resolve, reject) => {
		backend.chat(
			[
				{
					speaker: 'you',
					text: `The user has a code editor open to a current file inside the current codebase. Does the following question from the user require knowledge of other files in the current codebase (not the current file)? Answer ONLY with a single word, "yes" or "no".\n${query}`,
					// text: `Does the following question require knowledge about a specific codebase to answer? Answer ONLY with "yes" or "no".\n${query}`,
				},
			],
			{
				onChange: () => {},
				onComplete: message => {
					resolve(message.trim().split(' ')[0].toLocaleLowerCase() === 'yes')
				},
				onError: error => reject(error),
			}
		)
	})
	const needsCurrentFileContext = await new Promise<boolean>((resolve, reject) => {
		backend.chat(
			[
				{
					speaker: 'you',
					text: `The user has a code editor open to a current file inside the current codebase. Does the following question from the user require knowledge of the current file? Answer ONLY with a single word, "yes" or "no".\n${query}`,
				},
			],
			{
				onChange: () => {},
				onComplete: message => {
					resolve(message.trim().split(' ')[0].toLocaleLowerCase() === 'yes')
				},
				onError: error => reject(error),
			}
		)
	})
	return { needsCodebaseContext, needsCurrentFileContext }
}
