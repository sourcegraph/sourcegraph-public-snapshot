import { FC, memo, useMemo, useState, MouseEvent } from "react"
import { Alert, Button, H3, Icon, LoadingSpinner, Modal, Select, Tab, TabList, TabPanel, TabPanels, Tabs, Text, Tooltip } from "@sourcegraph/wildcard"
import { setDeveloperSettingsFeatureFlagsView, setDeveloperSettingsTemporarySettingsView, toggleDevSettingsDialog, updateOverrideCounter, useDeveloperSettings, useOverrideCounter } from "../stores"
import { gql, useQuery } from "@sourcegraph/http-client"
import { FEATURE_FLAGS, FeatureFlagName } from "../featureFlags/featureFlags"
import styles from './DeveloperDialog.module.scss'
import { getFeatureFlagOverrideValue, removeFeatureFlagOverride, setFeatureFlagOverride } from "../featureFlags/lib/feature-flag-local-overrides"
import classNames from "classnames"
import { DeveloperSettingsEvaluatedFeatureFlagsResult} from "../graphql-operations"
import { TEMPORARY_SETTINGS_KEYS } from "@sourcegraph/shared/src/settings/temporary/TemporarySettings"
import { useTemporarySetting } from "@sourcegraph/shared/src/settings/temporary"
import { TemporarySettings } from "@sourcegraph/shared/src/settings/temporary/TemporarySettings"
import { setTemporarySettingOverride, removeTemporarySettingOverride, getTemporarySettingOverride } from "@sourcegraph/shared/src/settings/temporary/localOverride"
import { mdiChevronDown, mdiChevronRight } from "@mdi/js"
import { CodeMirrorEditor, defaultEditorTheme, jsonHighlighting } from "@sourcegraph/shared/src/components/CodeMirrorEditor"
import { EditorView } from "@codemirror/view"
import { EditorState } from "@codemirror/state"
import { json } from "@codemirror/lang-json"
import { useIsLightTheme } from "@sourcegraph/shared/src/theme"

interface DeveloperDialogProps {

}

export const DeveloperDialog: FC<DeveloperDialogProps> = () => {
    const counter = useOverrideCounter()
    const index = useDeveloperSettings(settings => settings.selectedTab)

    return (
        <Modal aria-label="Development settings" className={styles.dialog} onDismiss={() => toggleDevSettingsDialog(false)}>
            <H3>Developer Settings</H3>
            <Text>
                Here you can enforce or reset various settings during development. All
                settings are stored locally in the browser.
            </Text>
            <Tabs lazy={true} behavior="memoize" size="medium" className="overflow-hidden d-flex flex-column" index={index} onChange={index => useDeveloperSettings.setState({selectedTab: index})}>
                <TabList>
                    <Tab>Feature flags ({counter.featureFlags})</Tab>
                    <Tab>Temporary settings ({counter.temporarySettings})</Tab>
                </TabList>
                <TabPanels className="overflow-hidden flex-1 min-w-0 d-flex">
                    <TabPanel className={styles.content}>
                        <FeatureFlags />
                    </TabPanel>
                    <TabPanel className={styles.content}>
                        <TemporarySettingsPanel />
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
    const view = useDeveloperSettings(settings => settings.selectedView.featureFlags)
    const {data, loading} = useQuery<DeveloperSettingsEvaluatedFeatureFlagsResult>(EVALUATED_FEATURE_FLAGS, {fetchPolicy: 'cache-first'})

    const evaluatedFeatureFlags = useMemo(() => {
        const flags = new Map<string, boolean>
        for (const {name, value} of data?.evaluatedFeatureFlags ?? []) {
            flags.set(name, value)
        }
        return flags
    }, [data])

    if (loading) {
        return <div className="d-flex mt-3 align-items-center">
            <LoadingSpinner />
            <span>Loading flags...</span>
        </div>
    }

    return (
        <>
            <Alert variant="info" className="my-2">
                Click on the respective "Override value" entry to cycle through enabled, disabled and not set. You might have to reload the page after changing the value.
            </Alert>
            <div className="d-flex align-items-center my-2">
                <span>View: &nbsp;</span>
                <Select className="mb-0" aria-label="View" value={view} onChange={event => setDeveloperSettingsFeatureFlagsView(event.target.value)}>
                    <option>All</option>
                    <option>Enabled</option>
                    <option>Overridden</option>
                </Select>
            </div>
            <div className="flex-1 overflow-auto min-h-0 mt-2">
            <table>
                <thead>
                    <tr>
                        <th>Name</th>
                        <th className="text-center">Override value</th>
                        <th className="text-center">Actual value</th>
                    </tr>
                </thead>
                <tbody>
                    {FEATURE_FLAGS.slice().sort().map(featureFlag => <FeatureFlagOverride key={featureFlag} featureFlag={featureFlag} filter={view} serverValue={evaluatedFeatureFlags.get(featureFlag)}/>)}
                </tbody>
            </table>
            </div>
        </>
    )
}

const FeatureFlagOverride: FC<{featureFlag: FeatureFlagName, filter: string, serverValue?: boolean}> = memo(({featureFlag, filter, serverValue}) => {
    const [overrideValue, setOverrideValue] = useState(getFeatureFlagOverrideValue(featureFlag))

    const enabled = overrideValue === true || overrideValue === null && serverValue
    const overridden = overrideValue !== null

    function onClick() {
        switch (overrideValue) {
            case null:
                setFeatureFlagOverride(featureFlag, true)
                updateOverrideCounter()
                break;
            case true:
                setFeatureFlagOverride(featureFlag, false)
                break;
            case false:
                removeFeatureFlagOverride(featureFlag)
                updateOverrideCounter()
                break;
        }
        setOverrideValue(getFeatureFlagOverrideValue(featureFlag))
    }

    if (filter === 'Enabled' && !enabled || filter === 'Overridden' && !overridden) {
        return null
    }

    return (
        <tr className={classNames({[styles.enabled]:enabled, [styles.override]: overridden})}>
            <td>{featureFlag}</td>
            <td className="text-center"><Button variant="link" onClick={onClick}><code>{overrideValue === null ? 'n/a' : String(overrideValue)}</code></Button></td>
            <td className="text-center"><code>{String(serverValue ?? 'n/a')}</code></td>
        </tr>
    )
})

const TemporarySettingsPanel: FC<{}> = () => {
    const view = useDeveloperSettings(settings => settings.selectedView.temporarySettings)

    return (
        <>
            <Alert variant="info" className="my-2">
                Because we cannot check the validity of a temporary settings value, this UI only allows you to intercept and reset a temporary setting. Due to the nature of the API we can only show the current (overridden or actual) value.
            </Alert>
            <div className="d-flex align-items-center my-2">
                <span>View: &nbsp;</span>
                <Select className="mb-0" aria-label="View" value={view} onChange={event => setDeveloperSettingsTemporarySettingsView(event.target.value)}>
                    <option>All</option>
                    <option>Overridden</option>
                </Select>
            </div>
            <div className="flex-1 overflow-auto min-h-0 mt-2">
            <table>
                <thead>
                    <tr>
                        <th></th>
                        <th>Name</th>
                        <th className="text-center">Value</th>
                        <th className="text-center">Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {TEMPORARY_SETTINGS_KEYS.slice().sort().map(setting => <TemporarySettingOverride key={setting} setting={setting} filter={view} />)}
                </tbody>
            </table>
            </div>
        </>
    )
}

const TemporarySettingOverride: FC<{setting: keyof TemporarySettings, filter: string}> = memo(({setting, filter}) => {
    const [value] = useTemporarySetting(setting)
    const [overrideValue, setOverrideValue] = useState(getTemporarySettingOverride(setting))
    const [open, setOpen] = useState(false)

    const overridden = overrideValue !== null
    const isObject = typeof value === 'object' && value

    function toggleOpen(event: MouseEvent) {
        if (isObject) {
            event.stopPropagation()
            setOpen(open => !open)
        }
    }

    function reset(event: MouseEvent) {
        event.stopPropagation()
        setTemporarySettingOverride(setting, {value: undefined})
        setOverrideValue(getTemporarySettingOverride(setting))
        updateOverrideCounter()
    }

    function removeOverride(event: MouseEvent) {
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
        <tr className={classNames({[styles.override]: overridden})} onClick={toggleOpen}>
            <td>{isObject && <Button className="d-inline-flex" variant="icon" onClick={toggleOpen}><Icon svgPath={open ? mdiChevronDown : mdiChevronRight} aria-hidden={true}/></Button>}
            </td>
            <td>{setting}</td>
            <td className="text-center"><code>{value === undefined ? 'n/a' : isObject? 'object' : JSON.stringify(value)}</code></td>
            <td className="text-center">
                    <Button variant="link" onClick={reset} disabled={!!overrideValue && overrideValue.value === undefined}>
                        Clear
                    </Button>
                    <Tooltip content="Restore actual value (remove override)">
                        <Button className="mx-1" variant="link" onClick={removeOverride} disabled={!overridden}>
                            Restore
                        </Button>
                    </Tooltip>
            </td>
        </tr>
            {isObject && open && <tr><td colSpan={4}><JSONView value={value} /> </td></tr>}
        </>
    )
})

const JSONView: FC<{value: unknown}> = ({value}) => {
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

    return <CodeMirrorEditor value={jsonValue} extensions={extensions}/>
}
