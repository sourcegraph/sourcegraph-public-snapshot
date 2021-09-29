import classNames from 'classnames'
import LanguageTypescriptIcon from 'mdi-react/LanguageTypescriptIcon'
import React, { useMemo, useState } from 'react'
import { useHistory } from 'react-router'

import { WorkspaceRoot } from '@sourcegraph/extension-api-types'
import { useDebounce } from '@sourcegraph/wildcard'

import { LanguageIcon } from '../../../components/languageIcons'
import { HighlightedLink } from '../../../fuzzyFinder/components/HighlightedLink'
import { downloadFilenames, FuzzyFinderProps, FuzzyFSM } from '../../../fuzzyFinder/fsm'
import { FuzzyModalProps, FuzzyModalState, renderFuzzyResult } from '../../../fuzzyFinder/FuzzyModal'
import { getModeFromPath } from '../../../languages'
import { PlatformContextProps } from '../../../platform/context'
import { parseRepoURI } from '../../../util/url'

import styles from './FuzzyFinderResult.module.scss'
import { Message } from './Message'
import { NavigableList } from './NavigableList'

interface FuzzyFinderResultProps extends PlatformContextProps<'requestGraphQL' | 'urlToFile' | 'clientApplication'> {
    value: string
    onClick: () => void
    workspaceRoot: WorkspaceRoot | undefined
}

// Note: filenames don't have to be in monospace font.

export const FuzzyFinderResult: React.FC<FuzzyFinderResultProps> = ({
    value,
    onClick,
    workspaceRoot,
    platformContext,
}) => {
    const history = useHistory()

    // The state machine of the fuzzy finder. See `FuzzyFSM` for more details
    // about the state transititions.
    const [fsm, setFsm] = useState<FuzzyFSM>({ key: 'empty' })

    // TODO: use for infinite scrolling list
    const [maxResults, setMaxResults] = useState(100)

    const repoUrl = useMemo(() => (workspaceRoot?.uri ? parseRepoURI(workspaceRoot.uri) : null), [workspaceRoot?.uri])

    // TODO: run fuzzy search in web worker
    const debouncedValue = useDebounce(value, 300)

    const fuzzyResult = useMemo(() => {
        const commitID = repoUrl?.commitID ?? workspaceRoot?.inputRevision ?? ''
        const repoName = repoUrl?.repoName ?? ''

        const fuzzyFinderProps: FuzzyFinderProps = {
            repoName,
            commitID,
            platformContext,
        }

        const state: FuzzyModalState = {
            query: debouncedValue,
            maxResults,
        }
        const fuzzyModalProps: Pick<
            FuzzyModalProps,
            | 'fsm'
            | 'setFsm'
            | 'downloadFilenames'
            | 'parseRepoUrl'
            | 'onClose'
            | 'repoName'
            | 'commitID'
            | 'platformContext'
        > = {
            fsm,
            setFsm,
            parseRepoUrl: () => ({ ...repoUrl, commitID, repoName }), // TODO error handling,
            downloadFilenames: () => downloadFilenames(fuzzyFinderProps),
            commitID,
            repoName,
            onClose: onClick,
            platformContext,
        }
        return renderFuzzyResult(fuzzyModalProps, state)
    }, [debouncedValue, onClick, fsm, platformContext, repoUrl, workspaceRoot?.inputRevision, maxResults])

    if (!workspaceRoot) {
        return <Message type="muted">Navigate to a repo to use fuzzy finder</Message>
    }

    // TODO: language icon by file extension

    const onFuzzyResultClick = (url?: string): void => {
        if (url) {
            if (platformContext.clientApplication === 'sourcegraph') {
                history.push(url)
            } else {
                window.location.href = 'https://sourcegraph.test:3443' + url
            }
        }
    }

    // const renderLanguageIcon

    return (
        <>
            {fuzzyResult.element && <Message>{fuzzyResult.element}</Message>}
            <NavigableList items={fuzzyResult.linksToRender}>
                {(file, { active }) => (
                    <NavigableList.Item
                        key={file.text}
                        active={active}
                        onClick={() => {
                            onFuzzyResultClick(file.url)
                            file.onClick?.()
                        }}
                    >
                        <span role="link" data-href={file.url ?? ''} className={styles.linkContainer}>
                            <LanguageIcon
                                language={getModeFromPath(file.text)}
                                className={classNames(styles.languageIcon, 'icon-inline')}
                            />
                            <HighlightedLink
                                {...{
                                    ...file,
                                    // Opt out of link rendering
                                    url: undefined,
                                }}
                            />
                        </span>
                    </NavigableList.Item>
                )}
            </NavigableList>
        </>
    )
}

// function useFuzzyFSM() {}
