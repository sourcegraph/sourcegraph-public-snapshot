import ErrorIcon from '@sourcegraph/icons/lib/Error'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { fromEvent } from 'rxjs/observable/fromEvent'
import { catchError } from 'rxjs/operators/catchError'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { Subscription } from 'rxjs/Subscription'
import { configurationCascade } from '../settings/configuration'

export interface SavedQueryFields {
    description: string
    query: string
    subject: GQLID
    showOnHomepage: boolean
}

interface Props {
    defaultValues?: Partial<SavedQueryFields>
    title?: string
    submitLabel: string
    onSubmit: (fields: SavedQueryFields) => Observable<void>
    onDidCommit: () => void
    onDidCancel: () => void
}

interface State {
    values: SavedQueryFields

    subjectOptions: GQL.ConfigurationSubject[]
    isSubmitting: boolean
    isFocused: boolean
    error?: any
}

export class SavedQueryForm extends React.Component<Props, State> {
    private handleQueryChange = this.createInputChangeHandler('query')
    private handleDescriptionChange = this.createInputChangeHandler('description')
    private handleSubjectChange = this.createInputChangeHandler('subject')
    private handleShowOnHomeChange = this.createInputChangeHandler('showOnHomepage')

    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        const { defaultValues } = props

        this.state = {
            values: {
                query: (defaultValues && defaultValues.query) || '',
                description: (defaultValues && defaultValues.description) || '',
                subject: (defaultValues && defaultValues.subject) || '',
                showOnHomepage: !!(defaultValues && defaultValues.showOnHomepage),
            },
            subjectOptions: [],
            isSubmitting: false,
            isFocused: false,
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            configurationCascade
                .pipe(map(({ subjects }) => subjects), filter(subjects => !!subjects))
                .subscribe(subjects => {
                    const subject = subjects.find(s => !!s.id)

                    this.setState(state => ({
                        subjectOptions: subjects,
                        values: {
                            ...state.values,
                            subject: state.values.subject || (subject && subject.id) || '',
                        },
                    }))
                })
        )

        this.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                .pipe(filter(event => !this.state.isFocused && event.key === 'Escape' && !this.state.isSubmitting))
                .subscribe(() => this.props.onDidCancel())
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element {
        const { onDidCancel, title, submitLabel } = this.props
        const {
            values: { query, description, subject, showOnHomepage },
            subjectOptions,
            isSubmitting,
            error,
        } = this.state
        console.log(this.state.isFocused)

        return (
            <form className="saved-query-form" onSubmit={this.handleSubmit}>
                {title && <h3 className="saved-query-form__title">{title}</h3>}
                <span>Search query</span>
                <input
                    type="text"
                    name="query"
                    className="form-control"
                    placeholder="Query"
                    onChange={this.handleQueryChange}
                    value={query || ''}
                    autoCorrect="off"
                    spellCheck={false}
                    autoCapitalize="off"
                    autoFocus={!query}
                    onFocus={this.handleInputFocus}
                    onBlur={this.handleInputBlur}
                />
                <span>Description</span>
                <input
                    type="text"
                    name="description"
                    className="form-control"
                    placeholder="Description"
                    onChange={this.handleDescriptionChange}
                    value={description || ''}
                    required={true}
                    autoFocus={!!query && !description}
                    onFocus={this.handleInputFocus}
                    onBlur={this.handleInputBlur}
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
                                        onChange={this.handleSubjectChange}
                                        type="radio"
                                        value={subjectOption.id}
                                        checked={subject === subjectOption.id}
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
                                onChange={this.handleShowOnHomeChange}
                            />
                            Show on homepage
                        </label>
                    </span>
                </div>
                <div className="saved-query-form__actions">
                    <button
                        type="submit"
                        className="btn btn-primary saved-query-form__button"
                        disabled={!(description && query && subject) || isSubmitting}
                    >
                        {submitLabel}
                    </button>
                    <button
                        type="reset"
                        className="btn btn-secondary saved-query-form__button saved-query-form__button-cancel"
                        disabled={isSubmitting}
                        onClick={onDidCancel}
                    >
                        Cancel
                    </button>
                </div>
                {error &&
                    !isSubmitting && (
                        <div className="saved-query-form__error">
                            <ErrorIcon className="icon-inline saved-query-form__error-icon" />
                            {error.message}
                        </div>
                    )}
            </form>
        )
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
        event.preventDefault()

        this.setState({ isSubmitting: true })

        this.subscriptions.add(
            this.props
                .onSubmit(this.state.values)
                .pipe(
                    catchError(error => {
                        console.error(error)
                        this.setState({ error })
                        if (!(error instanceof Error)) {
                            return [new Error(error)]
                        }

                        return [error]
                    }),
                    filter(v => !(v instanceof Error))
                )
                .subscribe(this.props.onDidCommit)
        )
    }

    private handleInputFocus = (event: React.FocusEvent<HTMLInputElement>) => {
        this.setState(() => ({ isFocused: true }))
    }

    private handleInputBlur = (event: React.FocusEvent<HTMLInputElement>) => {
        this.setState(() => ({ isFocused: false }))
    }

    /**
     * Returns an input change handler that updates the SavedQueryFields in the component's state
     *
     * @param key The key of saved query fields that a change of this input should update
     */
    private createInputChangeHandler(key: keyof SavedQueryFields): React.FormEventHandler<HTMLInputElement> {
        return event => {
            const { currentTarget: { value, type } } = event

            this.setState(state => ({
                values: {
                    ...state.values,
                    [key]: type === 'checkbox' ? Boolean(value) : String(value),
                },
            }))
        }
    }
}

function configurationSubjectLabel(s: GQL.IUser | GQL.IOrg): string {
    switch (s.__typename) {
        case 'User':
            return `${s.username} (user settings)`
        case 'Org':
            return `${s.name} (org settings)`
    }
}
