import CloseIcon from 'mdi-react/CloseIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import React, { useCallback, useState } from 'react'
import TextareaAutosize from 'react-textarea-autosize'
import { CampaignFormControl } from './CampaignForm'

interface Props extends CampaignFormControl {
    autoFocus?: boolean
    className?: string
}

/**
 * The common form fields for campaigns.
 */
export const CampaignFormCommonFields: React.FunctionComponent<Props> = ({
    value,
    onChange,
    disabled,
    autoFocus,
    className = '',
}) => {
    const onNameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => onChange({ ...value, name: e.currentTarget.value }),
        [onChange, value]
    )
    const onBodyChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(
        e => onChange({ ...value, body: e.currentTarget.value }),
        [onChange, value]
    )
    const [rawDueDate, setRawDueDate] = useState<string | null>(null)
    const onDueDateChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => {
            setRawDueDate(e.currentTarget.value)
            onChange({
                ...value,
                dueDate: e.currentTarget.valueAsDate ? e.currentTarget.valueAsDate.toISOString() : undefined,
            })
        },
        [onChange, value]
    )
    const onAddDueDateClick = useCallback(() => setRawDueDate(''), [])
    const onClearDueDateClick = useCallback(() => setRawDueDate(null), [])

    return (
        <div className={className}>
            <div className="form-group">
                <label htmlFor="campaign-form-common-fields__name">Campaign name</label>
                <input
                    type="text"
                    id="campaign-form-common-fields__name"
                    className="form-control"
                    required={true}
                    minLength={1}
                    placeholder="Campaign name"
                    value={value.name}
                    onChange={onNameChange}
                    autoFocus={autoFocus}
                    disabled={disabled}
                />
            </div>
            <div className="form-group mb-1">
                <label htmlFor="campaign-form-common-fields__body">Campaign description</label>
                <TextareaAutosize
                    type="text"
                    id="campaign-form-common-fields__body"
                    className="form-control"
                    placeholder="Describe the purpose of this campaign, link to relevant internal documentation, etc."
                    value={value.body || ''}
                    minRows={3}
                    onChange={onBodyChange}
                    disabled={disabled}
                />
            </div>
            <div className="form-group">
                {rawDueDate === null ? (
                    <button type="button" className="btn btn-link pl-0 pr-1" onClick={onAddDueDateClick}>
                        <PlusIcon className="icon-inline" /> Due date
                    </button>
                ) : (
                    <>
                        <label htmlFor="campaign-form-common-fields__dueDate">Campaign due date</label>
                        <div className="input-group">
                            <input
                                type="date"
                                id="campaign-form-common-fields__dueDate"
                                className="form-control w-auto flex-0"
                                min="2019-08-11"
                                value={rawDueDate}
                                onChange={onDueDateChange}
                                disabled={disabled}
                            />
                            <div className="input-group-append">
                                <button type="button" className="btn btn-link" onClick={onClearDueDateClick}>
                                    <CloseIcon className="icon-inline" /> Remove due date
                                </button>
                            </div>
                        </div>
                    </>
                )}
            </div>
        </div>
    )
}
