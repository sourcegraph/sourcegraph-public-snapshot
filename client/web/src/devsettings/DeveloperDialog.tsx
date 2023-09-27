import { FC, memo, useState } from "react"
import { AuthenticatedUser } from "../auth"
import { Button, H3, Modal, Select, Tab, TabList, TabPanel, TabPanels, Tabs, Text } from "@sourcegraph/wildcard"
import { toggleDevSettingsDialog } from "../stores"
import { gql, useQuery } from "@sourcegraph/http-client"
import { FEATURE_FLAGS, FeatureFlagName } from "../featureFlags/featureFlags"
import styles from './DeveloperDialog.module.scss'
import { EVALUATE_FEATURE_FLAG_QUERY } from "../featureFlags/useFeatureFlag"
import { getFeatureFlagOverrideValue, removeFeatureFlagOverride, setFeatureFlagOverride } from "../featureFlags/lib/feature-flag-local-overrides"
import classNames from "classnames"
import { EvaluateFeatureFlagResult, EvaluateFeatureFlagVariables } from "../graphql-operations"

interface DeveloperDialogProps {
    authenticatedUser: AuthenticatedUser
}

export const DeveloperDialog: FC<DeveloperDialogProps> = ({authenticatedUser}) => {
    return (
        <Modal aria-label="Development settings" className={styles.dialog} onDismiss={() => toggleDevSettingsDialog(false)}>
            <H3>Developer Settings</H3>
            <Text>
                Here you can enforce or reset various settings during development. All
                settings are stored locally in the browser.
            </Text>
            <Tabs lazy={true} behavior="memoize" className="overflow-hidden d-flex flex-column">
                <TabList>
                    <Tab>Feature flags</Tab>
                    <Tab>Temporary settings</Tab>
                </TabList>
                <TabPanels className="overflow-hidden flex-1 min-w-0 d-flex">
                    <TabPanel className={styles.content}>
                        <FeatureFlags />
                    </TabPanel>
                    <TabPanel>
                    </TabPanel>
                </TabPanels>

            </Tabs>
        </Modal>
    )
}

const FeatureFlags: FC<{}> = () => {
    const [filter, setFilter] = useState('All')
    return (
        <>
            <Text className="m-1">
                Click on the respective "Local value" entry to cycle through enabled, disabled and not set.
            </Text>
            <div className="d-flex align-items-center m-1">
                <span>View: &nbsp;</span>
                <Select className="mb-0" aria-label="View" value={filter} onChange={event => setFilter(event.target.value)}>
                    <option>All</option>
                    <option>Enabled</option>
                    <option>Overridden</option>
                </Select>
            </div>
            <div className="flex-1 overflow-auto min-h-0 mt-1">
            <table>
                <thead>
                    <tr>
                        <th>Name</th>
                        <th>Local value</th>
                        <th>Server value</th>
                    </tr>
                </thead>
                <tbody>
                    {FEATURE_FLAGS.slice().sort().map(featureFlag => <FeatureFlagOverride key={featureFlag} featureFlag={featureFlag} filter={filter}/>)}
                </tbody>
            </table>
            </div>
        </>
    )
}

const FeatureFlagOverride: FC<{featureFlag: FeatureFlagName, filter: string}> = memo(({featureFlag, filter}) => {
    // We cannot use 'useFeatureFlag' here because it itself returns the overriden value.
    // We always want to show the server value here.
    const { data, loading } = useQuery<EvaluateFeatureFlagResult, EvaluateFeatureFlagVariables>(
        EVALUATE_FEATURE_FLAG_QUERY,
        {
            variables: { flagName: featureFlag },
            fetchPolicy: 'cache-first',
        }
    )
    const serverValue = data?.evaluateFeatureFlag
    const [overrideValue, setOverrideValue] = useState(getFeatureFlagOverrideValue(featureFlag))

    const enabled = overrideValue === true || overrideValue === null && serverValue
    const overridden = overrideValue !== null

    function onClick() {
        switch (overrideValue) {
            case null:
                setFeatureFlagOverride(featureFlag, true)
                break;
            case true:
                setFeatureFlagOverride(featureFlag, false)
                break;
            case false:
                removeFeatureFlagOverride(featureFlag)
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
            <td><Button variant="link" onClick={onClick}><code>{overrideValue === null ? 'n/a' : String(overrideValue)}</code></Button></td>
            <td><code>{loading ? '(loading...)' : String(serverValue ?? 'n/a')}</code></td>
        </tr>
    )
})
