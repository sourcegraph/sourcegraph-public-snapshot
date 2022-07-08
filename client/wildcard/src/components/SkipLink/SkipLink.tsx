import React, { useEffect } from 'react'

import VisuallyHidden from '@reach/visually-hidden'

import { useSkipLinkContext } from './SkipLinkProvider'

export interface SkipLinkState {
    /**
     * Unique identifier to allow navigating to this skip link.
     */
    id: `skip-to-${string}`
    /**
     * Name to display to the user at the top of the page.
     */
    name: string
}

interface SkipLinkProps extends SkipLinkState {
    /**
     * If the component should render an empty anchor tag to support skip link navigation.
     * Defaults to `true`.
     */
    renderAnchor?: boolean
}

/**
 * Adds a SkipLink to the current page.
 * This will appears when it is navigated to using the keyboard.
 *
 * By default, this component will render a hidden anchor wherever it is used that allows for quick and simple navigation.
 *
 * If you wish for more control, you should set an `id` on your specific element and set `renderAnchor` to `false` here.
 * If you do this, please ensure your `id` exists for as long as the SkipLink is used!
 *
 * **Note:** Skip links should be used sparingly and only for certain parts of the application that make sense to allow quick keyboard access to.
 * If you are trying to improve screen reader navigation, you may want to consider using a [landmark](https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Roles#3._landmark_roles) instead
 */
export const SkipLink: React.FunctionComponent<SkipLinkProps> = ({ id, name, renderAnchor = true }) => {
    const context = useSkipLinkContext()

    useEffect(() => {
        context.addLink({ id, name })

        return () => {
            context.removeLink(id)
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [id])

    if (!renderAnchor) {
        return null
    }

    return (
        <VisuallyHidden>
            <span id={id}>Start of {name}</span>
        </VisuallyHidden>
    )
}
