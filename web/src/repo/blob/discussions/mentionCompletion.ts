import { Position } from '@sourcegraph/extension-api-types'
import { from, Observable, of, Subscription, Unsubscribable } from 'rxjs'
import { first, map, switchMap } from 'rxjs/operators'
import { CompletionList } from 'sourcegraph'
import { COMMENT_URI_SCHEME, positionToOffset } from '../../../../../shared/src/api/client/types/textDocument'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { getWordAtText } from '../../../../../shared/src/util/wordHelpers'
import { fetchAllUsers } from '../../../site-admin/backend'

/**
 * Registers contributions for username mention completion in discussion comments.
 */
export function registerDiscussionsMentionCompletionContributions({
    extensionsController,
}:
    | ExtensionsControllerProps
    | {
          extensionsController: {
              services: {
                  completionItems: {
                      registerProvider: ExtensionsControllerProps['extensionsController']['services']['completionItems']['registerProvider']
                  }
                  model: {
                      models: ExtensionsControllerProps['extensionsController']['services']['model']['models']
                  }
              }
          }
      }): Unsubscribable {
    const subscriptions = new Subscription()
    subscriptions.add(
        extensionsController.services.completionItems.registerProvider(
            {
                documentSelector: [{ scheme: COMMENT_URI_SCHEME }],
            },
            params =>
                from(extensionsController.services.model.models).pipe(
                    switchMap(models => {
                        const model = models.find(m => m.uri === params.textDocument.uri)
                        if (!model) {
                            throw new Error(`model not found: ${params.textDocument.uri}`)
                        }
                        if (model.text === undefined) {
                            return of(null)
                        }
                        return provideMentionCompletions(model.text, params.position)
                    }),
                    first()
                )
        )
    )
    return subscriptions
}

/**
 * Provides username mention completions for the cursor position. This is usually not called
 * directly; it is registered in {@link registerDiscussionsMentionCompletionContributions} and
 * invoked via the completion item provider registry.
 *
 * @param queryUsernamesFunction For mocking in tests.
 */
export function provideMentionCompletions(
    text: string,
    position: Position,
    queryUsernamesFunction = queryUsernames
): Observable<CompletionList | null> {
    // Check the text that the user is currently typing to see if they have typed "@" (and aren't
    // typing an email address, i.e., the word begins with "@").
    const word = getWordAtText(positionToOffset(text, position), text)
    if (word && word.word.startsWith('@')) {
        return queryUsernamesFunction(word.word.slice(1)).pipe(
            map(usernames => ({ items: usernames.map(username => ({ label: username, insertText: `@${username} ` })) }))
        )
    }
    return of(null)
}

/**
 * Finds usernames matching the query.
 *
 * @param query A partial username.
 */
function queryUsernames(query: string): Observable<string[]> {
    return fetchAllUsers({ first: 100, query }).pipe(map(({ nodes }) => nodes.map(({ username }) => username)))
}
