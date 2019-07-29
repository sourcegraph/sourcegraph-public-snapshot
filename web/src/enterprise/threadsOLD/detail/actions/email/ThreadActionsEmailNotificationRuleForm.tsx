import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import CheckIcon from 'mdi-react/CheckIcon'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { EmailNotificationRule, ThreadSettings } from '../../../settings'
import { ThreadSettingsForm } from '../components/threadSettingsForm/ThreadSettingsForm'

interface Props extends ExtensionsControllerProps {
    thread: Pick<GQL.IDiscussionThread, 'id' | 'idWithoutKind' | 'title' | 'settings'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings
    extraAction?: JSX.Element | null
}

// tslint:disable: jsx-no-lambda
export const ThreadActionsEmailNotificationRuleForm: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    extensionsController,
}) => {
    const DEFAULTS: EmailNotificationRule = {
        subject: `🎯 ${thread.title}`,
        to: '',
        cc: '',
    }
    const initialUncommittedSettings: ThreadSettings = {
        ...threadSettings,
        emailNotificationRules: [
            {
                ...DEFAULTS,
                ...(threadSettings.emailNotificationRules ? threadSettings.emailNotificationRules[0] : null),
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
                const updateUncommittedSettings = (data: Partial<EmailNotificationRule>) =>
                    setUncommittedSettings({
                        ...uncommittedSettings,
                        emailNotificationRules: [
                            {
                                ...(uncommittedSettings.emailNotificationRules
                                    ? uncommittedSettings.emailNotificationRules[0]
                                    : null),
                                ...data,
                            },
                        ],
                    })
                const { subject, to, cc } = uncommittedSettings.emailNotificationRules
                    ? uncommittedSettings.emailNotificationRules[0]
                    : DEFAULTS

                return (
                    <>
                        <div className="form-group">
                            <label htmlFor="thread-actions-email-notification-rule-form__subject">Subject</label>
                            <input
                                type="text"
                                className="form-control"
                                id="thread-actions-email-notification-rule-form__subject"
                                value={subject}
                                onChange={e =>
                                    updateUncommittedSettings({
                                        subject: e.currentTarget.value,
                                    })
                                }
                                required={true}
                            />
                        </div>
                        <div className="form-group">
                            <label htmlFor="thread-actions-email-notification-rule-form__to">To</label>
                            <input
                                type="text"
                                className="form-control"
                                id="thread-actions-email-notification-rule-form__to"
                                aria-describedby="thread-actions-email-notification-rule-form__to-help"
                                placeholder="alice@example.com, bob@example.com"
                                value={to}
                                onChange={e =>
                                    updateUncommittedSettings({
                                        to: e.currentTarget.value,
                                    })
                                }
                                required={true}
                            />
                            <small
                                className="form-help text-muted"
                                id="thread-actions-email-notification-rule-form__to-help"
                            >
                                Separate multiple email addresses with commas
                            </small>
                        </div>
                        <div className="form-group">
                            <label htmlFor="thread-actions-email-notification-rule-form__cc">Cc</label>
                            <input
                                type="text"
                                className="form-control"
                                id="thread-actions-email-notification-rule-form__cc"
                                aria-describedby="thread-actions-email-notification-rule-form__cc-help"
                                placeholder="alice@example.com, bob@example.com"
                                value={cc}
                                onChange={e =>
                                    updateUncommittedSettings({
                                        cc: e.currentTarget.value,
                                    })
                                }
                            />
                            <small
                                className="form-help text-muted"
                                id="thread-actions-email-notification-rule-form__cc-help"
                            >
                                Separate multiple email addresses with commas
                            </small>
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
