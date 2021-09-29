import { Shortcut } from '@slimsag/react-shortcuts'
import React, { useState } from 'react'

import { downloadFilenames, FuzzyFSM } from '@sourcegraph/shared/src/fuzzyFinder/fsm'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'

import {
    KEYBOARD_SHORTCUT_CLOSE_FUZZY_FINDER,
    KEYBOARD_SHORTCUT_FUZZY_FINDER,
} from '../../keyboardShortcuts/keyboardShortcuts'
import { parseBrowserRepoURL } from '../../util/url'

import { FuzzyModal } from './FuzzyModal'

const DEFAULT_MAX_RESULTS = 100

export interface FuzzyFinderProps extends PlatformContextProps<'requestGraphQL' | 'urlToFile' | 'clientApplication'> {
    repoName: string
    commitID: string

    /**
     * The maximum number of files a repo can have to use case-insensitive fuzzy finding.
     *
     * Case-insensitive fuzzy finding is more expensive to compute compared to
     * word-sensitive fuzzy finding.  The fuzzy modal will use case-insensitive
     * fuzzy finding when the repo has fewer files than this number, and
     * word-sensitive fuzzy finding otherwise.
     */
    caseInsensitiveFileCountThreshold?: number
}

export const FuzzyFinder: React.FunctionComponent<FuzzyFinderProps> = props => {
    const [isVisible, setIsVisible] = useState(false)

    // The state machine of the fuzzy finder. See `FuzzyFSM` for more details
    // about the state transititions.
    const [fsm, setFsm] = useState<FuzzyFSM>({ key: 'empty' })

    return (
        <>
            <Shortcut
                {...KEYBOARD_SHORTCUT_FUZZY_FINDER.keybindings[0]}
                onMatch={() => {
                    setIsVisible(true)
                    const input = document.querySelector<HTMLInputElement>('#fuzzy-modal-input')
                    input?.focus()
                    input?.select()
                }}
            />
            <Shortcut {...KEYBOARD_SHORTCUT_CLOSE_FUZZY_FINDER.keybindings[0]} onMatch={() => setIsVisible(false)} />
            {isVisible && (
                <FuzzyModal
                    {...props}
                    isVisible={isVisible}
                    onClose={() => setIsVisible(false)}
                    initialQuery=""
                    initialMaxResults={DEFAULT_MAX_RESULTS}
                    fsm={fsm}
                    setFsm={setFsm}
                    downloadFilenames={() => downloadFilenames(props)}
                    parseRepoUrl={() => parseBrowserRepoURL(location.pathname + location.search + location.hash)}
                />
            )}
        </>
    )
}
