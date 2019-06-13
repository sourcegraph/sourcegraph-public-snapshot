import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AddIcon from 'mdi-react/AddIcon'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { SlackNotificationRule, ThreadSettings } from '../../../settings'
import { ThreadSettingsForm } from '../components/threadSettingsForm/ThreadSettingsForm'
import CheckIcon from 'mdi-react/CheckIcon'

interface Props extends ExtensionsControllerProps {
    thread: Pick<GQL.IDiscussionThread, 'id' | 'idWithoutKind' | 'title' | 'settings'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings
    extraAction?: JSX.Element | null
}

// tslint:disable: jsx-no-lambda
export const ThreadActionsSlackNotificationRuleForm: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    extensionsController,
}) => {
    // tslint:disable-next-line: no-invalid-template-strings
    const MESSAGE_LINK = '<${check_url}|check ${check_number}: ${check_title}>'
    const DEFAULTS: SlackNotificationRule = {
        message: '🎗️ Your action is needed on ' + MESSAGE_LINK,
        target: '',
        targetReviewers: false,
        remind: false,
    }
    const initialUncommittedSettings: ThreadSettings = {
        ...threadSettings,
        slackNotificationRules: [
            {
                ...DEFAULTS,
                ...(threadSettings.slackNotificationRules ? threadSettings.slackNotificationRules[0] : null),
            },
        ],
    }

    return (
        <ThreadSettingsForm
            thread={thread}
            onThreadUpdate={onThreadUpdate}
            threadSettings={threadSettings}
            initialUncommittedSettings={initialUncommittedSettings}
            extensionsController={extensionsController}
        >
            {({ uncommittedSettings, setUncommittedSettings, isLoading }) => {
                const updateUncommittedSettings = (data: Partial<SlackNotificationRule>) =>
                    setUncommittedSettings({
                        ...uncommittedSettings,
                        slackNotificationRules: [
                            {
                                ...(uncommittedSettings.slackNotificationRules
                                    ? uncommittedSettings.slackNotificationRules[0]
                                    : null),
                                ...data,
                            },
                        ],
                    })
                const { message, target, targetReviewers, remind } = uncommittedSettings.slackNotificationRules
                    ? uncommittedSettings.slackNotificationRules[0]
                    : DEFAULTS

                return (
                    <>
                        <div className="form-group">
                            <label>Preview</label>
                            <div className="rounded border bg-white text-dark p-3 mb-3 d-flex align-items-start">
                                <div className="mr-4 text-muted">12:34</div>
                                {/* TODO!(sqs) hacky interpolation */}
                                <div>
                                    <strong className="text-info mr-2">@sourcegraph</strong>{' '}
                                    {(message || '').replace(MESSAGE_LINK, '')}{' '}
                                    <a href="#" className="text-primary">
                                        check #{thread.idWithoutKind}: {thread.title}
                                    </a>
                                </div>
                            </div>
                        </div>
                        <hr className="my-3" />
                        <div className="form-group">
                            <label htmlFor="thread-actions-slack-notification-rule-form__message">Message</label>
                            <textarea
                                className="form-control"
                                id="thread-actions-slack-notification-rule-form__message"
                                value={message}
                                rows={4}
                                onChange={e =>
                                    updateUncommittedSettings({
                                        message: e.currentTarget.value,
                                    })
                                }
                                required={true}
                            />
                            <small
                                id="thread-pull-request-template-edit-form__description-help"
                                className="form-text text-muted"
                            >
                                Template variables:{' '}
                                <code data-tooltip="The check title">
                                    {'${check_title}'} {/* tslint:disable-line: no-invalid-template-strings */}
                                </code>{' '}
                                &nbsp;{' '}
                                <code data-tooltip="The check number (example: #49)">
                                    {'${check_number}'} {/* tslint:disable-line: no-invalid-template-strings */}
                                </code>{' '}
                                &nbsp;{' '}
                                <code data-tooltip="The URL to the check's page on Sourcegraph">
                                    {'${check_url}'} {/* tslint:disable-line: no-invalid-template-strings */}
                                </code>{' '}
                            </small>
                        </div>
                        <div className="form-group">
                            <label htmlFor="thread-actions-slack-notification-rule-form__target">Recipients</label>
                            <input
                                type="text"
                                className="form-control"
                                id="thread-actions-slack-notification-rule-form__target"
                                aria-describedby="thread-actions-slack-notification-rule-form__target-help"
                                placeholder="@alice, #dev-frontend, @bob"
                                value={target}
                                onChange={e =>
                                    updateUncommittedSettings({
                                        target: e.currentTarget.value,
                                    })
                                }
                            />
                            <small
                                className="form-help text-muted"
                                id="thread-actions-slack-notification-rule-form__target-help"
                            >
                                Supports multiple @usernames and #channels (comma-separated)
                            </small>
                        </div>
                        <div className="form-check mb-3">
                            <input
                                className="form-check-input"
                                type="checkbox"
                                id="thread-actions-slack-notification-rule-form__targetReviewers"
                                checked={targetReviewers}
                                onChange={e =>
                                    updateUncommittedSettings({
                                        targetReviewers: e.currentTarget.checked,
                                    })
                                }
                            />
                            <label
                                className="form-check-label"
                                htmlFor="thread-actions-slack-notification-rule-form__targetReviewers"
                            >
                                Also send to PR reviewers
                            </label>
                        </div>
                        <div className="form-check mb-3">
                            <input
                                className="form-check-input"
                                type="checkbox"
                                id="thread-actions-slack-notification-rule-form__remind"
                                checked={remind}
                                onChange={e =>
                                    updateUncommittedSettings({
                                        remind: e.currentTarget.checked,
                                    })
                                }
                            />
                            <label
                                className="form-check-label"
                                htmlFor="thread-actions-slack-notification-rule-form__remind"
                            >
                                Remind recipients every 48 hours (if action is needed)
                            </label>
                        </div>
                        <div>
                            <button type="submit" className="btn btn-primary mr-2" disabled={isLoading}>
                                {isLoading ? (
                                    <LoadingSpinner className="icon-inline" />
                                ) : (
                                    <CheckIcon className="icon-inline" />
                                )}{' '}
                                Save
                            </button>
                        </div>
                    </>
                )
            }}
        </ThreadSettingsForm>
    )
}
