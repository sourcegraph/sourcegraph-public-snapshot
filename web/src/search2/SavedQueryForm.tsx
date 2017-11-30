import ErrorIcon from '@sourcegraph/icons/lib/Error'
import * as React from 'react'
import reactive from 'rx-component'
import { Observable } from 'rxjs/Observable'
import { fromEvent } from 'rxjs/observable/fromEvent'
import { merge } from 'rxjs/observable/merge'
import { catchError } from 'rxjs/operators/catchError'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { scan } from 'rxjs/operators/scan'
import { take } from 'rxjs/operators/take'
import { tap } from 'rxjs/operators/tap'
import { withLatestFrom } from 'rxjs/operators/withLatestFrom'
import { Subject } from 'rxjs/Subject'
import { configurationCascade } from '../settings/configuration'

export interface SavedQueryFields {
    description: string
    query: string
    scopeQuery: string
    subject: GQLID
}

interface Props {
    defaultValues?: Partial<SavedQueryFields>
    onDidCommit: () => void
    onDidCancel: () => void
    title?: string
    submitLabel: string
    cancelLabel: string
    onSubmit: (fields: SavedQueryFields) => Observable<void>
}

interface State extends SavedQueryFields {
    subjectOptions: GQL.ConfigurationSubject[]
    onDidCancel: () => void
    title?: string
    submitLabel: string
    cancelLabel: string
    submitting: boolean
    error?: any
}

type Update = (s: State) => State

export const SavedQueryForm = reactive<Props>(props => {
    let descriptionInput: HTMLInputElement | null = null
    let queryInput: HTMLInputElement | null = null
    let scopeQueryInput: HTMLInputElement | null = null
    let subjectInput: HTMLSelectElement | null = null

    const submits = new Subject<React.FormEvent<HTMLFormElement>>()
    const nextSubmit = (e: React.FormEvent<HTMLFormElement>) => submits.next(e)

    const inputChanges = new Subject<SavedQueryFields>()
    const nextInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
        inputChanges.next({
            description: descriptionInput ? descriptionInput.value : '',
            query: queryInput ? queryInput.value : '',
            scopeQuery: scopeQueryInput ? scopeQueryInput.value : '',
            subject: subjectInput ? subjectInput.value : '',
        })
    }

    return merge(
        // TODO(sqs): don't allow Escape while submitting
        fromEvent<KeyboardEvent>(window, 'keydown').pipe(
            filter((e: KeyboardEvent) => e.key === 'Escape'),
            withLatestFrom(props),
            tap(([, props]) => props.onDidCancel())
        ),

        // Apply default values.
        //
        // TODO(sqs): is there a better way to do this?
        props.pipe(
            take(1),
            map(({ defaultValues }) => defaultValues),
            filter<SavedQueryFields>(defaultValues => !!defaultValues),
            tap(defaultValues => {
                inputChanges.next({
                    description: defaultValues.description,
                    query: defaultValues.query,
                    scopeQuery: defaultValues.scopeQuery,
                    subject: defaultValues.subject,
                })
            }),
            map((defaultValues): Update => state => ({
                ...state,
                description: defaultValues.description,
                query: defaultValues.query,
                scopeQuery: defaultValues.scopeQuery,
                subject: defaultValues.subject,
            }))
        ),

        configurationCascade.pipe<{}>(
            map(({ subjects }): Update => state => ({
                ...state,
                subjectOptions: subjects,
            }))
        ),

        props.pipe(
            map(({ onDidCancel, title, submitLabel, cancelLabel }): Update => state => ({
                ...state,
                onDidCancel,
                title,
                submitLabel,
                cancelLabel,
            }))
        ),

        inputChanges.pipe(map((fields): Update => state => ({ ...state, ...fields }))),

        // Prevent default and set submitting = true when submits occur.
        submits.pipe(tap(e => e.preventDefault()), map((): Update => state => ({ ...state, submitting: true }))),

        // Handle submit.
        submits.pipe(
            withLatestFrom(inputChanges, props),
            mergeMap(([, fields, props]) =>
                props.onSubmit(fields).pipe(
                    tap(() => props.onDidCommit()),
                    map((): Update => state => ({
                        ...state,
                        submitting: false,
                        description: '',
                        query: '',
                        scopeQuery: '',
                        subject: '',
                    })),
                    catchError((error): Update[] => {
                        console.error(error)
                        return [state => ({ ...state, error, submitting: false })]
                    })
                )
            )
        )
    ).pipe(
        scan<Update, State>((state: State, update: Update) => update(state), {} as State),
        map(
            ({
                description,
                query,
                scopeQuery,
                subject,
                subjectOptions,
                onDidCancel,
                title,
                submitLabel,
                cancelLabel,
                submitting,
                error,
            }: State): JSX.Element | null => (
                <form className="saved-query-form" onSubmit={nextSubmit}>
                    {title && <h3 className="saved-query-form__title">{title}</h3>}
                    <input
                        type="text"
                        name="description"
                        className="ui-text-box"
                        placeholder="Description"
                        autoFocus={true}
                        onChange={nextInputChange}
                        value={description || ''}
                        required={true}
                        ref={e => (descriptionInput = e)}
                    />
                    <input
                        type="text"
                        name="query"
                        className="ui-text-box"
                        placeholder="Query"
                        onChange={nextInputChange}
                        value={query || ''}
                        required={true}
                        autoCorrect="off"
                        spellCheck={false}
                        autoCapitalize="off"
                        ref={e => (queryInput = e)}
                    />
                    <input
                        type="text"
                        name="scope-query"
                        className="ui-text-box"
                        placeholder="Scope query"
                        onChange={nextInputChange}
                        value={scopeQuery || ''}
                        autoCorrect="off"
                        spellCheck={false}
                        autoCapitalize="off"
                        ref={e => (scopeQueryInput = e)}
                    />
                    <select
                        name="subject"
                        className="ui-text-box"
                        onChange={nextInputChange}
                        value={subject}
                        ref={e => (subjectInput = e)}
                    >
                        {subjectOptions.map((subjectOption, i) => (
                            <option key={i} value={subjectOption.id}>
                                {configurationSubjectLabel(subjectOption)}
                            </option>
                        ))}
                    </select>
                    <div className="saved-query-form__actions">
                        <button
                            type="submit"
                            className="btn btn-primary saved-query-form__button"
                            disabled={submitting}
                        >
                            {submitLabel}
                        </button>
                        <button
                            type="reset"
                            className="btn btn-secondary saved-query-form__button saved-query-form__button-cancel"
                            disabled={submitting}
                            onClick={onDidCancel}
                        >
                            {cancelLabel}
                        </button>
                    </div>
                    {error &&
                        !submitting && (
                            <div className="saved-query-form__error">
                                <ErrorIcon className="icon-inline saved-query-form__error-icon" />
                                {error.message}
                            </div>
                        )}
                </form>
            )
        )
    )
})

function configurationSubjectLabel(s: GQL.ConfigurationSubject): string {
    switch (s.__typename) {
        case 'User':
            return `${s.username} (user settings)`
        case 'Org':
            return `${s.name} (org settings)`
        default:
            throw new Error('no configuration subject')
    }
}
