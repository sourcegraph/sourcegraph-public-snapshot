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
import { Container } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { ALLOW_NAVIGATION, AwayPrompt } from '../components/AwayPrompt'
import { fetchRepository } from '../repo/backend'
import { SearchContextProps } from '../search'

import { DeleteSearchContextModal } from './DeleteSearchContextModal'
import { parseConfig } from './repositoryRevisionsConfigParser'
import styles from './SearchContextForm.module.scss'
import {
    getSelectedNamespace,
    getSelectedNamespaceFromUser,
    SearchContextOwnerDropdown,
    SelectedNamespace,
    SelectedNamespaceType,
} from './SearchContextOwnerDropdown'
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
        selectedNamespaceType === 'global-owner'
            ? 'Only site-admins can view this context.'
            : selectedNamespaceType === 'org'
            ? 'Only organization members can view this context.'
            : 'Only you can view this context.'

    return [
        {
            visibility: 'public',
            title: 'Public',
            description:
                'Anyone can view this context. Public repositories will be visible to everyone. ' +
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
        <code className={styles.searchContextFormPreview} data-testid="search-context-preview">
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

export interface SearchContextFormProps
    extends RouteComponentProps,
        ThemeProps,
        TelemetryProps,
        Pick<SearchContextProps, 'deleteSearchContext'> {
    searchContext?: ISearchContext
    authenticatedUser: AuthenticatedUser

    onSubmit: (
        id: Scalars['ID'] | undefined,
        searchContext: SearchContextInput,
        repositories: SearchContextRepositoryRevisionsInput[]
    ) => Observable<ISearchContext>
}

const searchContextVisibility = (searchContext: ISearchContext): SelectedVisibility =>
    searchContext.public ? 'public' : 'private'

export const SearchContextForm: React.FunctionComponent<SearchContextFormProps> = props => {
    const { authenticatedUser, onSubmit, searchContext, deleteSearchContext } = props
    const history = useHistory()

    const [name, setName] = useState(searchContext ? searchContext.name : '')
    const [description, setDescription] = useState(searchContext ? searchContext.description : '')
    const [visibility, setVisibility] = useState<SelectedVisibility>(
        searchContext ? searchContextVisibility(searchContext) : 'public'
    )

    const isValidName = useMemo(() => name.length === 0 || name.match(VALIDATE_NAME_REGEXP) !== null, [name])

    const [selectedNamespace, setSelectedNamespace] = useState<SelectedNamespace>(
        searchContext ? getSelectedNamespace(searchContext.namespace) : getSelectedNamespaceFromUser(authenticatedUser)
    )

    const visibilityRadioButtons = useMemo(() => getVisibilityRadioButtons(selectedNamespace.type), [selectedNamespace])

    const searchContextSpecPreview = isValidName ? (
        getSearchContextSpecPreview(selectedNamespace, name)
    ) : (
        <div className="text-danger">Invalid context name</div>
    )

    const [hasRepositoriesConfigChanged, setHasRepositoriesConfigChanged] = useState(false)
    const [repositoriesConfig, setRepositoriesConfig] = useState('')
    const onRepositoriesConfigChange = useCallback(
        (config, isInitialValue) => {
            setRepositoriesConfig(config)
            if (!isInitialValue && config !== repositoriesConfig) {
                setHasRepositoriesConfigChanged(true)
            }
        },
        [repositoriesConfig, setRepositoriesConfig, setHasRepositoriesConfigChanged]
    )

    const hasChanges = useMemo(() => {
        if (!searchContext) {
            return (
                name.length > 0 ||
                description.length > 0 ||
                visibility !== 'public' ||
                selectedNamespace.type !== 'user' ||
                hasRepositoriesConfigChanged
            )
        }
        return (
            searchContext.name !== name ||
            searchContext.description !== description ||
            searchContextVisibility(searchContext) !== visibility ||
            hasRepositoriesConfigChanged
        )
    }, [description, name, searchContext, selectedNamespace, visibility, hasRepositoriesConfigChanged])

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
                                    history.push('/contexts?order=updated-at-desc', ALLOW_NAVIGATION)
                                }
                            })
                        )
                    ),
                    catchError(error => [asError(error)])
                ),
            [onSubmit, parseRepositories, name, description, visibility, selectedNamespace, history, searchContext]
        )
    )

    const onCancel = useCallback(() => {
        history.push('/contexts')
    }, [history])

    const [showDeleteModal, setShowDeleteModal] = useState(false)
    const toggleDeleteModal = useCallback(() => setShowDeleteModal(show => !show), [setShowDeleteModal])

    return (
        <Form onSubmit={submitRequest}>
            <Container className="mb-3">
                <div className="d-flex">
                    <div className="mr-2">
                        <div className="mb-2">Owner</div>
                        <SearchContextOwnerDropdown
                            isDisabled={!!searchContext}
                            selectedNamespace={selectedNamespace}
                            setSelectedNamespace={setSelectedNamespace}
                            authenticatedUser={authenticatedUser}
                        />
                    </div>
                    <div className="flex-1">
                        <div className="mb-2">Context name</div>
                        <input
                            className={classNames('w-100', 'form-control', styles.searchContextFormNameInput)}
                            data-testid="search-context-name-input"
                            value={name}
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
                    <small>
                        The best context names are short and semantic. {MAX_NAME_LENGTH} characters max. Alphanumeric
                        and <kbd>.</kbd>
                        <kbd>_</kbd>
                        <kbd>/</kbd>
                        <kbd>-</kbd> characters only.
                    </small>
                </div>
                <div>
                    <div className={classNames('mb-1', styles.searchContextFormPreviewTitle)}>Preview</div>
                    {searchContextSpecPreview}
                </div>
                <hr className={classNames('my-4', styles.searchContextFormDivider)} />
                <div>
                    <div className="mb-2">
                        Description <span className="text-muted">(optional)</span>
                    </div>
                    <textarea
                        className="form-control w-100"
                        data-testid="search-context-description-input"
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
                    <div className="mt-1 text-muted">
                        <small>
                            <span>Markdown formatting is supported</span>
                            <span className="px-1">&middot;</span>
                            <span>{MAX_DESCRIPTION_LENGTH - description.length} characters remaining</span>
                        </small>
                    </div>
                </div>
                <div className="mt-3">
                    <div className="mb-3">Visibility</div>
                    {visibilityRadioButtons.map(radio => (
                        <label key={radio.visibility} className="d-flex mt-2">
                            <div className="mr-2">
                                <input
                                    className={styles.searchContextFormVisibilityRadio}
                                    name="visibility"
                                    type="radio"
                                    value={radio.visibility}
                                    checked={visibility === radio.visibility}
                                    required={true}
                                    onChange={() => setVisibility(radio.visibility)}
                                />
                            </div>
                            <div>
                                <strong className={styles.searchContextFormVisibilityTitle}>{radio.title}</strong>
                                <div className="text-muted">
                                    <small>{radio.description}</small>
                                </div>
                            </div>
                        </label>
                    ))}
                </div>
                <hr className={classNames('my-4', styles.searchContextFormDivider)} />
                <div>
                    <div className="mb-1">Repositories and revisions</div>
                    <div className="text-muted mb-3">
                        Define which repositories and revisions should be included in this search context.
                    </div>
                    <SearchContextRepositoriesFormArea
                        {...props}
                        onChange={onRepositoriesConfigChange}
                        validateRepositories={validateRepositories}
                        repositories={searchContext?.repositories}
                    />
                </div>
            </Container>
            <div className="d-flex">
                <button
                    type="submit"
                    className="btn btn-primary mr-2 test-search-context-submit-button"
                    data-testid="search-context-submit-button"
                    disabled={searchContextOrError && searchContextOrError === LOADING}
                >
                    {!searchContext ? 'Create search context' : 'Save'}
                </button>
                <button type="button" onClick={onCancel} className="btn btn-outline-secondary">
                    Cancel
                </button>
                {searchContext && (
                    <>
                        <div className="flex-grow-1" />
                        <button
                            type="button"
                            data-testid="search-context-delete-button"
                            className="btn btn-outline-secondary text-danger"
                            onClick={toggleDeleteModal}
                        >
                            Delete
                        </button>
                        <DeleteSearchContextModal
                            isOpen={showDeleteModal}
                            deleteSearchContext={deleteSearchContext}
                            searchContext={searchContext}
                            toggleDeleteModal={toggleDeleteModal}
                        />
                    </>
                )}
            </div>
            {isErrorLike(searchContextOrError) && (
                <div className="alert alert-danger mt-2">
                    Failed to create search context: {searchContextOrError.message}
                </div>
            )}
            <AwayPrompt
                header="Discard unsaved changes?"
                message="All unsaved changes will be lost."
                button_ok_text="Discard"
                when={hasChanges}
            />
        </Form>
    )
}
