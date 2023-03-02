import { useEffect, useRef } from 'react'

import { Tree, TreeNode, TreeProps } from '@sourcegraph/wildcard/src'

export interface FocusableTreeProps {
    focusKey?: string
}

/**
 * Renders {@link Tree} and focuses it when {@link focusKey} changes.
 */
export function FocusableTree<N extends TreeNode>(props: FocusableTreeProps & TreeProps<N>) {
    const { focusKey, ...rest } = props

    const ref = useRef<HTMLUListElement>(null)

    useEffect(() => {
        if (ref.current && focusKey) {
            ref.current.focus()
        }
    }, [focusKey])

    return <Tree ref={ref} {...rest} />
}
