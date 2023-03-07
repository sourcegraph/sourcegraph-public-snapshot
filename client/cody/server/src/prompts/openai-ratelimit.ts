import { AxiosResponse } from 'axios'
import * as openai from 'openai'

let queue: Promise<any> = Promise.resolve()
let queueLength = 0

// This is a shim to enforce the constraint that API requests to openai happen in serial (otherwise we get HTTP 429s)
export function createCompletion(
	oa: openai.OpenAIApi,
	createCompletionRequest: openai.CreateCompletionRequest
): Promise<Pick<AxiosResponse<openai.CreateCompletionResponse, any>, 'data'>> {
	const ret = queue.then(async () => {
		const result = await oa.createCompletion(createCompletionRequest)
		queueLength--
		if (queueLength === 0) {
			console.log('refreshing openai queue')
			queue = Promise.resolve() // reset queue, so old promise chain can be gc'd
		}
		return result
	})
	queue = ret.finally()
	queueLength++
	return ret
}
