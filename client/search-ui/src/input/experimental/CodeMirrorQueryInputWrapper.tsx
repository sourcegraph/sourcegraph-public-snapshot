import { useEffect, useRef, useState } from 'react'

import { History } from 'history'

import { QueryState, SearchPatternType } from '@sourcegraph/search'

import CodeMirrorQueryInput from './CodeMirrorQueryInput.svelte'
import { Source } from './suggestions'

export interface CodeMirrorQueryInputWrapperProps {
    queryState: QueryState
    onChange: (queryState: QueryState) => void
    onSubmit: () => void
    isLightTheme: boolean
    interpretComments: boolean
    patternType: SearchPatternType
    placeholder: string
    suggestionSource: Source
    history: History
}

export const CodeMirrorQueryInputWrapper: React.FunctionComponent<CodeMirrorQueryInputWrapperProps> = props => {
    const [parent, setParent] = useState<HTMLDivElement | null>(null)
    const [instance, setInstance] = useState<CodeMirrorQueryInput | null>(null)
    const instanceRef = useRef(instance)
    instanceRef.current = instance

    useEffect(() => {
        if (!parent) {
            return
        }
        const instance = new CodeMirrorQueryInput({
            target: parent,
            props,
        })
        setInstance(instance)
        return () => instance.$destroy()
        // props are updated below
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [parent])

    useEffect(() => {
        instanceRef.current?.$set(props)
    }, [props])

    // eslint-disable-next-line react/forbid-dom-props
    return <div ref={setParent} style={{ display: 'contents' }} />
}
