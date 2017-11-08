import ErrorIcon from '@sourcegraph/icons/lib/Error'
import * as React from 'react'
import reactive from 'rx-component'
import 'rxjs/add/observable/merge'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/do'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/scan'
import 'rxjs/add/operator/withLatestFrom'
import { Observable } from 'rxjs/Observable'
import { Subject } from 'rxjs/Subject'
import { eventLogger } from '../tracking/eventLogger'
import { addCommentToThread } from './backend'

interface Props {
    onOpenEditor: () => void
    onThreadUpdated: (updatedThread: GQL.ISharedItemThread) => void
    threadID: number
    ulid: string
}

interface State {
    onOpenEditor: () => void
    textAreaValue: string
    submitting: boolean
    error?: any
}

type Update = (s: State) => State

export const CommentsInput = reactive<Props>(props => {
    const submits = new Subject<React.FormEvent<HTMLFormElement>>()
    const nextSubmit = (e: React.FormEvent<HTMLFormElement>) => submits.next(e)

    const textAreaChanges = new Subject<string>()
    const nextTextAreaChange = (e: React.ChangeEvent<HTMLTextAreaElement>) =>
        textAreaChanges.next(e.currentTarget.value)

    return Observable.merge(
        props.map(({ onOpenEditor }): Update => state => ({ ...state, onOpenEditor })),

        textAreaChanges.map((textAreaValue): Update => state => ({ ...state, textAreaValue })),

        // Prevent default and set submitting = true when submits occur.
        submits
            .do(e => e.preventDefault())
            .do(e => eventLogger.log('RepliedToThread'))
            .map((): Update => state => ({ ...state, submitting: true })),

        // Add comment to thread when submits occur.
        submits.withLatestFrom(textAreaChanges, props).mergeMap(([, textAreaValue, props]) =>
            addCommentToThread(props.threadID, textAreaValue, props.ulid)
                .do(updatedThread => props.onThreadUpdated(updatedThread))
                .map((updatedThread): Update => state => ({ ...state, submitting: false, textAreaValue: '' }))
                .catch((error): Update[] => {
                    console.error(error)
                    return [state => ({ ...state, error, submitting: false })]
                })
        )
    )
        .scan<Update, State>((state: State, update: Update) => update(state), {} as State)
        .map(({ onOpenEditor, textAreaValue, submitting, error }: State): JSX.Element | null => (
            <form className="comments-input" onSubmit={nextSubmit}>
                <textarea
                    className="ui-text-box comments-input__text-box"
                    placeholder="Leave a comment..."
                    autoFocus={true}
                    onChange={nextTextAreaChange}
                    value={textAreaValue}
                />
                <div className="comments-input__bottom-container">
                    {error && (
                        <span className="comments-input__error">
                            <ErrorIcon className="icon-inline comments-input__error-icon" />
                            {error.message}
                        </span>
                    )}
                    {!error && <span className="comments-input__markdown-supported">Markdown supported.</span>}
                    <button className="btn btn-primary comments-input__button" type="button" onClick={onOpenEditor}>
                        Open in Sourcegraph Editor
                    </button>
                    <button type="submit" className="btn btn-primary comments-input__button" disabled={submitting}>
                        Comment
                    </button>
                </div>
            </form>
        ))
})
