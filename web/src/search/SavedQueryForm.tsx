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
    subject: GQLID
    showOnHomepage: boolean
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
    canSubmit: boolean
    error?: any
}

type Update = (s: State) => State

export const SavedQueryForm = reactive<Props>(props => {
    let descriptionInput: HTMLInputElement | null = null
    let queryInput: HTMLInputElement | null = null
    let viewLocationInput: HTMLInputElement | null = null
    let subjectValue: string

    const submits = new Subject<React.FormEvent<HTMLFormElement>>()
    const nextSubmit = (e: React.FormEvent<HTMLFormElement>) => submits.next(e)

    const inputChanges = new Subject<SavedQueryFields>()
    const nextInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.currentTarget.type === 'radio') {
            subjectValue = e.target.value
        }

        inputChanges.next({
            description: descriptionInput ? descriptionInput.value : '',
            query: queryInput ? queryInput.value : '',
            subject: subjectValue,
            showOnHomepage: viewLocationInput ? viewLocationInput.checked : false,
        })
    }

    const isChecked = (currentSubject: string, subjectOptionID: string): boolean => {
        const isChecked = currentSubject === subjectOptionID
        // If the user has not selected a subject value, automatically check the first subject value.
        if (!subjectValue || subjectValue === subjectOptionID) {
            subjectValue = subjectOptionID
            return true
        }

        if (isChecked) {
            subjectValue = subjectOptionID
        }
        return isChecked
    }

    return merge(
        // TODO(sqs): don't allow Escape while submitting
        fromEvent<KeyboardEvent>(window, 'keydown').pipe(
            filter((e: KeyboardEvent) => e.key === 'Escape'),
            withLatestFrom(props),
            tap(([, props]) => props.onDidCancel()),
            // Don't produce any state updates
            mergeMap(() => [])
        ),

        // Apply default values.
        //
        // TODO(sqs): is there a better way to do this?
        props.pipe(
            take(1),
            map(({ defaultValues }) => defaultValues),
            filter(
                (defaultValues?: Partial<SavedQueryFields>): defaultValues is Partial<SavedQueryFields> =>
                    !!defaultValues
            ),
            map(defaultValues => ({
                description: defaultValues.description || '',
                query: defaultValues.query || '',
                subject: defaultValues.subject || '',
                showOnHomepage: defaultValues.showOnHomepage || false,
            })),
            tap(concreteValues => {
                inputChanges.next(concreteValues)
            }),
            map((defaultValues): Update => state => ({
                ...state,
                description: defaultValues.description,
                query: defaultValues.query,
                subject: defaultValues.subject,
                showOnHomepage: defaultValues.showOnHomepage,
            }))
        ),

        configurationCascade.pipe(
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
                subject,
                subjectOptions,
                showOnHomepage,
                onDidCancel,
                title,
                submitLabel,
                cancelLabel,
                submitting,
                error,
            }: State): JSX.Element | null => (
                <form className="saved-query-form" onSubmit={nextSubmit}>
                    {title && <h3 className="saved-query-form__title">{title}</h3>}
                    <span>Search query</span>
                    <input
                        type="text"
                        name="query"
                        className="form-control"
                        placeholder="Query"
                        onChange={nextInputChange}
                        value={query || ''}
                        autoCorrect="off"
                        spellCheck={false}
                        autoCapitalize="off"
                        ref={e => (queryInput = e)}
                        autoFocus={!query}
                    />
                    <span>Description</span>
                    <input
                        type="text"
                        name="description"
                        className="form-control"
                        placeholder="Description"
                        onChange={nextInputChange}
                        value={description || ''}
                        required={true}
                        ref={e => (descriptionInput = e)}
                        autoFocus={!!query}
                    />
                    <span>Save location</span>
                    <div className="saved-query-form__save-location">
                        {subjectOptions
                            .filter(
                                (subjectOption: GQL.ConfigurationSubject): subjectOption is GQL.IOrg | GQL.IUser =>
                                    subjectOption.__typename === 'Org' || subjectOption.__typename === 'User'
                            )
                            .map((subjectOption, i) => (
                                <span className="saved-query-form__save-location-options" key={i}>
                                    <label>
                                        <input
                                            className="saved-query-form__save-location-input"
                                            onChange={nextInputChange}
                                            type="radio"
                                            value={subjectOption.id}
                                            checked={isChecked(subject, subjectOption.id)}
                                        />
                                        {configurationSubjectLabel(subjectOption)}
                                    </label>
                                </span>
                            ))}
                    </div>
                    <div className="saved-query-form__save-location">
                        <span className="saved-query-form__save-location-options">
                            <label>
                                <input
                                    className="saved-query-form__save-location-input"
                                    type="checkbox"
                                    defaultChecked={showOnHomepage}
                                    onChange={nextInputChange}
                                    ref={e => (viewLocationInput = e)}
                                />
                                Show on homepage
                            </label>
                        </span>
                    </div>
                    <div className="saved-query-form__actions">
                        <button
                            type="submit"
                            className="btn btn-primary saved-query-form__button"
                            disabled={!(description && query && subject) || submitting}
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

function configurationSubjectLabel(s: GQL.IUser | GQL.IOrg): string {
    switch (s.__typename) {
        case 'User':
            return `${s.username} (user settings)`
        case 'Org':
            return `${s.name} (org settings)`
    }
}
