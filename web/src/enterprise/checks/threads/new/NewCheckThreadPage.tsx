import H from 'history'
import React, { useCallback, useEffect, useState } from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'
import { CheckTemplate } from '../../../../../../shared/src/api/client/services/checkTemplates'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PageTitle } from '../../../../components/PageTitle'
import { CheckTemplateItem } from '../../components/CheckTemplateItem'
import { ChecksAreaContext } from '../../global/ChecksArea'
import { CheckThreadTemplateSelectFormControl } from './CheckThreadTemplateSelectFormControl'
import { NewCheckThreadForm } from './NewCheckThreadForm'

interface Props
    extends Pick<ChecksAreaContext, 'project'>,
        ExtensionsControllerProps<'services'>,
        RouteComponentProps<{}> {
    history: H.History
    location: H.Location
}

/**
 * A page for adding a new check based on one of the registered check templates.
 */
export const NewCheckThreadPage: React.FunctionComponent<Props> = ({
    project,
    history,
    location,
    match: { url: baseUrl },
    extensionsController,
}) => {
    const checkTemplateId = new URLSearchParams(location.search).get('template')
    const [checkTemplate, setCheckTemplate] = useState<CheckTemplate>()
    useEffect(() => {
        if (checkTemplateId === null) {
            setCheckTemplate(undefined)
            return undefined
        }
        const subscription = extensionsController.services.checkTemplates
            .getCheckTemplate(checkTemplateId)
            .subscribe(checkTemplate => setCheckTemplate(checkTemplate || undefined))
        return () => subscription.unsubscribe()
    }, [checkTemplateId, extensionsController.services.checkTemplates])

    const urlForCheckTemplate = useCallback(
        (checkTemplateId: string | null): H.LocationDescriptor =>
            checkTemplateId !== null ? `${baseUrl}?${new URLSearchParams({ template: checkTemplateId })}` : baseUrl,
        [baseUrl]
    )

    return (
        <div className="new-check-thread-page container mt-3">
            <PageTitle title="New check" />
            <h1 className="mb-3">New check</h1>
            <div className="row">
                <div className="col-md-9 col-lg-8 col-xl-7">
                    <label>Type</label>
                    {!checkTemplate ? (
                        <CheckThreadTemplateSelectFormControl
                            urlForCheckTemplate={urlForCheckTemplate}
                            extensionsController={extensionsController}
                        />
                    ) : (
                        <>
                            <CheckTemplateItem
                                checkTemplate={checkTemplate}
                                className="border rounded"
                                endFragment={
                                    <Link
                                        to={urlForCheckTemplate(null)}
                                        className="btn btn-secondary text-decoration-none"
                                        data-tooltip="Choose a different template"
                                    >
                                        Change
                                    </Link>
                                }
                            />
                            <NewCheckThreadForm
                                project={project}
                                checkTemplate={checkTemplate}
                                className="mt-3"
                                history={history}
                            />
                        </>
                    )}
                </div>
            </div>
        </div>
    )
}
