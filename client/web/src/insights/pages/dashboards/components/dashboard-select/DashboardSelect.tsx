import {
    ListboxOption,
    ListboxInput,
    ListboxButton,
    ListboxPopover,
    ListboxList,
    ListboxGroup,
    ListboxGroupLabel,
} from '@reach/listbox'
import { VisuallyHidden } from '@reach/visually-hidden'
import classnames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import React, { useState } from 'react'

import styles from './DashboardSelect.module.scss'

const LABEL_ID = 'insights-dashboards--select'

export interface DashboardSelectProps {}

export const DashboardSelect: React.FunctionComponent = props => {
    const {} = props

    const [value, setValue] = useState()

    const handleChange = (value: string) => {}

    return (
        <div>
            <VisuallyHidden id={LABEL_ID}>Choose a dashboard</VisuallyHidden>

            <ListboxInput value={value} onChange={handleChange}>
                <ListboxButton className={styles.listboxButton}>
                    {({ value, label, isExpanded }) => (
                        <>
                            <span>{value} </span>

                            {isExpanded ? <ChevronUpIcon /> : <ChevronDownIcon />}
                        </>
                    )}
                </ListboxButton>

                <ListboxPopover className={classnames(styles.listboxPopover)} portal={true}>
                    <ListboxList className={classnames(styles.listboxList, 'dropdown-menu')}>
                        <ListboxOption className={styles.listboxOption} value="all">
                            All Insights
                        </ListboxOption>

                        <ListboxGroup>
                            <ListboxGroupLabel className={classnames(styles.listboxGroupLabel, 'text-muted')}>
                                Private
                            </ListboxGroupLabel>
                            <ListboxOption className={styles.listboxOption} value="vova">
                                Vova Kulikov Insights
                            </ListboxOption>
                            <ListboxOption className={styles.listboxOption} value="okr">
                                OKRs 2022
                            </ListboxOption>
                        </ListboxGroup>

                        <ListboxGroup>
                            <ListboxGroupLabel className={classnames(styles.listboxGroupLabel, 'text-muted')}>
                                Org 1
                            </ListboxGroupLabel>
                            <ListboxOption className={styles.listboxOption} value="org1">
                                Org 1 Insights
                            </ListboxOption>
                            <ListboxOption className={styles.listboxOption} value="migration">
                                Migrations
                            </ListboxOption>
                        </ListboxGroup>

                        <ListboxGroup>
                            <ListboxGroupLabel className={classnames(styles.listboxGroupLabel, 'text-muted')}>
                                Sourcegraph
                            </ListboxGroupLabel>
                            <ListboxOption className={styles.listboxOption} value="sg">
                                Sourcegraph Insights
                                <span className={classnames('badge badge-secondary', styles.listboxBadge)}>
                                    Sourcegraph
                                </span>
                            </ListboxOption>
                            <ListboxOption className={styles.listboxOption} value="code-search">
                                Code Search
                            </ListboxOption>
                            <ListboxOption className={styles.listboxOption} value="long">
                                <span className={styles.listboxOptionText}>
                                    Very looooooong insight dashboard name that doesn't fit
                                </span>
                                <span className={classnames('badge badge-secondary', styles.listboxBadge)}>
                                    Sourcegraph
                                </span>
                            </ListboxOption>
                            <ListboxOption className={styles.listboxOption} value="ext">
                                Extensibility
                            </ListboxOption>
                        </ListboxGroup>
                    </ListboxList>
                </ListboxPopover>
            </ListboxInput>
        </div>
    )
}
