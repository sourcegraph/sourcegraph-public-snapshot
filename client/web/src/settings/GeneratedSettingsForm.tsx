import React from 'react'

import classNames from 'classnames'

import { ErrorLike } from '@sourcegraph/common'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { Checkbox, H3, Input, Label, Select, Text } from '@sourcegraph/wildcard'

import { SettingsSchema } from './SettingsFile'

import styles from './GeneratedSettingsForm.module.scss'

interface GeneratedSettingsFormProps {
    jsonSchema: SettingsSchema
    currentSettings?: Settings | ErrorLike | null
}

interface settingsNode {
    title: string
    description: string
    type: 'array' | 'boolean' | 'integer' | 'null' | 'number' | 'object' | 'string'
    enum?: string[]
    properties: Record<string, settingsNode>
    minimum?: number
    maximum?: number
}

const groupLevelClasses = [styles.groupLevel1, styles.groupLevel2, styles.groupLevel3]

export const GeneratedSettingsForm: React.FunctionComponent<GeneratedSettingsFormProps> = ({
    jsonSchema,
    currentSettings,
}) => {
    if (!currentSettings) {
        return <div>Settings not loaded yet</div>
    }
    if ((currentSettings as ErrorLike).message) {
        return <div>Can't load settings: {(currentSettings as ErrorLike).message}</div>
    }

    return <section className={classNames(styles.wrapper, 'mt-3')}>
        {convertPropertiesToComponents(jsonSchema as unknown as settingsNode, [], currentSettings as Settings)}
    </section>
}

function convertPropertiesToComponents(
    node: settingsNode,
    parentNames: string[],
    currentSettings: Settings
): JSX.Element[] {
    return Object.entries(node.properties).map(([name, subNode]) => {
        const id = parentNames.concat(name).join('.')
        switch (subNode.type) {
            case 'boolean':
                return (
                    <BooleanSettingItem
                        key={id}
                        id={id}
                        name={name}
                        title={subNode.title}
                        description={subNode.description}
                        value={currentSettings[name] as boolean}
                    />
                )
            case 'integer':
            case 'number':
                return (
                    <NumberSettingItem
                        key={id}
                        id={id}
                        name={name}
                        title={subNode.title}
                        description={subNode.description}
                        minimum={subNode?.minimum}
                        maximum={subNode?.maximum}
                        value={currentSettings[name] as number}
                    />
                )
            case 'string':
                return subNode.enum?.length ? (
                    <EnumItem
                        key={id}
                        id={id}
                        name={name}
                        enum={subNode.enum}
                        title={subNode.title}
                        description={subNode.description}
                        value={currentSettings[name] as string}
                    />
                ) : (
                    <StringSettingItem
                        key={id}
                        id={id}
                        name={name}
                        title={subNode.title}
                        description={subNode.description}
                        value={currentSettings[name] as string}
                    />
                )
            case 'object':
                if (subNode.properties) {
                    return (
                        <div key={id}>
                            <H3>{subNode.title}</H3>
                            <Text>{subNode.description}</Text>
                            <div className={groupLevelClasses[parentNames.length]}>
                                {convertPropertiesToComponents(
                                    subNode,
                                    [...parentNames, name],
                                    currentSettings[name] as Settings
                                )}
                            </div>
                        </div>
                    )
                }
                // TODO: Handle subNode.additionalProperties = true > allow obj properties to come in
                return <div key={id}>Unsupported object setting type</div>
            case 'array':
                // TODO: Handle array types > Operate similar to subNode.additionalProperties
                return <div key={id}>Unsupported array type</div>
            default:
                return <div key={id}>Unsupported setting type</div>
        }
    })
}

function BooleanSettingItem(props: {
    id: string
    name: string
    title: string
    description: string
    value: boolean
}): JSX.Element {
    return (
        <>
            <Checkbox id={props.id} label={props.name} checked={props.value} onChange={() => {}} />
            <Text className="text-muted">{props.description}</Text>
        </>
    )
}

function StringSettingItem(props: {
    id: string
    name: string
    title: string
    description: string
    value: string
}): JSX.Element {
    return (
        <>
            <Label htmlFor={props.id}>{props.name}</Label>
            <Input id={props.id} type="text" value={props.value} />
            <Text className="text-muted">{props.description}</Text>
        </>
    )
}

function NumberSettingItem(props: {
    id: string
    name: string
    title: string
    description: string
    minimum?: number
    maximum?: number
    value: number
}): JSX.Element {
    return (
        <>
            <Label htmlFor={props.id}>{props.name}</Label>
            <Input id={props.id} type="number" min={props.minimum ?? 0} max={props.maximum ?? ''} value={props.value} />
            <Text className="text-muted">{props.description}</Text>
        </>
    )
}

function EnumItem(props: {
    id: string
    name: string
    title: string
    description: string
    enum: string[]
    value: string
}): JSX.Element {
    return (
        <>
            <Select id={props.id} name={props.name} label={props.name} onChange={() => {}} className="mb-0">
                {props.enum.map(value => (
                    <option key={value} value={value} label={value} selected={props.value === value} />
                ))}
            </Select>
            <Text className="text-muted">{props.description}</Text>
        </>
    )
}
