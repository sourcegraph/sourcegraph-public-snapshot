import * as H from 'history'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import VideoInputAntennaIcon from 'mdi-react/VideoInputAntennaIcon'
import React, { useCallback, useMemo, useState } from 'react'
import { Form } from '../../../../branded/src/components/Form'
import { Toggle } from '../../../../branded/src/components/Toggle'
import { AuthenticatedUser } from '../../auth'
import { BreadcrumbSetters, BreadcrumbsProps } from '../../components/Breadcrumbs'
import { PageHeader } from '../../components/PageHeader'
import { PageTitle } from '../../components/PageTitle'

export interface CreateCodeMonitorPageProps extends BreadcrumbsProps, BreadcrumbSetters {
    location: H.Location
    authenticatedUser: AuthenticatedUser | null
}
export const CreateCodeMonitorPage: React.FunctionComponent<CreateCodeMonitorPageProps> = props => {
    props.useBreadcrumb(
        useMemo(
            () => ({
                key: 'Create Code Monitor',
                element: <>Create new code monitor</>,
            }),
            []
        )
    )

    const [showQueryForm, setShowQueryForm] = useState(false)
    const toggleQueryForm: React.MouseEventHandler = useCallback(event => {
        event.preventDefault()
        setShowQueryForm(show => !show)
    }, [])

    const [showEmailNotificationForm, setShowEmailNotificationForm] = useState(false)
    const toggleEmailNotificationForm: React.MouseEventHandler = useCallback(event => {
        event.preventDefault()
        setShowEmailNotificationForm(show => !show)
    }, [])

    const [emailNotificationEnabled, setEmailNotificationEnabled] = useState(true)
    const toggleEmailNotificationEnabled: (value: boolean) => void = useCallback(enabled => {
        setEmailNotificationEnabled(enabled)
    }, [])

    const [active, setActive] = useState(true)
    const toggleActive: (active: boolean) => void = useCallback(active => {
        setActive(active)
    }, [])

    return (
        <div className="container mt-3 web-content">
            <PageTitle title="Create new code monitor" />
            <PageHeader title="Create new code monitor" icon={VideoInputAntennaIcon} />
            Code monitors watch your code for specific triggers and run actions in response.{' '}
            <a href="" target="_blank" rel="noopener">
                {/* TODO: populate link */}
                Learn more
            </a>
            .
            <Form className="my-4">
                <div className="flex mb-4">
                    Name
                    <div>
                        <input type="text" className="form-control my-2" />
                    </div>
                    <small className="text-muted">
                        Give it a short, descriptive name to reference events on Sourcegraph and in notifications. Do
                        not include:{' '}
                        <a href="" target="_blank" rel="noopener">
                            {/* TODO: populate link */}
                            confidential information
                        </a>
                        .
                    </small>
                </div>
                <div className="flex">
                    Owner
                    <select className="form-control my-2 w-auto" disabled={true}>
                        <option value={props.authenticatedUser?.displayName || props.authenticatedUser?.username}>
                            {props.authenticatedUser?.username}
                        </option>
                    </select>
                    <small className="text-muted">Event history and configuration will not be shared.</small>
                </div>
                <hr className="my-4" />
                <div className="mb-4">
                    <h3>Trigger</h3>
                    <div className="card p-3 my-3">
                        <a href="" onClick={toggleQueryForm} className="font-weight-bold">
                            When there are new search results
                        </a>
                        <span className="text-muted">
                            This trigger will fire when new search results are found for a given search query.
                        </span>
                        {showQueryForm && (
                            <>
                                <div className="create-monitor-page__query-input">
                                    <input type="text" className="form-control my-2" />
                                    <a
                                        href=""
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className="create-monitor-page__query-input-preview-link"
                                    >
                                        {/* TODO: populate link */}
                                        Preview results <OpenInNewIcon />
                                    </a>
                                </div>
                                <small className="text-muted mb-4">
                                    Code monitors only support <code className="bg-code">type:diff</code> and{' '}
                                    <code className="bg-code">type:commit</code> search queries.
                                </small>
                                <div>
                                    <button type="button" className="btn btn-outline-secondary mr-1">
                                        Continue
                                    </button>
                                    <button type="button" className="btn btn-outline-secondary">
                                        Cancel
                                    </button>
                                </div>
                            </>
                        )}
                    </div>
                    {!showQueryForm && (
                        <small className="text-muted">
                            {' '}
                            What other events would you like to monitor? {/* TODO: populate link */}
                            <a href="" target="_blank" rel="noopener">
                                {/* TODO: populate link */}
                                Share feedback.
                            </a>
                        </small>
                    )}
                </div>
                <div>
                    <h3>Actions</h3>
                    <p>Run any number of actions in response to an event</p>
                    <div className="card p-3 my-3">
                        {/* This should be its own component when you can add multiple email actions */}
                        <a href="" onClick={toggleEmailNotificationForm} className="font-weight-bold">
                            Send email notifications
                        </a>
                        {showEmailNotificationForm && (
                            <>
                                <div className="mt-4">
                                    Recipients
                                    <input
                                        type="text"
                                        className="form-control my-2"
                                        value={`${props.authenticatedUser?.email || ''} (you)`}
                                        disabled={true}
                                    />
                                    <small className="text-muted">
                                        Code monitors are currently limited to sending emails to your primary email
                                        address.
                                    </small>
                                </div>
                                <div className="flex my-4">
                                    <Toggle
                                        title="Enabled"
                                        value={emailNotificationEnabled}
                                        onToggle={toggleEmailNotificationEnabled}
                                        className="mr-2"
                                    />
                                    Enabled
                                </div>
                                <div>
                                    <button type="button" className="btn btn-outline-secondary mr-1">
                                        Done
                                    </button>
                                    <button type="button" className="btn btn-outline-secondary">
                                        Cancel
                                    </button>
                                </div>
                            </>
                        )}
                    </div>
                    <small className="text-muted">
                        What other actions would you like to do?{' '}
                        <a href="" target="_blank" rel="noopener">
                            {/* TODO: populate link */}
                            Share feedback.
                        </a>
                    </small>
                </div>
                <div>
                    <div className="d-flex my-4">
                        <div>
                            <Toggle title="Active" value={active} onToggle={toggleActive} className="mr-2" />{' '}
                        </div>
                        <div className="flex-column">
                            <div>Active</div>
                            <div className="text-muted">We will watch for the trigger and run actions in response</div>
                        </div>
                    </div>
                    <div className="flex my-4">
                        <button type="button" className="btn btn-primary mr-2">
                            Create code monitor
                        </button>
                        <button type="button" className="btn btn-outline-secondary">
                            Cancel
                        </button>
                    </div>
                </div>
            </Form>
        </div>
    )
}
