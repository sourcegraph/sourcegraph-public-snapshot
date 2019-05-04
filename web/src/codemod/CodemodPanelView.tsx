import H from 'history'
import RayEndArrowIcon from 'mdi-react/RayEndArrowIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import React, { useEffect } from 'react'
import { of } from 'rxjs'
import { PanelViewWithComponent } from '../../../shared/src/api/client/services/view'
import { ContributableViewContainer } from '../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { DismissibleAlert } from '../components/DismissibleAlert'
import { Form } from '../components/Form'
import { CODEMOD_PANEL_VIEW_ID } from './contributions'

interface Props extends ExtensionsControllerProps<'services'> {
    navbarSearchQuery: string
    location: H.Location
}

const CodemodPanelView: React.FunctionComponent<Props> = ({ navbarSearchQuery, location, extensionsController }) => (
    <div className="p-3">
        <DismissibleAlert className="alert-info" partialStorageKey="codemod-experimental">
            Code modification is an experimental feature.
        </DismissibleAlert>
        <Form className="form" onSubmit={() => void 0}>
            <div className="row">
                <div className="col-md-10 col-lg-8">
                    <div className="form-group">
                        <label htmlFor="codemod-panel-view__title">Pull request title</label>
                        <input
                            type="text"
                            className="form-control"
                            id="codemod-panel-view__title"
                            placeholder="PrintMultiFileDiff -> RenderDiffs (Sourcegraph codemod)"
                        />
                    </div>
                    <div className="form-group">
                        <label htmlFor="codemod-panel-view__branchName">Branch</label>
                        <div className="d-flex align-items-center">
                            <code
                                className="border rounded text-muted p-1"
                                data-tooltip="Changing the base branch is not yet supported"
                            >
                                <SourceBranchIcon className="icon-inline mr-1" />
                                master
                            </code>{' '}
                            <RayEndArrowIcon className="icon-inline mx-2 text-muted" />
                            <input
                                type="text"
                                className="form-control form-control-sm flex-0 w-auto text-monospace"
                                id="codemod-panel-view__branchName"
                                placeholder="codemod/printmultifilediff-renderdiffs"
                                size={30}
                            />
                        </div>
                    </div>
                    <div className="form-group">
                        <label htmlFor="codemod-panel-view__description">Pull request description</label>
                        <textarea
                            className="form-control"
                            id="codemod-panel-view__description"
                            aria-describedby="codemod-panel-view__description-help"
                            rows={4}
                            defaultValue={
                                'Sourcegraph codemod: [${query}](${query_url})\n\nRelated PRs: ${related_links}'
                            }
                        />
                        <small id="codemod-panel-view__description-help" className="form-text text-muted">
                            {/* tslint:disable-next-line: no-invalid-template-strings */}
                            Template variables: <code data-tooltip="The full search query">
                                {'${query}'}
                            </code> &nbsp;{' '}
                            <code data-tooltip="The URL to the search results page on Sourcegraph">
                                {'${query_url}'}
                            </code>{' '}
                            &nbsp;{' '}
                            <code data-tooltip="Formatted links to all other pull requests (in other repositories) created by this codemod">
                                {'${related_links}'}
                            </code>
                        </small>
                    </div>
                    <button type="submit" className="btn btn-primary d-flex align-items-center">
                        <SourcePullIcon className="icon-inline mr-1" /> Create 1 pull request
                    </button>
                </div>
            </div>
        </Form>
    </div>
)

export const CodemodPanelViewRegistration: React.FunctionComponent<Props> = props => {
    useEffect(() => {
        const subscription = props.extensionsController.services.views.registerProvider(
            { container: ContributableViewContainer.Panel, id: CODEMOD_PANEL_VIEW_ID },
            of<PanelViewWithComponent | null>({
                title: 'Codemod',
                content: '',
                priority: 100,
                reactElement: <CodemodPanelView {...props} />,
            })
        )
        return () => subscription.unsubscribe()
    })

    return null
}
