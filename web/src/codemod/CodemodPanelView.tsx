import H from 'history'
import React, { useEffect } from 'react'
import { of } from 'rxjs'
import { PanelViewWithComponent } from '../../../shared/src/api/client/services/view'
import { ContributableViewContainer } from '../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { CODEMOD_PANEL_VIEW_ID } from './contributions'

interface Props extends ExtensionsControllerProps<'services'> {
    navbarSearchQuery: string
    location: H.Location
}

export const CodemodPanelView: React.FunctionComponent<Props> = ({
    navbarSearchQuery,
    location,
    extensionsController,
}) => {
    useEffect(() => {
        const subscription = extensionsController.services.views.registerProvider(
            { container: ContributableViewContainer.Panel, id: CODEMOD_PANEL_VIEW_ID },
            of<PanelViewWithComponent | null>({
                title: 'Codemod',
                content: '',
                priority: 100,
                reactElement: (
                    <div className="p-2">
                        Hello, <strong>world!</strong>
                        {navbarSearchQuery}
                    </div>
                ),
            })
        )
        return () => subscription.unsubscribe()
    })

    return null
}
