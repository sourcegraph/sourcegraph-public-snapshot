import { useCallback } from 'react'

import classNames from 'classnames'

import { mdiMagnifyPlus } from '@mdi/js'
import { useQuery } from '@sourcegraph/http-client'
import { Button, Icon, Popover,PopoverContent, PopoverTail, PopoverTrigger,Position, TextArea, Tooltip } from '@sourcegraph/wildcard'

import { ALL_EVENT_NAMES } from './queries'
import styles from './userNode.module.scss'
import { AllEventNamesResult } from '../../../graphql-operations'


interface eventNamesInputButton {
    display: string
    value: string
    description?: string
    onClick?: (value: string) => void
}

const buttons: eventNamesInputButton[] = [
    {
        display: "Searches",
        value: "SearchResultsQueried",
        description: "Count of search queries run by a user in the Sourcegraph web UI.",
    },
    {
        display: "Code navigation actions",
        value: "findRefs,hover,goToDef",
        description: "Count of code navigation actions like 'Find References', 'Go to Definition', and hovers run by a user in the Sourcegraph web UI.",
    },
    {
        display: "Cody autocomplete suggestions",
        value: "CodyVSCodeExtension:completions:suggested,CodtyJetBrainsExtension:completions:suggested",
        description: "Count of Cody autocompletions suggested to the user in editor extensions.",
    },
    {
        display: "Cody autocomplete acceptances",
        value: "CodyVSCodeExtension:completions:accepted,CodyJetBrainsExtension:completions:accepted",
        description: "Count of Cody autocompletions accepted by the user in editor extensions.",
    },
]

interface Props {
    onChange: (value: string) => void
    value: string
    label: string
}

export const EventNamesInputSection = ({
    label,
    value,
    onChange,
}: Props): JSX.Element => {
    const allEventNamesResults = useQuery<AllEventNamesResult>(ALL_EVENT_NAMES, {})

    return (<div className="container">
        <h3>
            {label}
        </h3>
        <div className="row d-flex">
            <TextArea
                id="event-names-input"
                className={classNames('flex-grow-1 m-0', styles.textarea)}
                value={value}
                onChange={useCallback(e => onChange(e.target.value), [])}
            />
            <AddEventNamePopover eventNames={allEventNamesResults.data?.site.analytics.allEventNames || []} onSelect={(e: string) => onChange(value.length > 0 ? value + ", " + e : e)} />
        </div>
        <div className="row mt-2">
            {
                buttons.map(b => {
                    return (
                        <Tooltip key={b.value} content={b.description} placement="top">
                            <Button
                                key={b.value}
                                value={b.value}
                                onClick={() => {
                                    onChange(b.value)
                                    return b.onClick ? b.onClick(b.value) : undefined
                                }}
                                outline={true}
                                variant="secondary"
                                display="inline"
                                size="sm"        
                                className='mr-1'
                            >
                                {b.display}
                            </Button>
                        </Tooltip>
                    )
                })
            }
        </div>
    </div>
)}

interface PopoverProps {
    eventNames: (string | null)[]
    onSelect: (e: string) => void
}

const AddEventNamePopover = ({eventNames, onSelect}: PopoverProps): JSX.Element => {
    return (<Popover>
        <PopoverTrigger
            as={Button}
            variant="primary"
            aria-label="Add an event"
        >
            <Icon svgPath={mdiMagnifyPlus} size={16} />
        </PopoverTrigger>

        <PopoverContent position={Position.bottom} className="d-flex flex-column">
            {
                eventNames.map(e => (
                    e ? <Button key={e} onClick={_ => { onSelect(e) }}>{e}</Button> : undefined
                ))
            }
        </PopoverContent>
        <PopoverTail size="sm" />
    </Popover>
    )
}
