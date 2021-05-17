import classNames from 'classnames'
import React, { useCallback, useMemo, useState } from 'react'
import { RouteComponentProps, useHistory } from 'react-router'
import { Form } from 'reactstrap'
import { from, Observable, of, throwError } from 'rxjs'
import { catchError, concatMap, filter, map, startWith, switchMap, tap, toArray } from 'rxjs/operators'

import {
    Scalars,
    SearchContextInput,
    SearchContextRepositoryRevisionsInput,
} from '@sourcegraph/shared/src/graphql-operations'
import { ISearchContext, ISearchContextRepositoryRevisionsInput } from '@sourcegraph/shared/src/graphql/schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { asError, createAggregateError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useEventObservable } from '@sourcegraph/shared/src/util/useObservable'

import { AuthenticatedUser } from '../auth'
import { fetchRepository } from '../repo/backend'

import { parseConfig } from './repositoryRevisionsConfigParser'
import styles from './SearchContextForm.module.scss'
import { SearchContextOwnerDropdown, SelectedNamespace, SelectedNamespaceType } from './SearchContextOwnerDropdown'
import { SearchContextRepositoriesFormArea } from './SearchContextRepositoriesFormArea'

const MAX_DESCRIPTION_LENGTH = 1024
const MAX_NAME_LENGTH = 32
const VALIDATE_NAME_REGEXP = /^[\w./-]+$/

type SelectedVisibility = 'public' | 'private'

interface VisibilityRadioButton {
    visibility: SelectedVisibility
    title: string
    description: string
}

function getVisibilityRadioButtons(selectedNamespaceType: SelectedNamespaceType): VisibilityRadioButton[] {
    const privateVisibilityDescription =
        selectedNamespaceType === 'no-owner'
            ? 'Only site-admins can view this context.'
            : selectedNamespaceType === 'org'
            ? 'Only organization members can view this context.'
            : 'Only you can view this context.'

    return [
        {
            visibility: 'public',
            title: 'Public',
            description:
                'Anyone on Sourcegraph can view this context. Public repositories will be visible to all users. ' +
                'Private repositories will only be visible to users that have permission to view the repository via the code host.',
        },
        {
            visibility: 'private',
            title: 'Private',
            description: privateVisibilityDescription,
        },
    ]
}

function getSearchContextSpecPreview(selectedNamespace: SelectedNamespace, searchContextName: string): JSX.Element {
    return (
        <code className={classNames('test-search-context-preview', styles.searchContextFormPreview)}>
            <span className="search-filter-keyword">context:</span>
            {selectedNamespace.name.length > 0 && (
                <>
                    <span className="search-keyword">@</span>
                    <span>{selectedNamespace.name}/</span>
                </>
            )}
            <span>{searchContextName}</span>
        </code>
    )
}

const LOADING = 'loading' as const

export interface SearchContextFormProps extends RouteComponentProps, ThemeProps, TelemetryProps {
    searchContext?: ISearchContext
    authenticatedUser: AuthenticatedUser

    onSubmit: (
        id: Scalars['ID'] | undefined,
        searchContext: SearchContextInput,
        repositories: SearchContextRepositoryRevisionsInput[]
    ) => Observable<ISearchContext>
}

export const SearchContextForm: React.FunctionComponent<SearchContextFormProps> = props => {
    const { authenticatedUser, onSubmit, searchContext } = props
    const history = useHistory()

    const [name, setName] = useState('')
    const [description, setDescription] = useState('')
    const [visibility, setVisibility] = useState<SelectedVisibility>('public')

    const isValidName = useMemo(() => name.length === 0 || name.match(VALIDATE_NAME_REGEXP) !== null, [name])

    const selectedUserNamespace = {
        id: authenticatedUser.id,
        type: 'user' as SelectedNamespaceType,
        name: authenticatedUser.username,
    }
    const [selectedNamespace, setSelectedNamespace] = useState<SelectedNamespace>(selectedUserNamespace)

    const visibilityRadioButtons = useMemo(() => getVisibilityRadioButtons(selectedNamespace.type), [selectedNamespace])

    const searchContextSpecPreview = isValidName ? (
        getSearchContextSpecPreview(selectedNamespace, name)
    ) : (
        <div className="text-danger">Invalid context name</div>
    )

    const [repositoriesConfig, setRepositoriesConfig] = useState('')
    const onRepositoriesConfigChange = useCallback(
        config => {
            setRepositoriesConfig(config)
        },
        [setRepositoriesConfig]
    )

    const hasChanges = useMemo(() => {
        if (!searchContext) {
            return (
                name.length > 0 ||
                description.length > 0 ||
                visibility !== 'public' ||
                selectedNamespace.type !== 'user'
            )
        }
        // TODO: Check for changes when editing context
        return true
    }, [description, name, searchContext, selectedNamespace, visibility])

    const onCancel = useCallback(() => {
        if (hasChanges) {
            if (window.confirm('Leave page? All unsaved changes will be lost.')) {
                history.push('/contexts')
            }
        } else {
            history.push('/contexts')
        }
    }, [hasChanges, history])

    const parseRepositories = useCallback(
        () =>
            of(parseConfig(repositoriesConfig)).pipe(
                switchMap(config => {
                    if (config === null) {
                        return of([
                            new Error('Invalid configuration format. Check for inline editor warnings and errors.'),
                        ])
                    }
                    return from(config).pipe(
                        concatMap(({ repository: repoName, revisions }) =>
                            fetchRepository({ repoName }).pipe(
                                map(repository => ({ repositoryID: repository.id, revisions })),
                                catchError(error => [asError(error)])
                            )
                        ),
                        toArray()
                    )
                })
            ),
        [repositoriesConfig]
    )

    const validateRepositories = useCallback(
        () =>
            parseRepositories().pipe(
                switchMap(items => items),
                filter(repoOrError => isErrorLike(repoOrError)),
                map(repoOrError => repoOrError as Error),
                toArray()
            ),
        [parseRepositories]
    )

    const [submitRequest, searchContextOrError] = useEventObservable(
        useCallback(
            (submit: Observable<React.FormEvent<HTMLFormElement>>) =>
                submit.pipe(
                    tap(event => event.preventDefault()),
                    switchMap(parseRepositories),
                    switchMap(repositories => {
                        const validationErrors = repositories.filter(repository => isErrorLike(repository)) as Error[]
                        if (validationErrors.length > 0) {
                            return throwError(createAggregateError(validationErrors))
                        }
                        const validRepositories = repositories.filter(
                            repository => !isErrorLike(repository)
                        ) as ISearchContextRepositoryRevisionsInput[]
                        return of(validRepositories)
                    }),
                    switchMap(repositoryRevisionsArray =>
                        onSubmit(
                            searchContext?.id,
                            { name, description, public: visibility === 'public', namespace: selectedNamespace.id },
                            repositoryRevisionsArray
                        ).pipe(
                            startWith(LOADING),
                            catchError(error => [asError(error)]),
                            tap(successOrError => {
                                if (!isErrorLike(successOrError) && successOrError !== LOADING) {
                                    history.push('/contexts?order=updated-at-desc')
                                }
                            })
                        )
                    ),
                    catchError(error => [asError(error)])
                ),
            [onSubmit, parseRepositories, name, description, visibility, selectedNamespace, history, searchContext]
        )
    )

    return (
        <Form onSubmit={submitRequest}>
            <div className="d-flex">
                <div className="mr-2">
                    <div className="mb-2">Owner</div>
                    <SearchContextOwnerDropdown
                        selectedNamespace={selectedNamespace}
                        setSelectedNamespace={setSelectedNamespace}
                        selectedUserNamespace={selectedUserNamespace}
                        authenticatedUser={authenticatedUser}
                    />
                </div>
                <div className="flex-1">
                    <div className="mb-2">Context name</div>
                    <input
                        className={classNames(
                            'w-100 form-control test-search-context-name-input',
                            styles.searchContextFormNameInput
                        )}
                        type="text"
                        pattern="^[a-zA-Z0-9_\-\/\.]+$"
                        required={true}
                        maxLength={MAX_NAME_LENGTH}
                        onChange={event => {
                            setName(event.target.value)
                        }}
                    />
                </div>
            </div>
            <div className="text-muted my-2">
                The best context names are short and semantic. Context names are limited to {MAX_NAME_LENGTH}{' '}
                characters. They can contain alphanumeric and following non-alphanumeric characters: <kbd>.</kbd>
                <kbd>_</kbd>
                <kbd>/</kbd>
                <kbd>-</kbd> (no spaces).
            </div>
            <div>
                <div className={classNames('mb-1', styles.searchContextFormPreviewTitle)}>Preview</div>
                {searchContextSpecPreview}
            </div>
            <hr className="my-4" />
            <div>
                <div className="mb-2">
                    Description <span className="text-muted">(optional)</span>
                </div>
                <textarea
                    className="form-control w-100 test-search-context-description-input"
                    maxLength={MAX_DESCRIPTION_LENGTH}
                    value={description}
                    rows={5}
                    onChange={event => {
                        const value = event.target.value
                        if (value.length <= MAX_DESCRIPTION_LENGTH) {
                            setDescription(event.target.value)
                        }
                    }}
                />
                <div className="mt-2 text-muted">
                    <span>Markdown formatting is supported</span>
                    <span className="px-1">&middot;</span>
                    <span>{MAX_DESCRIPTION_LENGTH - description.length} characters remaining</span>
                </div>
            </div>
            <hr className="my-4" />
            <div>
                <div>Visibility</div>
                {visibilityRadioButtons.map(radio => (
                    <label key={radio.visibility} className="d-flex mt-2">
                        <input
                            className={styles.searchContextFormVisibilityRadio}
                            name="visibility"
                            type="radio"
                            value={radio.visibility}
                            checked={visibility === radio.visibility}
                            required={true}
                            onChange={() => setVisibility(radio.visibility)}
                        />
                        <div className="ml-2">
                            <strong className={styles.searchContextFormVisibilityTitle}>{radio.title}</strong>
                            <div className="text-muted">{radio.description}</div>
                        </div>
                    </label>
                ))}
            </div>
            <hr className="my-4" />
            <div>
                <div className="mb-1">Repositories and revisions</div>
                <div className="text-muted mb-2">
                    Define which repositories and revisions should be included in this search context.
                </div>
                <SearchContextRepositoriesFormArea
                    {...props}
                    onChange={onRepositoriesConfigChange}
                    validateRepositories={validateRepositories}
                    repositories={searchContext?.repositories}
                />
            </div>
            <hr className="my-4" />
            <div>
                <button
                    type="submit"
                    className="btn btn-primary mr-2 test-create-search-context-button"
                    disabled={searchContextOrError && searchContextOrError === LOADING}
                >
                    Create search context
                </button>
                <button type="button" onClick={onCancel} className="btn btn-outline-secondary">
                    Cancel
                </button>
            </div>
            {isErrorLike(searchContextOrError) && (
                <div className="alert alert-danger mt-2">
                    Failed to create search context: {searchContextOrError.message}
                </div>
            )}
        </Form>
    )
}
