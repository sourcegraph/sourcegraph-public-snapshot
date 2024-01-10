import { type FC, memo, useMemo, useState, type MouseEvent } from 'react'

import { json } from '@codemirror/lang-json'
import { EditorState } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { mdiChevronDown, mdiChevronRight, mdiClose } from '@mdi/js'
import classNames from 'classnames'

import { gql, useQuery } from '@sourcegraph/http-client'
import {
    CodeMirrorEditor,
    defaultEditorTheme,
    jsonHighlighting,
} from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import {
    setTemporarySettingOverride,
    removeTemporarySettingOverride,
    getTemporarySettingOverride,
} from '@sourcegraph/shared/src/settings/temporary/localOverride'
import {
    TEMPORARY_SETTINGS_KEYS,
    type TemporarySettings,
} from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import {
    Alert,
    Button,
    Code,
    H3,
    H4,
    Icon,
    Input,
    Label,
    LoadingSpinner,
    Modal,
    Select,
    Tab,
    TabList,
    TabPanel,
    TabPanels,
    Tabs,
    Text,
    Tooltip,
    Badge,
    Container,
} from '@sourcegraph/wildcard'

import { FEATURE_FLAGS, type FeatureFlagName } from '../featureFlags/featureFlags'
import {
    getFeatureFlagOverride,
    removeFeatureFlagOverride,
    setFeatureFlagOverride,
} from '../featureFlags/lib/feature-flag-local-overrides'
import type { DeveloperSettingsEvaluatedFeatureFlagsResult } from '../graphql-operations'
import {
    setDeveloperSettingsFeatureFlags,
    setDeveloperSettingsTemporarySettings,
    setDeveloperSettingsSearchOptions,
    toggleDevSettingsDialog,
    updateOverrideCounter,
    useDeveloperSettings,
    useOverrideCounter,
} from '../stores'

import { ReloadButton } from './DeveloperSettingsGlobalNavItem'
import { EventLoggingDebugToggle } from './settings/eventLoggingDebug'

import styles from './DeveloperDialog.module.scss'

export const DeveloperDialog: FC<{}> = () => {
    const counter = useOverrideCounter()
    const index = useDeveloperSettings(settings => settings.selectedTab)

    return (
        <Modal
            aria-label="Development settings"
            className={styles.dialog}
            onDismiss={() => toggleDevSettingsDialog(false)}
        >
            <H3>Developer Settings</H3>
            <Text>
                You can temporarily override settings here for development purposes. Any changes will be stored locally
                in your browser only.
            </Text>
            <Tabs
                lazy={true}
                behavior="memoize"
                size="medium"
                className="overflow-hidden d-flex flex-column"
                index={index}
                onChange={index => useDeveloperSettings.setState({ selectedTab: index })}
            >
                <TabList>
                    <Tab>
                        Feature flags{' '}
                        <Badge pill={true}>
                            {counter.featureFlags}/{FEATURE_FLAGS.length}
                        </Badge>
                    </Tab>
                    <Tab>
                        Temporary settings{' '}
                        <Badge pill={true}>
                            {counter.temporarySettings}/{TEMPORARY_SETTINGS_KEYS.length}
                        </Badge>
                    </Tab>
                    <Tab>Misc</Tab>
                    <Tab>Zoekt</Tab>
                </TabList>
                <TabPanels className="overflow-hidden flex-1 min-w-0 d-flex">
                    <TabPanel className={styles.content}>
                        <FeatureFlags />
                    </TabPanel>
                    <TabPanel className={styles.content}>
                        <TemporarySettingsPanel />
                    </TabPanel>
                    <TabPanel className={styles.content}>
                        <ul className={styles.settingsList}>
                            <li>
                                <EventLoggingDebugToggle />
                            </li>
                        </ul>
                    </TabPanel>
                    <TabPanel className={styles.content}>
                        <ZoektSettings />
                    </TabPanel>
                </TabPanels>
            </Tabs>
        </Modal>
    )
}

const EVALUATED_FEATURE_FLAGS = gql`
    query DeveloperSettingsEvaluatedFeatureFlags {
        evaluatedFeatureFlags {
            name
            value
        }
    }
`

const FeatureFlags: FC<{}> = () => {
    const { view, filter } = useDeveloperSettings(settings => settings.featureFlags)
    const { data, loading } = useQuery<DeveloperSettingsEvaluatedFeatureFlagsResult>(EVALUATED_FEATURE_FLAGS, {
        fetchPolicy: 'cache-first',
    })

    const evaluatedFeatureFlags = useMemo(() => {
        const flags = new Map<string, boolean>()
        for (const { name, value } of data?.evaluatedFeatureFlags ?? []) {
            flags.set(name, value)
        }
        return flags
    }, [data])

    if (loading) {
        return (
            <div className="d-flex mt-3 align-items-center">
                <LoadingSpinner />
                <span>Loading flags...</span>
            </div>
        )
    }

    const hasFilter = !!filter
    const featureFlags = hasFilter ? filterSmartCase(FEATURE_FLAGS, filter) : FEATURE_FLAGS.slice()
    const finalView = hasFilter ? 'All' : view

    return (
        <>
            <Alert variant="info" className="my-2">
                Click on the respective "Override value" entry to cycle through enabled, disabled and not set. If the
                feature flag is used on the server, reload the page via the reload button to apply them to the intial
                page load as well.
            </Alert>
            <div className="d-flex align-items-center my-2">
                <Label className="mb-0" htmlFor="feature-flag-view">
                    View: &nbsp;
                </Label>
                <Select
                    id="feature-flag-view"
                    className="mb-0"
                    aria-label="View"
                    value={finalView}
                    onChange={event => setDeveloperSettingsFeatureFlags({ view: event.target.value.trim() })}
                    disabled={hasFilter}
                >
                    <option>All</option>
                    <option>Enabled</option>
                    <option>Overridden</option>
                </Select>
                <Label className="ml-3 mb-0" htmlFor="feature-flag-filter">
                    Filter: &nbsp;
                </Label>
                <FilterInput
                    id="feature-flag-filter"
                    value={filter}
                    onChange={value => setDeveloperSettingsFeatureFlags({ filter: value })}
                    placeholder="Filter feature flags..."
                />
                <ReloadButton className="ml-3 flex-1" variant="primary">
                    Reload
                </ReloadButton>
            </div>
            <div className="flex-1 overflow-auto min-h-0 mt-2">
                <table>
                    <thead>
                        <tr>
                            <th>Name</th>
                            <th className="text-center">Override</th>
                            <th className="text-center">Actual</th>
                        </tr>
                    </thead>
                    <tbody>
                        {featureFlags.sort().map(featureFlag => (
                            <FeatureFlagOverride
                                key={featureFlag}
                                featureFlag={featureFlag}
                                filter={finalView}
                                serverValue={evaluatedFeatureFlags.get(featureFlag)}
                            />
                        ))}
                    </tbody>
                </table>
            </div>
        </>
    )
}

const FeatureFlagOverride: FC<{ featureFlag: FeatureFlagName; filter: string; serverValue?: boolean }> = memo(
    ({ featureFlag, filter, serverValue }) => {
        const [overrideValue, setOverrideValue] = useState(getFeatureFlagOverride(featureFlag))

        const enabled = overrideValue === true || (overrideValue === null && serverValue)
        const overridden = overrideValue !== null

        function onClick(): void {
            switch (overrideValue) {
                case null: {
                    setFeatureFlagOverride(featureFlag, true)
                    break
                }
                case true: {
                    setFeatureFlagOverride(featureFlag, false)
                    break
                }
                case false: {
                    removeFeatureFlagOverride(featureFlag)
                    break
                }
            }
            updateOverrideCounter()
            setOverrideValue(getFeatureFlagOverride(featureFlag))
        }

        if ((filter === 'Enabled' && !enabled) || (filter === 'Overridden' && !overridden)) {
            return null
        }

        return (
            <tr className={overridden ? styles.override : ''}>
                <td>{featureFlag}</td>
                <td className="text-center">
                    <Button variant="link" onClick={onClick}>
                        <Code>{overrideValue === null ? 'n/a' : String(overrideValue)}</Code>
                    </Button>
                </td>
                <td className="text-center">
                    <Code>{String(serverValue ?? 'n/a')}</Code>
                </td>
            </tr>
        )
    }
)

const TemporarySettingsPanel: FC<{}> = () => {
    const { view, filter } = useDeveloperSettings(settings => settings.temporarySettings)

    const hasFilter = !!filter
    const temporarySettings = hasFilter
        ? filterSmartCase(TEMPORARY_SETTINGS_KEYS, filter)
        : TEMPORARY_SETTINGS_KEYS.slice()
    const finalView = hasFilter ? 'All' : view

    return (
        <>
            <Alert variant="info" className="my-2">
                Because we cannot check the validity of a temporary settings value, this UI only allows you to intercept
                and reset a temporary setting. Due to the nature of the API only the current (overridden or actual)
                value is shown.
            </Alert>
            <div className="d-flex align-items-center my-2">
                <Label className="mb-0" htmlFor="temporary-settings-view">
                    View:&nbsp;
                </Label>
                <Select
                    id="temporary-settings-view"
                    className="mb-0"
                    aria-label="View"
                    value={finalView}
                    onChange={event => setDeveloperSettingsTemporarySettings({ view: event.target.value.trim() })}
                    disabled={hasFilter}
                >
                    <option>All</option>
                    <option>Overridden</option>
                </Select>
                <Label className="ml-3 mb-0" htmlFor="temporary-settings-filter">
                    Filter: &nbsp;
                </Label>
                <FilterInput
                    id="temporary-settings-filter"
                    value={filter}
                    onChange={value => setDeveloperSettingsTemporarySettings({ filter: value })}
                    placeholder="Filter temporary settings..."
                />
            </div>
            <div className="flex-1 overflow-auto min-h-0 mt-2">
                <table>
                    <thead>
                        <tr>
                            <th />
                            <th>Name</th>
                            <th className="text-center">Value</th>
                            <th className="text-center">Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {temporarySettings
                            .slice()
                            .sort()
                            .map(setting => (
                                <TemporarySettingOverride key={setting} setting={setting} filter={view} />
                            ))}
                    </tbody>
                </table>
            </div>
        </>
    )
}

const TemporarySettingOverride: FC<{ setting: keyof TemporarySettings; filter: string }> = memo(
    ({ setting, filter }) => {
        const [value] = useTemporarySetting(setting)
        const [overrideValue, setOverrideValue] = useState(getTemporarySettingOverride(setting))
        const [open, setOpen] = useState(false)

        const overridden = overrideValue !== null
        const isObject = typeof value === 'object' && value

        function toggleOpen(event: MouseEvent): void {
            if (isObject) {
                event.stopPropagation()
                setOpen(open => !open)
            }
        }

        function reset(event: MouseEvent): void {
            event.stopPropagation()
            setTemporarySettingOverride(setting, { value: undefined })
            setOverrideValue(getTemporarySettingOverride(setting))
            updateOverrideCounter()
        }

        function removeOverride(event: MouseEvent): void {
            event.stopPropagation()
            removeTemporarySettingOverride(setting)
            setOverrideValue(getTemporarySettingOverride(setting))
            updateOverrideCounter()
        }

        if (filter === 'Overridden' && !overridden) {
            return null
        }

        return (
            <>
                <tr className={classNames({ [styles.override]: overridden })} onClick={toggleOpen}>
                    <td>
                        {isObject && (
                            <Button className="d-inline-flex" variant="icon" onClick={toggleOpen}>
                                <Icon svgPath={open ? mdiChevronDown : mdiChevronRight} aria-hidden={true} />
                            </Button>
                        )}
                    </td>
                    <td>{setting}</td>
                    <td className="text-center">
                        <Code>{value === undefined ? 'n/a' : isObject ? '(object)' : JSON.stringify(value)}</Code>
                    </td>
                    <td className="text-center">
                        <Button
                            className="py-0"
                            variant="link"
                            onClick={reset}
                            disabled={!!overrideValue && overrideValue.value === undefined}
                        >
                            Clear
                        </Button>
                        <br />
                        <Tooltip content="Restore actual value (remove override)">
                            <Button className="py-0" variant="link" onClick={removeOverride} disabled={!overridden}>
                                Restore
                            </Button>
                        </Tooltip>
                    </td>
                </tr>
                {isObject && open && (
                    <tr>
                        <td colSpan={4}>
                            <JSONView value={value} />{' '}
                        </td>
                    </tr>
                )}
            </>
        )
    }
)

const ZoektSettings: FC<{}> = () => {
    const { searchOptions } = useDeveloperSettings(settings => settings.zoekt)

    const [inputValue, setInputValue] = useState<string>(searchOptions)

    const handleClick = (): void => {
        setDeveloperSettingsSearchOptions({ searchOptions: inputValue })
    }

    const isLightTheme = useIsLightTheme()
    const extensions = useMemo(
        () => [
            EditorView.darkTheme.of(!isLightTheme),
            json(),
            defaultEditorTheme,
            jsonHighlighting,
            EditorView.updateListener.of(update => {
                if (update.docChanged) {
                    setInputValue(update.state.sliceDoc())
                }
            }),
        ],
        [isLightTheme]
    )

    return (
        <div className="mt-2 d-flex flex-column">
            <H4>Search Options</H4>
            <Text>Enter a valid JSON below. Missing values are replaced with defaults.</Text>
            <Container className="p-1">
                <CodeMirrorEditor value={inputValue} extensions={extensions} />
            </Container>
            <Button variant="primary" className="mt-2 align-self-end" onClick={handleClick}>
                Apply
            </Button>
        </div>
    )
}

const JSONView: FC<{ value: unknown }> = ({ value }) => {
    const isLightTheme = useIsLightTheme()
    const jsonValue = useMemo(() => JSON.stringify(value, undefined, 4), [value])
    const extensions = useMemo(
        () => [
            EditorView.darkTheme.of(!isLightTheme),
            EditorState.readOnly.of(true),
            json(),
            defaultEditorTheme,
            jsonHighlighting,
        ],
        [isLightTheme]
    )

    return <CodeMirrorEditor value={jsonValue} extensions={extensions} />
}

const FilterInput: FC<{ id: string; value: string; placeholder: string; onChange: (value: string) => void }> = ({
    id,
    value,
    placeholder,
    onChange,
}) => (
    <div className="position-relative d-flex align-items-center">
        <Input
            id={id}
            className="mb-0"
            inputClassName="pr-4"
            value={value}
            placeholder={placeholder}
            onChange={event => onChange(event.target.value)}
        />
        {value && (
            <Button className={styles.clearButton} variant="icon" onClick={() => onChange('')}>
                <Icon svgPath={mdiClose} aria-hidden="true" />
            </Button>
        )}
    </div>
)

function filterSmartCase<T extends string>(items: readonly T[], filter: string): T[] {
    const caseSensitive = /[A-Z]/.test(filter)
    return items.filter(item => (caseSensitive ? item : item.toLowerCase()).includes(filter))
}
