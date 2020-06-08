import * as H from 'history'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { PatchInterface } from '../../../../../../shared/src/graphql/schema'
import { Observer } from 'rxjs'
import { PatchNode } from './PatchNode'
import React from 'react'
import { HiddenPatchNode } from './HiddenPatchNode'

export interface PatchInterfaceNodeProps extends ThemeProps {
    node: PatchInterface
    campaignUpdates?: Pick<Observer<void>, 'next'>
    history: H.History
    location: H.Location
    /** Shows the publish button */
    enablePublishing: boolean
}

export const PatchInterfaceNode: React.FunctionComponent<PatchInterfaceNodeProps> = ({ node, ...props }) => {
    if (node.__typename === 'Patch') {
        return <PatchNode node={node} {...props} />
    }
    return <HiddenPatchNode />
}
