import { flatMap, from, map, of, tap } from 'rxjs'

import { Settings } from '../../app/useSettings'
import { TaskProvider } from '../../tasks/taskProvider'
import { GOOGLE_API_KEY } from './authProvider'

export const googleTasksTaskProvider: TaskProvider = {
    name: 'Google Tasks',
    tasks: ({ googleAuth }: Settings) => {
        if (!googleAuth || !window.gapi) {
            return of([])
        }

        return from(
            window.gapi.client.init({
                apiKey: GOOGLE_API_KEY,
                clientId: googleAuth.jwtPayload.aud,
                discoveryDocs: ['https://tasks.googleapis.com/$discovery/rest'],
                scope: 'https://www.googleapis.com/auth/tasks.readonly',
            })
        ).pipe(
            flatMap(() => window.gapi.client.tasks.tasklists.list({ oauth_token: 'asdf' })),
            map(x => x.result.items || []),
            tap(x => console.log('TL', x)),
            map(tasks => tasks.map(task => ({ text: task.title || '', url: task.selfLink || '' })))
        )
    },
}
