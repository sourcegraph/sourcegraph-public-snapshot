import { FunctionComponent, useState } from 'react'

import classNames from 'classnames'

import { Button } from '@sourcegraph/wildcard'

import { RepositoryPreview } from './RepositoryPreview'
import { ReposMatchingPattern } from './ReposMatchingPattern'

import styles from './ReposMatchingPatternList.module.scss'

export interface ReposMatchingPatternListProps {
    repositoryPatterns: string[] | null
    setRepositoryPatterns: (updater: (patterns: string[] | null) => string[] | null) => void
    disabled: boolean
}

export const ReposMatchingPatternList: FunctionComponent<React.PropsWithChildren<ReposMatchingPatternListProps>> = ({
    repositoryPatterns,
    setRepositoryPatterns,
    disabled,
}) => {
    const [autoFocusIndex, setAutoFocusIndex] = useState(-1)

    const addRepositoryPattern = (): void => {
        setRepositoryPatterns(repositoryPatterns => (repositoryPatterns || []).concat(['']))
        setAutoFocusIndex(repositoryPatterns?.length ?? -1)
    }

    const handleDelete = (index: number): void => {
        setRepositoryPatterns(repositoryPatterns =>
            (repositoryPatterns || []).filter((___, index_) => index !== index_)
        )
        setAutoFocusIndex(-1)
    }

    return (
        <div className="mb-2">
            {repositoryPatterns === null || repositoryPatterns.length === 0 ? (
                <div className="pb-2">
                    This configuration policy applies to all repositories.{' '}
                    {!disabled && (
                        <>
                            To restrict the set of repositories to which this configuration applies,{' '}
                            <Button
                                variant="link"
                                className={styles.addRepositoryPattern}
                                onClick={addRepositoryPattern}
                            >
                                add a repository pattern
                            </Button>
                            .
                        </>
                    )}
                </div>
            ) : (
                <>
                    <div className={styles.grid}>
                        {repositoryPatterns.map((repositoryPattern, index) => (
                            <ReposMatchingPattern
                                key={index}
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
                        ))}
                    </div>

                    {!disabled && (
                        <>
                            <div className="py-3">
                                <Button
                                    variant="link"
                                    className={classNames(styles.addRepositoryPattern)}
                                    onClick={addRepositoryPattern}
                                >
                                    Add a repository pattern
                                </Button>
                            </div>
                        </>
                    )}

                    <div className={classNames(styles.preview, 'form-group d-flex flex-column')}>
                        <RepositoryPreview patterns={repositoryPatterns} />
                    </div>
                </>
            )}
        </div>
    )
}
