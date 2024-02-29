import React, { useEffect, useMemo, useState } from 'react'

import classNames from 'classnames'

import { ErrorLike, isErrorLike, pluralize } from '@sourcegraph/common'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { Checkbox, H3, Input, Label, Select, Text } from '@sourcegraph/wildcard'

import { SaveToolbar } from '../components/SaveToolbar'
import { eventLogger } from '../tracking/eventLogger'

import styles from './GeneratedSettingsForm.module.scss'

interface Change<T> {
    parentsAndName: string[]
    newValue: T
    isValid: boolean
}

interface GeneratedSettingsFormProps {
    jsonSchema: SettingsNode
    currentSettings?: Settings | ErrorLike | null
    reportDirtiness: (dirty: boolean) => void
}

export interface SettingsNode {
    title: string
    description: string
    type: 'array' | 'boolean' | 'integer' | 'null' | 'number' | 'object' | 'string'
    enum?: string[]
    properties: Record<string, SettingsNode>
    minimum?: number
    maximum?: number
}

const groupLevelClasses = [styles.groupLevel1, styles.groupLevel2, styles.groupLevel3]

export const GeneratedSettingsForm: React.FunctionComponent<GeneratedSettingsFormProps> = ({
    jsonSchema,
    currentSettings,
    reportDirtiness,
}) => {
    const [saving, setSaving] = useState<boolean>(false)
    const [changes, setChanges] = useState<{ [key: string]: Change<boolean | number | string> }>({})
    const invalidCount = useMemo(() => Object.values(changes).filter(change => !change.isValid).length, [changes])
    const [resetTrigger, setResetTrigger] = useState<number>(0)

    const onChange = (change: Change<boolean | number | string>): void => {
        if (isErrorLike(currentSettings) || currentSettings === null) {
            return
        }

        const oldValue = getSetting(
            currentSettings as { [key: string]: boolean | number | string | object },
            change.parentsAndName
        )
        if (oldValue === change.newValue) {
            setChanges(changes => {
                const newChanges = { ...changes }
                delete newChanges[change.parentsAndName.join('.')]
                return newChanges
            })
        } else {
            setChanges(changes => ({ ...changes, [change.parentsAndName.join('.')]: change }))
        }
    }

    useEffect(() => {
        reportDirtiness(Object.values(changes).length > 0)
    }, [changes, reportDirtiness])

    if (!currentSettings) {
        return <div>Settings not loaded yet</div>
    }
    if ((currentSettings as ErrorLike).message) {
        return <div>Can't load settings: {(currentSettings as ErrorLike).message}</div>
    }

    const elements = convertPropertiesToFields(
        jsonSchema as unknown as SettingsNode,
        [],
        currentSettings as Settings,
        onChange,
        resetTrigger
    )
    return (
        <>
            <section className={classNames(styles.wrapper, 'mt-3')}>
                {/* TODO: Remove this H3 and ul */}
                <H3>Changes: (only here while WIP)</H3>
                <ul>
                    {Object.values(changes).map(change => (
                        <li key={change.parentsAndName.join('.')}>
                            {change.parentsAndName.join('.')} = {change.newValue.toString()} (
                            {change.isValid ? 'valid' : 'invalid'})
                        </li>
                    ))}
                </ul>
                {elements}
            </section>
            <SaveToolbar
                dirty={Object.values(changes).length > 0}
                error={
                    invalidCount > 0
                        ? new Error(`${pluralize('setting', invalidCount)} are invalid, see the warnings above`)
                        : undefined
                }
                saving={saving}
                onSave={() => {
                    eventLogger.log('SettingsFileSaved')
                    setSaving(true)

                    // TODO: Save
                    // Then: reportDirtiness(false)
                }}
                onDiscard={() => {
                    if (Object.values(changes).length === 0 || window.confirm('Discard settings edits?')) {
                        setResetTrigger(resetTrigger + 1)
                        setChanges({})
                        eventLogger.log('SettingsFileDiscard')
                        // TODO: Call this.props.onDidDiscard() of parent
                    } else {
                        eventLogger.log('SettingsFileDiscardCanceled')
                    }
                }}
            />
        </>
    )
}

function convertPropertiesToFields(
    node: SettingsNode,
    parentNames: string[],
    currentSettings: Settings,
    onChange: (change: Change<boolean | number | string>) => void,
    resetTrigger: number
): JSX.Element[] {
    return Object.entries(node.properties).map(([name, subNode]) => {
        const parentsAndName = parentNames.concat(name)
        const id = parentsAndName.join('.')
        switch (subNode.type) {
            case 'boolean':
                return (
                    <BooleanField
                        key={id}
                        parentsAndName={parentsAndName}
                        title={subNode.title}
                        description={subNode.description}
                        initialValue={(currentSettings[name] as boolean) ?? false}
                        onChange={onChange}
                        resetTrigger={resetTrigger}
                    />
                )
            case 'integer':
            case 'number':
                return (
                    <NumberField
                        key={id}
                        parentsAndName={parentsAndName}
                        title={subNode.title}
                        description={subNode.description}
                        minimum={subNode?.minimum}
                        maximum={subNode?.maximum}
                        initialValue={(currentSettings[name] as number) ?? 0}
                        onChange={onChange}
                        resetTrigger={resetTrigger}
                    />
                )
            case 'string':
                return subNode.enum?.length ? (
                    <EnumField
                        key={id}
                        parentsAndName={parentsAndName}
                        enum={subNode.enum}
                        title={subNode.title}
                        description={subNode.description}
                        initialValue={(currentSettings[name] as string) ?? subNode.enum[0]}
                        onChange={onChange}
                        resetTrigger={resetTrigger}
                    />
                ) : (
                    <StringField
                        key={id}
                        parentsAndName={parentsAndName}
                        title={subNode.title}
                        description={subNode.description}
                        initialValue={(currentSettings[name] as string) ?? ''}
                        onChange={onChange}
                        resetTrigger={resetTrigger}
                    />
                )
            case 'object':
                if (subNode.properties) {
                    return (
                        <div key={id}>
                            <H3>{subNode.title}</H3>
                            <Text>{subNode.description}</Text>
                            <div className={groupLevelClasses[parentNames.length]}>
                                {convertPropertiesToFields(
                                    subNode,
                                    [...parentNames, name],
                                    (currentSettings[name] as Settings) || {},
                                    onChange,
                                    resetTrigger
                                )}
                            </div>
                        </div>
                    )
                }
                // TODO: Handle subNode.additionalProperties = true → allow obj properties to come in
                return <div key={id}>Unsupported object setting type</div>
            case 'array':
                // TODO: Handle array types – Operate similar to subNode.additionalProperties
                return <div key={id}>Unsupported array type</div>
            default:
                return <div key={id}>Unsupported setting type</div>
        }
    })
}

function BooleanField(props: {
    parentsAndName: string[]
    title: string
    description: string
    initialValue: boolean
    onChange: (change: Change<boolean>) => void
    resetTrigger: number
}): JSX.Element {
    const id = props.parentsAndName.join('.')
    const name = props.parentsAndName[props.parentsAndName.length - 1]
    const [value, setValue] = useState<boolean>(props.initialValue)
    useEffect(() => {
        setValue(props.initialValue)
    }, [props.initialValue, props.resetTrigger])
    return (
        <>
            <Checkbox
                id={id}
                label={name}
                checked={value}
                onChange={event => {
                    setValue(event.target.checked)
                    props.onChange({
                        parentsAndName: props.parentsAndName,
                        newValue: event.target.checked,
                        isValid: true,
                    })
                }}
            />
            <Text className="text-muted">{props.description}</Text>
        </>
    )
}

function StringField(props: {
    parentsAndName: string[]
    title: string
    description: string
    initialValue: string
    onChange: (change: Change<string>) => void
    resetTrigger: number
}): JSX.Element {
    const id = props.parentsAndName.join('.')
    const title = props.title ?? props.parentsAndName[props.parentsAndName.length - 1]
    const [value, setValue] = useState<string>(props.initialValue)
    useEffect(() => {
        setValue(props.initialValue)
    }, [props.initialValue, props.resetTrigger])
    return (
        <>
            <Label htmlFor={id} className="mb-0">
                {title}
            </Label>
            <Text className="text-muted mb-0">{props.description}</Text>
            <Input
                type="text"
                id={id}
                className="mb-2"
                value={value}
                onChange={event => {
                    setValue(event.target.value)
                    props.onChange({
                        parentsAndName: props.parentsAndName,
                        newValue: event.target.value,
                        isValid: true,
                    })
                }}
            />
            <Text className="text-muted">{props.description}</Text>
        </>
    )
}

function NumberField(props: {
    parentsAndName: string[]
    title: string
    description: string
    minimum?: number
    maximum?: number
    initialValue: number
    onChange: (change: Change<number>) => void
    resetTrigger: number
}): JSX.Element {
    const id = props.parentsAndName.join('.')
    const title = props.title ?? props.parentsAndName[props.parentsAndName.length - 1]
    const [value, setValue] = useState<number>(props.initialValue)
    useEffect(() => {
        setValue(props.initialValue)
    }, [props.initialValue, props.resetTrigger])
    return (
        <>
            <Label htmlFor={id}>{title}</Label>
            <Input
                id={id}
                type="number"
                min={props.minimum ?? 0}
                max={props.maximum ?? ''}
                value={value}
                onChange={event => {
                    setValue(parseFloat(event.target.value))
                    props.onChange({
                        parentsAndName: props.parentsAndName,
                        newValue: parseFloat(event.target.value),
                        isValid: true,
                    })
                }}
            />
            <Text className="text-muted">{props.description}</Text>
        </>
    )
}

function EnumField(props: {
    parentsAndName: string[]
    title: string
    description: string
    enum: string[]
    initialValue: string
    onChange: (change: Change<string>) => void
    resetTrigger: number
}): JSX.Element {
    const id = props.parentsAndName.join('.')
    const title = props.title ?? props.parentsAndName[props.parentsAndName.length - 1]
    const [value, setValue] = useState<string>(props.initialValue)
    useEffect(() => {
        setValue(props.initialValue)
    }, [props.initialValue, props.resetTrigger])
    return (
        <>
            <Text className="text-muted mb-0">{props.description}</Text>
            <Select
                id={id}
                label={title}
                value={value}
                onChange={event => {
                    setValue(event.target.value)
                    props.onChange({
                        parentsAndName: props.parentsAndName,
                        newValue: event.target.value,
                        isValid: true,
                    })
                }}
                selectClassName="mb-2"
            >
                {props.enum.map(enumValue => (
                    <option key={enumValue} value={enumValue} label={enumValue} />
                ))}
            </Select>
            <Text className="text-muted">{props.description}</Text>
        </>
    )
}

function getSetting(
    settings: { [key: string]: boolean | number | string | object },
    parentsAndName: string[]
): boolean | number | string | undefined {
    let value: any = settings
    for (const name of parentsAndName) {
        if (typeof value === 'object' && value[name] !== undefined) {
            value = value[name]
        } else {
            return undefined
        }
    }
    return value as boolean | number | string
}
