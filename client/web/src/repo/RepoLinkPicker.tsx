import { type FC, useRef, useState } from 'react'

import {
    mdiAws,
    mdiBitbucket,
    mdiChevronDown,
    mdiGit,
    mdiGithub,
    mdiGitlab,
    mdiMicrosoftAzure,
    mdiSourceRepository,
} from '@mdi/js'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { useQuery, gql } from '@sourcegraph/http-client'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import {
    Button,
    Icon,
    Link,
    Popover,
    PopoverTrigger,
    PopoverContent,
    Combobox,
    ComboboxInput,
    ComboboxList,
    ComboboxOption,
    ComboboxOptionText,
    LoadingSpinner,
    ErrorAlert,
    useDebounce,
} from '@sourcegraph/wildcard'

import {
    ExternalServiceKind,
    type RepositoriesSuggestionsResult,
    type RepositoriesSuggestionsVariables,
} from '../graphql-operations'

import styles from './RepoLinkPicker.module.scss'

const REPOSITORIES_QUERY = gql`
    query RepositoriesSuggestions($query: String) {
        repositories(first: 15, query: $query) {
            nodes {
                id
                name
                url
                externalServices(first: 1) {
                    nodes {
                        id
                        kind
                    }
                }
            }
            pageInfo {
                hasNextPage
            }
        }
    }
`

interface RepoLinkPickerProps {
    repositoryURL: string
    repositoryName: string
    disabled?: boolean
    className?: string
}

export const RepoLinkPicker: FC<RepoLinkPickerProps> = props => {
    const { repositoryURL, repositoryName, disabled, className } = props

    const navigate = useNavigate()

    const rootRef = useRef<HTMLDivElement>(null)
    const [isSuggestionOpen, setSuggestionOpen] = useState<boolean>(false)
    const [searchTerm, setSearchTerm] = useState<string>('')
    const debouncedSearchTerm = useDebounce(searchTerm, 500)

    const {
        data: currentData,
        previousData,
        error,
        loading,
    } = useQuery<RepositoriesSuggestionsResult, RepositoriesSuggestionsVariables>(REPOSITORIES_QUERY, {
        skip: !isSuggestionOpen,
        variables: {
            query: searchTerm.length === 0 ? getInitialSearchTerm(repositoryName) : debouncedSearchTerm,
        },
        fetchPolicy: 'cache-first',
    })

    const handleSelect = (selectedValue: string): void => {
        navigate(`/${selectedValue}`)
        setSuggestionOpen(false)
        setSearchTerm('')
    }

    const data = currentData ?? previousData
    const suggestions = data?.repositories.nodes ?? []

    return (
        <div ref={rootRef} className={classNames(styles.root, className, { [styles.rootActive]: isSuggestionOpen })}>
            <Button
                as={Link}
                to={repositoryURL}
                disabled={disabled}
                size="sm"
                variant="secondary"
                outline={true}
                className={classNames('test-repo-header-repo-link', styles.linkButton)}
            >
                <Icon aria-hidden={true} svgPath={mdiSourceRepository} /> {displayRepoName(repositoryName)}
            </Button>
            <Popover isOpen={isSuggestionOpen} onOpenChange={state => setSuggestionOpen(state.isOpen)}>
                <PopoverTrigger
                    as={Button}
                    size="sm"
                    variant="secondary"
                    outline={true}
                    className={styles.dropdownButton}
                >
                    <Icon svgPath={mdiChevronDown} aria-label="Show repository picker" />
                </PopoverTrigger>
                <PopoverContent target={rootRef.current} className={styles.popover}>
                    <Combobox aria-label="Choose a repo" className={styles.combobox} onSelect={handleSelect}>
                        <ComboboxInput
                            value={searchTerm}
                            placeholder="Search repository..."
                            status={loading ? 'loading' : 'initial'}
                            autoFocus={true}
                            onChange={event => setSearchTerm(event.target.value)}
                        />

                        {!error && loading && suggestions.length === 0 && (
                            <div className={styles.loading}>
                                <LoadingSpinner />
                            </div>
                        )}
                        {error && <ErrorAlert error={error} className="mt-3 mb-0" />}

                        {!error && suggestions.length > 0 && (
                            <ComboboxList className={styles.list}>
                                {suggestions.map(suggestion => (
                                    <ComboboxOption
                                        key={suggestion.name}
                                        value={suggestion.name}
                                        className={styles.item}
                                    >
                                        <Icon
                                            svgPath={getCodeHostIconPath(suggestion.externalServices.nodes[0]?.kind)}
                                            height="1rem"
                                            width="1rem"
                                            inline={false}
                                            aria-hidden={true}
                                            className="mr-1"
                                        />
                                        <ComboboxOptionText />
                                    </ComboboxOption>
                                ))}
                            </ComboboxList>
                        )}
                    </Combobox>
                </PopoverContent>
            </Popover>
        </div>
    )
}

export const getCodeHostIconPath = (codeHostType?: ExternalServiceKind): string => {
    switch (codeHostType) {
        case ExternalServiceKind.GITHUB: {
            return mdiGithub
        }
        case ExternalServiceKind.BITBUCKETCLOUD: {
            return mdiBitbucket
        }
        case ExternalServiceKind.BITBUCKETSERVER: {
            return mdiBitbucket
        }
        case ExternalServiceKind.GITLAB: {
            return mdiGitlab
        }
        case ExternalServiceKind.GITOLITE: {
            return mdiGit
        }
        case ExternalServiceKind.AWSCODECOMMIT: {
            return mdiAws
        }
        case ExternalServiceKind.AZUREDEVOPS: {
            return mdiMicrosoftAzure
        }
        default: {
            return mdiSourceRepository
        }
    }
}

const getInitialSearchTerm = (repo: string) => {
    const r = repo.split('/')
    return r[r.length - 1]
}
