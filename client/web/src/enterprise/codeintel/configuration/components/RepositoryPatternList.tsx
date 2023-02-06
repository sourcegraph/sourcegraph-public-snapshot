import { FunctionComponent, useEffect, useMemo, useState } from 'react'

import { mdiDelete, mdiPlus } from '@mdi/js'
import classNames from 'classnames'
import { debounce } from 'lodash'

import {
    Alert,
    Button,
    Code,
    ErrorAlert,
    Icon,
    Input,
    InputErrorMessage,
    InputStatus,
    Link,
    LoadingSpinner,
    Text,
    Tooltip,
} from '@sourcegraph/wildcard'

import { ExternalRepositoryIcon } from '../../../../site-admin/components/ExternalRepositoryIcon'
import { usePreviewRepositoryFilter } from '../hooks/usePreviewRepositoryFilter'

import styles from './RepositoryPatternList.module.scss'

const DEBOUNCED_WAIT = 250

const DEFAULT_FETCH_LIMIT = 15

interface RepositoryPatternListProps {
    repositoryPatterns: string[]
    setRepositoryPatterns: (updater: (patterns: string[]) => string[]) => void
    disabled: boolean
}

export const RepositoryPatternList: FunctionComponent<RepositoryPatternListProps> = ({
    repositoryPatterns,
    setRepositoryPatterns,
    disabled,
}) => {
    const [autoFocusIndex, setAutoFocusIndex] = useState(-1)
    const [repositoryFetchLimit, setRepositoryFetchLimit] = useState(DEFAULT_FETCH_LIMIT)

    const addRepositoryPattern = (): void => {
        setRepositoryPatterns(repositoryPatterns => repositoryPatterns.concat(['']))
        setAutoFocusIndex(repositoryPatterns.length)
        setRepositoryFetchLimit(DEFAULT_FETCH_LIMIT)
    }

    const handleDelete = (index: number): void => {
        setRepositoryPatterns(repositoryPatterns => repositoryPatterns.filter((___, index_) => index !== index_))
        setAutoFocusIndex(index - 1)
        setRepositoryFetchLimit(DEFAULT_FETCH_LIMIT)
    }

    const {
        previewResult: preview,
        previewError,
        isLoadingPreview,
    } = usePreviewRepositoryFilter(repositoryPatterns, repositoryFetchLimit)

    return (
        <div className={styles.container}>
            {repositoryPatterns.map((repositoryPattern, index) => (
                <div key={index} className={styles.inputContainer}>
                    <RepositoryPattern
                        index={index}
                        autoFocus={index === autoFocusIndex}
                        pattern={repositoryPattern}
                        loading={isLoadingPreview}
                        disabled={disabled}
                        setPattern={value =>
                            setRepositoryPatterns(repositoryPatterns =>
                                (repositoryPatterns || []).map((value_, index_) => (index === index_ ? value : value_))
                            )
                        }
                    />
                    <div className={styles.inputActions}>
                        {index > 0 && (
                            <Tooltip content="Delete this repository pattern">
                                <Button
                                    aria-label="Delete the repository pattern"
                                    className={styles.inputAction}
                                    variant="icon"
                                    onClick={() => handleDelete(index)}
                                    disabled={disabled}
                                >
                                    <Icon className="text-danger" aria-hidden={true} svgPath={mdiDelete} />
                                </Button>
                            </Tooltip>
                        )}

                        {index === repositoryPatterns.length - 1 && (
                            <Tooltip content="Add an additional repository pattern">
                                <Button
                                    aria-label="Add an additional repository pattern"
                                    className={styles.inputAction}
                                    variant="icon"
                                    onClick={addRepositoryPattern}
                                    disabled={disabled}
                                >
                                    <Icon className="text-primary" aria-hidden={true} svgPath={mdiPlus} />
                                </Button>
                            </Tooltip>
                        )}
                    </div>
                </div>
            ))}

            <div className="d-flex flex-column">
                {previewError && <ErrorAlert prefix="Error fetching matching repositories" error={previewError} />}
                {preview &&
                    (preview.repositories.length === 0 ? (
                        !(repositoryPatterns.length === 1 && repositoryPatterns[0] === '') && (
                            <Alert variant="warning">
                                This set of repository patterns does not match any repository.
                            </Alert>
                        )
                    ) : (
                        <>
                            {preview.totalMatches > preview.totalCount && (
                                <Alert variant="warning">
                                    <Text weight="medium" className="mb-1">
                                        Too many matching repositories
                                    </Text>
                                    {preview.totalMatches} repositories are matched by this filter. The maximum{' '}
                                    <Code as={Link} to="/site-admin/configuration">
                                        codeIntelAutoIndexing.policyRepositoryMatchLimit
                                    </Code>{' '}
                                    setting is {preview.limit} repositories.
                                </Alert>
                            )}
                            <div className="d-flex justify-content-between">
                                <span>
                                    {preview.totalCount === 1 ? (
                                        <>{preview.totalCount} repository matches</>
                                    ) : (
                                        <>{preview.totalCount} repositories match</>
                                    )}{' '}
                                    {repositoryPatterns.filter(pattern => pattern !== '').length === 1 ? (
                                        <>this pattern</>
                                    ) : (
                                        <>
                                            these {repositoryPatterns.filter(pattern => pattern !== '').length} patterns
                                        </>
                                    )}
                                    {preview.repositories.length < preview.totalCount && (
                                        <> (showing only {preview.repositories.length})</>
                                    )}
                                    :
                                </span>
                                {preview.repositories.length < preview.totalCount && (
                                    <Button
                                        variant="link"
                                        className="p-0"
                                        onClick={() => setRepositoryFetchLimit(preview.totalCount)}
                                    >
                                        Show all {preview.totalCount} repositories
                                    </Button>
                                )}
                            </div>
                            <ul className={classNames('list-group', styles.list)}>
                                {preview.repositories.map(repo => (
                                    <li key={repo.name} className="list-group-item">
                                        {repo.externalRepository && (
                                            <ExternalRepositoryIcon externalRepo={repo.externalRepository} />
                                        )}
                                        <Link to={repo.url} target="_blank" rel="noopener noreferrer">
                                            {repo.name}
                                        </Link>
                                    </li>
                                ))}
                            </ul>
                        </>
                    ))}
            </div>
        </div>
    )
}

interface RepositoryPatternProps {
    index: number
    pattern: string
    setPattern: (value: string) => void
    disabled: boolean
    loading?: boolean
    autoFocus?: boolean
}

const RepositoryPattern: FunctionComponent<RepositoryPatternProps> = ({
    index,
    pattern,
    setPattern,
    disabled,
    autoFocus,
    loading,
}) => {
    const [localPattern, setLocalPattern] = useState('')
    useEffect(() => setLocalPattern(pattern), [pattern])
    const debouncedSetPattern = useMemo(() => debounce(value => setPattern(value), DEBOUNCED_WAIT), [setPattern])

    return (
        <div className={styles.input}>
            <div className="input-group">
                <div className="input-group-prepend">
                    <span className="input-group-text">{index === 0 ? 'Filter' : 'or'}</span>
                </div>
                <Input
                    type="text"
                    inputClassName="text-monospace"
                    value={localPattern}
                    onChange={({ target: { value } }) => {
                        setLocalPattern(value)
                        debouncedSetPattern(value)
                    }}
                    autoFocus={autoFocus}
                    disabled={disabled}
                    required={true}
                    status={loading ? InputStatus.loading : undefined}
                    placeholder={index === 0 ? 'Example: github.com/*' : undefined}
                />
            </div>
        </div>
    )
}
