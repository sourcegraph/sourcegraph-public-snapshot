import { useEffect, useRef } from 'react'

import { Tree, type TreeNode, type TreeProps, type TreeRef } from '@sourcegraph/wildcard/src'

export interface FocusableTreeProps {
    focusKey?: string
}

/**
 * Renders {@link Tree} and focuses it when {@link focusKey} changes.
 */
export function FocusableTree<N extends TreeNode>(props: FocusableTreeProps & TreeProps<N>): JSX.Element {
    const { focusKey, ...rest } = props

    const ref = useRef<TreeRef>(null)

    useEffect(() => {
        if (ref.current && focusKey) {
            ref.current.focus()
        }
    }, [focusKey])

    return <Tree ref={ref} {...rest} />
}
