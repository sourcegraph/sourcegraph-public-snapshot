import RayEndArrowIcon from 'mdi-react/RayEndArrowIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import React, { useCallback } from 'react'
import { RuleActionTypeComponentContext, RuleActionType } from '..'
import { GitPullRequestIcon } from '../../../../../../util/octicons'

export interface BranchReviewRuleAction {
    type: 'branchReview'
    title: string
    branch: string
    description: string
}

const BranchReviewRuleActionFormControl: React.FunctionComponent<
    RuleActionTypeComponentContext<BranchReviewRuleAction>
> = ({ value, onChange }) => {
    const onTitleChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => {
            onChange({ ...value, title: e.currentTarget.value })
        },
        [onChange, value]
    )
    const onBranchChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => {
            onChange({ ...value, branch: e.currentTarget.value })
        },
        [onChange, value]
    )
    const onDescriptionChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(
        e => {
            onChange({ ...value, description: e.currentTarget.value })
        },
        [onChange, value]
    )
    return (
        <>
            <div className="form-group">
                <label htmlFor="rule-action-form-control__title">Pull request title</label>
                <input
                    type="text"
                    className="form-control"
                    id="rule-action-form-control__title"
                    value={value.title}
                    onChange={onTitleChange}
                />
            </div>
            <div className="form-group">
                <label htmlFor="rule-action-form-control__branch">Branch</label>
                <div className="d-flex align-items-center">
                    <code
                        className="border rounded text-muted p-1"
                        data-tooltip="Changing the base branch is not yet supported"
                    >
                        <SourceBranchIcon className="icon-inline mr-1" />
                        master {/* TODO!(sqs): un-hardcode master */}
                    </code>{' '}
                    <RayEndArrowIcon className="icon-inline mx-2 text-muted" />
                    <input
                        type="text"
                        className="form-control form-control-sm flex-0 w-auto text-monospace"
                        id="rule-action-form-control__branch"
                        size={30}
                        value={value.branch}
                        onChange={onBranchChange}
                    />
                </div>
            </div>
            <div className="form-group">
                <label htmlFor="rule-action-form-control__description">Pull request description</label>
                <textarea
                    className="form-control"
                    id="rule-action-form-control__description"
                    aria-describedby="rule-action-form-control__description-help"
                    value={value.description}
                    onChange={onDescriptionChange}
                    rows={7}
                />
                <small id="rule-action-form-control__description-help" className="form-text text-muted">
                    Template variables:{' '}
                    <code data-tooltip="The campaign name">
                        {'${campaign_name}'} {/* eslint-disable-line no-template-curly-in-string */}
                    </code>{' '}
                    &nbsp;{' '}
                    <code data-tooltip="The campaign number (example: #49)">
                        {'${campaign_number}'} {/* eslint-disable-line no-template-curly-in-string */}
                    </code>{' '}
                    &nbsp;{' '}
                    <code data-tooltip="The URL to the campaign's page on Sourcegraph">
                        {'${campaign_url}'} {/* eslint-disable-line no-template-curly-in-string */}
                    </code>
                </small>
            </div>
        </>
    )
}

export const BranchReviewRuleActionType: RuleActionType<'branchReview', BranchReviewRuleAction> = {
    id: 'branchReview',
    title: 'Pull request',
    icon: GitPullRequestIcon,
    renderForm: BranchReviewRuleActionFormControl,
    initialValue: {
        type: 'branchReview',
        title: 'THREAD_TITLE TODO!(sqs)',
        branch: 'TODO!(sqs)',
        description:
            // eslint-disable-next-line no-template-curly-in-string
            'This change was automatically created as part of [Sourcegraph campaign ${campaign_number}](${campaign_url}).\n\nOther PRs created by this campaign:\n\n${related_links}',
    },
}
