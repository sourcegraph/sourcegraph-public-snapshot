import { FunctionComponent, useEffect, useMemo, useState } from 'react'

import { mdiDelete, mdiPlus } from '@mdi/js'
import classNames from 'classnames'
import { debounce } from 'lodash'

import {
    Alert,
    Button,
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

interface RepositoryPatternListProps {
    repositoryPatterns: string[]
    setRepositoryPatterns: (updater: (patterns: string[] | null) => string[] | null) => void
    disabled: boolean
}

export const RepositoryPatternList: FunctionComponent<RepositoryPatternListProps> = ({
    repositoryPatterns,
    setRepositoryPatterns,
    disabled,
}) => {
    const [autoFocusIndex, setAutoFocusIndex] = useState(-1)

    const addRepositoryPattern = (): void => {
        setRepositoryPatterns(repositoryPatterns =>
            repositoryPatterns && repositoryPatterns.length > 0 ? repositoryPatterns.concat(['']) : ['*']
        )
        setAutoFocusIndex(repositoryPatterns?.length ?? -1)
    }

    const handleDelete = (index: number): void => {
        setRepositoryPatterns(repositoryPatterns =>
            (repositoryPatterns || []).filter((___, index_) => index !== index_)
        )
        setAutoFocusIndex(-1)
    }

    const {
        previewResult: preview,
        isLoadingPreview: previewLoading,
        previewError,
    } = usePreviewRepositoryFilter(repositoryPatterns || [])

    if (!repositoryPatterns) {
        // todo remove this
        return null
    }

    return (
        <div className={styles.container}>
            <div>
                {repositoryPatterns.map((repositoryPattern, index) => (
                    <div key={index} className={styles.inputContainer}>
                        {/* TODO Maybe debounce this input */}
                        <RepositoryPattern
                            index={index}
                            autoFocus={index === autoFocusIndex}
                            pattern={repositoryPattern}
                            setPattern={value =>
                                setRepositoryPatterns(repositoryPatterns =>
                                    (repositoryPatterns || []).map((value_, index_) =>
                                        index === index_ ? value : value_
                                    )
                                )
                            }
                            onDelete={() => handleDelete(index)}
                            disabled={disabled}
                        />
                        <div className={styles.inputControls}>
                            {index > 0 && (
                                <Tooltip content="Delete this repository pattern">
                                    <Button
                                        aria-label="Delete the repository pattern"
                                        className={styles.inputControl}
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
                                        className={styles.inputControl}
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
            </div>

            <div className="form-group d-flex flex-column">
                <div>
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
                                    <Alert variant="danger">
                                        Each policy pattern can match a maximum of {preview.limit} repositories. There
                                        are {preview.totalMatches - preview.totalCount} additional repositories that
                                        match the filter not covered by this policy. Narrow the policy to a smaller set
                                        of repositories or increase the system limit.
                                    </Alert>
                                )}

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

                                <ul className="list-group p2">
                                    {preview.repositories.map(repo => (
                                        <li key={repo.name} className="list-group-item">
                                            {repo.externalRepository && (
                                                <ExternalRepositoryIcon externalRepo={repo.externalRepository} />
                                            )}
                                            <Link to={repo.url}>{repo.name}</Link>
                                        </li>
                                    ))}
                                </ul>
                            </>
                        ))}
                </div>
            </div>
        </div>
    )
}

interface RepositoryPatternProps {
    index: number
    pattern: string
    setPattern: (value: string) => void
    onDelete: () => void
    disabled: boolean
    autoFocus?: boolean
}

const RepositoryPattern: FunctionComponent<RepositoryPatternProps> = ({
    index,
    pattern,
    setPattern,
    onDelete,
    disabled,
    autoFocus,
}) => {
    const [localPattern, setLocalPattern] = useState('')
    useEffect(() => setLocalPattern(pattern), [pattern])
    const debouncedSetPattern = useMemo(() => debounce(value => setPattern(value), DEBOUNCED_WAIT), [setPattern])

    const { isLoadingPreview: previewLoading, previewError } = usePreviewRepositoryFilter([localPattern])

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
                    status={previewLoading ? InputStatus.loading : undefined}
                />
                {/* {localPattern === '' && (
                    <InputErrorMessage
                        message="Please supply a value."
                        className="d-flex align-items-center mt-0 ml-3"
                    />
                )} */}
            </div>

            {previewError && (
                <ErrorAlert prefix="Error fetching matching repositories" error={previewError} className="mt-2" />
            )}
        </div>
    )
}
