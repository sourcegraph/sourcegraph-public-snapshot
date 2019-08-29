import CloseIcon from 'mdi-react/CloseIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import React, { useCallback, useState } from 'react'
import TextareaAutosize from 'react-textarea-autosize'
import { CampaignFormControl } from './CampaignForm'
import { format, startOfDay, addDays, parse, addHours } from 'date-fns'

interface Props extends CampaignFormControl {
    autoFocus?: boolean
    className?: string
}

const DATE_FORMAT = 'yyyy-MM-dd'
const DATETIME_FORMAT = "yyyy-MM-dd'T'HH:mm"

const parseInputDateTimeLocal = (value: string | null): string | undefined => {
    if (!value) {
        return undefined
    }
    try {
        return parse(value, DATETIME_FORMAT, new Date()).toISOString()
    } catch (err) {
        console.error('Error parsing <input type="datetime-local"> value:', err)
        return undefined
    }
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

    const [rawStartDate, setRawStartDate] = useState<string | null>(null)
    const setStartDate = useCallback(
        (startDate: string | null) => {
            setRawStartDate(startDate)
            onChange({ ...value, startDate: parseInputDateTimeLocal(startDate) })
        },
        [onChange, value]
    )
    const onStartDateChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => setStartDate(e.currentTarget.value),
        [setStartDate]
    )
    const onAddStartDateClick = useCallback(() => setStartDate(format(addHours(new Date(), 1), DATETIME_FORMAT)), [
        setStartDate,
    ])
    const onClearStartDateClick = useCallback(() => setStartDate(null), [setStartDate])

    const [rawDueDate, setRawDueDate] = useState<string | null>(null)
    const setDueDate = useCallback(
        (dueDate: string | null) => {
            setRawDueDate(dueDate)
            onChange({ ...value, dueDate: parseInputDateTimeLocal(dueDate) })
        },
        [onChange, value]
    )
    const onDueDateChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => setDueDate(e.currentTarget.value),
        [setDueDate]
    )
    const onAddDueDateClick = useCallback(() => setDueDate(format(addDays(new Date(), 7), DATETIME_FORMAT)), [
        setDueDate,
    ])
    const onClearDueDateClick = useCallback(() => setDueDate(null), [setDueDate])

    const minStartDate = format(startOfDay(addDays(new Date(), 1)), DATE_FORMAT)
    const minDueDate = value.startDate ? format(addDays(new Date(value.startDate), 1), DATE_FORMAT) : undefined

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
            <div className="form-group">
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
            <div className="form-group mb-2">
                {rawStartDate === null ? (
                    <button
                        type="button"
                        className="btn btn-link text-decoration-none pl-0 pr-1"
                        onClick={onAddStartDateClick}
                    >
                        <PlusIcon className="icon-inline" /> Start date
                    </button>
                ) : (
                    <>
                        <label htmlFor="campaign-form-common-fields__startDate">Campaign start date</label>
                        <div className="input-group">
                            <input
                                type="datetime-local"
                                id="campaign-form-common-fields__startDate"
                                className="form-control w-auto flex-0"
                                required={true}
                                min={minStartDate}
                                value={rawStartDate}
                                onChange={onStartDateChange}
                                disabled={disabled}
                            />
                            <div className="input-group-append">
                                <button
                                    type="button"
                                    className="btn btn-link text-decoration-none"
                                    onClick={onClearStartDateClick}
                                >
                                    <CloseIcon className="icon-inline" /> Remove start date
                                </button>
                            </div>
                        </div>
                    </>
                )}
            </div>
            <div className="form-group">
                {rawDueDate === null ? (
                    <button
                        type="button"
                        className="btn btn-link text-decoration-none pl-0 pr-1"
                        onClick={onAddDueDateClick}
                    >
                        <PlusIcon className="icon-inline" /> Due date
                    </button>
                ) : (
                    <>
                        <label htmlFor="campaign-form-common-fields__dueDate">Campaign due date</label>
                        <div className="input-group">
                            <input
                                type="datetime-local"
                                id="campaign-form-common-fields__dueDate"
                                className="form-control w-auto flex-0"
                                required={true}
                                min={minDueDate}
                                value={rawDueDate}
                                onChange={onDueDateChange}
                                disabled={disabled}
                            />
                            <div className="input-group-append">
                                <button
                                    type="button"
                                    className="btn btn-link text-decoration-none"
                                    onClick={onClearDueDateClick}
                                >
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
