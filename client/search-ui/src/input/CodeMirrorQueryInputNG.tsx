import { QueryState, SearchPatternType } from '@sourcegraph/search'
import { useEffect, useState } from 'react'
import CodeMirrorQueryInput from './CodeMirrorQueryInput.svelte'

interface CodeMirrorQueryInputNGProps {
    queryState: QueryState
    onChange: (queryState: QueryState) => void
    isLightTheme: boolean
    interpretComments: boolean
    patternType: SearchPatternType
    placeholder: string
}

export const CodeMirrorQueryInputNG: React.FunctionComponent<CodeMirrorQueryInputNGProps> = props => {
    const [parent, setParent] = useState<HTMLDivElement | null>(null)
    const [instance, setInstance] = useState<CodeMirrorQueryInput | null>(null)

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
    }, [parent])

    useEffect(() => {
        instance?.$set(props)
    }, [props])

    return <div ref={setParent} style={{ display: 'contents' }} />
}
