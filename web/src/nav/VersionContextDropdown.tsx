import * as H from 'history'
import React from 'react'
import {
    ListboxOption,
    ListboxInput,
    ListboxButton,
    ListboxPopover,
    ListboxList,
    ListboxGroupLabel,
} from '@reach/listbox'
import classNames from 'classnames'
import { VersionContextProps } from '../../../shared/src/search/util'
import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'
import FlagVariantIcon from 'mdi-react/FlagVariantIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import { VersionContext } from '../schema/site.schema'
import { PatternTypeProps, CaseSensitivityProps, InteractiveSearchProps } from '../search'
import { submitSearch } from '../search/helpers'
import { isEmpty } from 'lodash'
import { useLocalStorage } from '../util/useLocalStorage'

const HAS_DISMISSED_INFO_KEY = 'sg-has-dismissed-version-context-info'

export interface VersionContextDropdownProps
    extends Pick<PatternTypeProps, 'patternType'>,
        Pick<CaseSensitivityProps, 'caseSensitive'>,
        Partial<Pick<InteractiveSearchProps, 'filtersInQuery'>>,
        VersionContextProps {
    setVersionContext: (versionContext: string | undefined) => void
    availableVersionContexts: VersionContext[] | undefined
    history: H.History
    navbarSearchQuery: string
}

export const VersionContextDropdown: React.FunctionComponent<VersionContextDropdownProps> = (
    props: VersionContextDropdownProps
) => {
    const [hasDismissedInfo, setHasDismissedInfo] = useLocalStorage(HAS_DISMISSED_INFO_KEY, 'false')

    const submitOnToggle = (versionContext: string): void => {
        const { history, navbarSearchQuery, filtersInQuery, caseSensitive, patternType } = props
        const searchQueryNotEmpty = navbarSearchQuery !== '' || (filtersInQuery && !isEmpty(filtersInQuery))
        const activation = undefined
        const source = 'filter'
        if (searchQueryNotEmpty) {
            submitSearch({
                history,
                query: navbarSearchQuery,
                source,
                patternType,
                caseSensitive,
                versionContext,
                activation,
                filtersInQuery,
            })
        }
    }

    const updateValue = (newValue: string): void => {
        props.setVersionContext(newValue)
        submitOnToggle(newValue)
    }

    const disableValue = (): void => {
        props.setVersionContext(undefined)
    }

    if (!props.availableVersionContexts || props.availableVersionContexts.length === 0) {
        return null
    }

    const onDismissInfo = (e: React.MouseEvent<HTMLButtonElement>): void => {
        e.preventDefault()
        setHasDismissedInfo('true')
    }

    const showInfo = (e: React.MouseEvent<HTMLButtonElement>): void => {
        e.preventDefault()
        setHasDismissedInfo('false')
    }

    return (
        <>
            {props.availableVersionContexts ? (
                <div className="version-context-dropdown">
                    <ListboxInput value={props.versionContext} onChange={updateValue}>
                        {({ isExpanded }) => (
                            <>
                                <ListboxButton className="version-context-dropdown__button btn btn-secondary">
                                    <FlagVariantIcon className="icon-inline small" />
                                    <span className="version-context-dropdown__button-text ml-2 mr-1">
                                        {!props.versionContext || props.versionContext === 'default'
                                            ? 'Select context'
                                            : `${props.versionContext} (Active)`}
                                    </span>
                                    <MenuDownIcon className="icon-inline" />
                                </ListboxButton>
                                <ListboxPopover
                                    className={classNames('version-context-dropdown__popover dropdown-menu', {
                                        show: isExpanded,
                                    })}
                                >
                                    <div className="version-context-dropdown__title pl-2 mb-1">
                                        <span>Select version context</span>
                                        <button type="button" className="btn btn-icon" onClick={showInfo}>
                                            <HelpCircleOutlineIcon className="icon-inline small" />
                                        </button>
                                    </div>
                                    {hasDismissedInfo !== 'true' && (
                                        <div className="version-context-dropdown__info card">
                                            <span className="font-weight-bold">About version contexts</span>
                                            <p className="mb-2">
                                                Version contexts (
                                                <a href="http://docs.sourcegraph.com/user/search#version-contexts">
                                                    documentation
                                                </a>
                                                ) allow you to search a set of repositories based on a commit hash, tag,
                                                or other interesting moment in time of multiple code bases. Your
                                                administrator can configure version contexts in the site configuration.
                                            </p>
                                            <button
                                                type="button"
                                                className="btn btn-outline-primary version-context-dropdown__info-dismiss"
                                                onClick={onDismissInfo}
                                            >
                                                Do not show this again
                                            </button>
                                        </div>
                                    )}
                                    <ListboxList className="version-context-dropdown__list">
                                        <ListboxGroupLabel
                                            disabled={true}
                                            value="title"
                                            className="version-context-dropdown__option version-context-dropdown__title"
                                        >
                                            <VersionContextInfoRow
                                                name="Name"
                                                description="Description"
                                                isActive={false}
                                                onDisableValue={disableValue}
                                            />
                                        </ListboxGroupLabel>
                                        {props.availableVersionContexts
                                            // Render the current version context at the top, then other available version
                                            // contexts in alphabetical order.
                                            ?.sort((a, b) => {
                                                if (a.name === props.versionContext) {
                                                    return -1
                                                }
                                                if (b.name === props.versionContext) {
                                                    return 1
                                                }
                                                return a.name > b.name ? 1 : -1
                                            })
                                            .map(versionContext => (
                                                <ListboxOption
                                                    key={versionContext.name}
                                                    value={versionContext.name}
                                                    label={versionContext.name}
                                                    className="version-context-dropdown__option"
                                                >
                                                    <VersionContextInfoRow
                                                        name={versionContext.name}
                                                        description={versionContext.description || ''}
                                                        isActive={props.versionContext === versionContext.name}
                                                        onDisableValue={disableValue}
                                                    />
                                                </ListboxOption>
                                            ))}
                                    </ListboxList>
                                </ListboxPopover>
                            </>
                        )}
                    </ListboxInput>
                </div>
            ) : null}
        </>
    )
}

const VersionContextInfoRow: React.FunctionComponent<{
    name: string
    description: string
    isActive: boolean
    onDisableValue: () => void
}> = ({ name, description, isActive, onDisableValue }) => (
    <>
        <div>
            {isActive && (
                <button
                    type="button"
                    className="btn btn-icon"
                    onClick={onDisableValue}
                    aria-label="Disable version context"
                >
                    <CloseIcon className="icon-inline small" />
                </button>
            )}
        </div>
        <span className="version-context-dropdown__option-name">{name}</span>
        <span className="version-context-dropdown__option-description">{description}</span>
    </>
)
