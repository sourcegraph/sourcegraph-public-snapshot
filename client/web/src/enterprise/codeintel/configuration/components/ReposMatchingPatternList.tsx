import { FunctionComponent } from 'react'

import classNames from 'classnames'

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
    const addRepositoryPattern = (): void =>
        setRepositoryPatterns(repositoryPatterns => (repositoryPatterns || []).concat(['']))

    return (
        <div className="mb-2">
            {repositoryPatterns === null || repositoryPatterns.length === 0 ? (
                <div className="pb-2">
                    This configuration policy applies to all repositories.{' '}
                    {!disabled && (
                        <>
                            To restrict the set of repositories to which this configuration applies,{' '}
                            <span
                                className={styles.addRepositoryPattern}
                                onClick={addRepositoryPattern}
                                aria-hidden="true"
                            >
                                add a repository pattern
                            </span>
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
                                pattern={repositoryPattern}
                                setPattern={value =>
                                    setRepositoryPatterns(repositoryPatterns =>
                                        (repositoryPatterns || []).map((value_, index_) =>
                                            index === index_ ? value : value_
                                        )
                                    )
                                }
                                onDelete={() =>
                                    setRepositoryPatterns(repositoryPatterns =>
                                        (repositoryPatterns || []).filter((___, index_) => index !== index_)
                                    )
                                }
                                disabled={disabled}
                            />
                        ))}
                    </div>

                    {!disabled && (
                        <>
                            <div className="py-3">
                                <span
                                    className={classNames(styles.addRepositoryPattern)}
                                    onClick={addRepositoryPattern}
                                    aria-hidden="true"
                                >
                                    Add a repository pattern
                                </span>
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
