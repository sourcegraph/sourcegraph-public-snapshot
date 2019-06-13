import H from 'history'
import RayEndArrowIcon from 'mdi-react/RayEndArrowIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import React, { useEffect } from 'react'
import { of } from 'rxjs'
import { PanelViewWithComponent } from '../../../../../shared/src/api/client/services/view'
import { ContributableViewContainer } from '../../../../../shared/src/api/protocol'
import { TabsWithURLViewStatePersistence } from '../../../../../shared/src/components/Tabs'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { DismissibleAlert } from '../../../components/DismissibleAlert'
import { Form } from '../../../components/Form'
import { parseSearchURLQuery } from '../../../search'
import { CODEMOD_PANEL_VIEW_ID } from '../contributions/search'
import { queryFindAndReplaceOptions } from '../query'

interface Props extends ExtensionsControllerProps<'services'> {
    navbarSearchQuery: string
    location: H.Location
    history: H.History
}

const CodemodPanelView: React.FunctionComponent<Props> = ({ navbarSearchQuery, location, extensionsController }) => {
    return (
        <div className="p-3">
            <DismissibleAlert className="alert-info" partialStorageKey="codemod-experimental">
                Code modification is an experimental feature.
            </DismissibleAlert>
        </div>
    )
}

// export const CodemodPanelViewRegistration: React.FunctionComponent<Props> = props => {
//     useEffect(() => {
//         const subscription = props.extensionsController.services.views.registerProvider(
//             { container: ContributableViewContainer.Panel, id: CODEMOD_PANEL_VIEW_ID },
//             of<PanelViewWithComponent | null>({
//                 title: 'Codemod',
//                 content: '',
//                 priority: 100,
//                 reactElement: <CodemodPanelView {...props} />,
//             })
//         )
//         return () => subscription.unsubscribe()
//     })
//
//     useEffect(() => {
//         // Open panel if query contains `replace:` and panel is not open.
//         //
//         // TODO!(sqs): this makes it so the panel "x" close button is a noop and the panel can't be closed
//         if (
//             queryFindAndReplaceOptions(parseSearchURLQuery(props.location.search) || '').replace &&
//             TabsWithURLViewStatePersistence.readFromURL(props.location, [
//                 { id: '', label: '' },
//                 { id: CODEMOD_PANEL_VIEW_ID, label: '' },
//             ]) !== CODEMOD_PANEL_VIEW_ID
//         ) {
//             props.history.push(TabsWithURLViewStatePersistence.urlForTabID(props.location, CODEMOD_PANEL_VIEW_ID))
//         }
//     })
//
//     return null
// }
