import { fetchEventSource } from '@microsoft/fetch-event-source'
import { Observable } from 'rxjs'

export interface CompleteParams {
    prompt: string
    numCompletions?: number
    maxTokens: string
    sourcegraphURL?: string
}

export interface CompleteResult {
    choices: CompletionChoice[]
}

export interface CompletionChoice {
    text: string
}

export function complete({
    prompt,
    numCompletions = 1,
    maxTokens,
    sourcegraphURL = 'https://sourcegraph.test:3443',
}: CompleteParams): Observable<CompletionChoice[]> {
    const allChoices: CompletionChoice[] = []
    return new Observable<CompletionChoice[]>(observer => {
        fetchEventSource(`${sourcegraphURL}/.api/mlx/complete`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                prompt,
                n: numCompletions,
                max_tokens: maxTokens,
            }),
            onmessage(event) {
                allChoices.push(JSON.parse(event.data))
                observer.next(allChoices)
            },
            onerror(event) {
                observer.error(event)
            },
        }).then(
            () => observer.complete(),
            error => observer.error(error)
        )
    })
}
