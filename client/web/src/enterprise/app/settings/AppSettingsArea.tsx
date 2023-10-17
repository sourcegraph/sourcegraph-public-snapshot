import type { FC } from 'react'

import AboutOutlineIcon from 'mdi-react/AboutOutlineIcon'
import { Routes, Route, Outlet, Navigate, useLocation } from 'react-router-dom'

import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Link, MenuDivider, PageHeader } from '@sourcegraph/wildcard'

import { RemoteRepositoriesStep } from '../../../setup-wizard/components'

import { AboutTab } from './about/AboutPage'
import { LocalRepositoriesTab } from './local-repositories/LocalRepositoriesTab'
import { RateLimitsTab } from './rate-limits/RateLimitsTab'

import styles from './AppSettingsArea.module.scss'

enum AppSettingURL {
    LocalRepositories = 'local-repositories',
    RemoteRepositories = 'remote-repositories',
    RateLimits = 'rate-limits',
    About = 'about',
}

export const AppSettingsArea: FC<TelemetryProps> = ({ telemetryService }) => (
    <Routes>
        <Route path="*" element={<AppSettingsLayout />}>
            <Route path={AppSettingURL.LocalRepositories} element={<LocalRepositoriesTab />} />
            <Route
                path={`${AppSettingURL.RemoteRepositories}/*`}
                element={<RemoteRepositoriesTab telemetryService={telemetryService} />}
            />
            <Route path={AppSettingURL.About} element={<AboutTab />} />
            <Route path={AppSettingURL.RateLimits} element={<RateLimitsTab />} />
            <Route path={AppSettingURL.About} element={<AboutTab />} />
            <Route path="*" element={<Navigate to={AppSettingURL.LocalRepositories} replace={true} />} />
        </Route>
    </Routes>
)

interface AppSetting {
    url: AppSettingURL
    name: string
}

const APP_SETTINGS: AppSetting[] = [
    { url: AppSettingURL.LocalRepositories, name: 'Local repositories' },
    { url: AppSettingURL.RemoteRepositories, name: 'Remote repositories' },
    { url: AppSettingURL.RateLimits, name: 'Usage Limits' },
]

const AppSettingsLayout: FC = () => {
    const location = useLocation()

    return (
        <div className={styles.root}>
            <ul className={styles.navigation}>
                {APP_SETTINGS.map(setting => (
                    <li key={setting.url}>
                        <Button
                            as={Link}
                            to={`../${setting.url}`}
                            variant={location.pathname.includes(`/${setting.url}`) ? 'primary' : undefined}
                            className={styles.navigationItemLink}
                        >
                            {setting.name}
                        </Button>
                    </li>
                ))}
                <li>
                    <MenuDivider />
                </li>
                <li>
                    <Button
                        as={Link}
                        to="../about"
                        variant={location.pathname.includes(AppSettingURL.About) ? 'primary' : undefined}
                        className={styles.navigationItemLink}
                    >
                        <AboutOutlineIcon size={16} /> About Cody
                    </Button>
                </li>
            </ul>

            <Outlet />
        </div>
    )
}

const RemoteRepositoriesTab: FC<TelemetryProps> = ({ telemetryService }) => (
    <div className={styles.content}>
        <PageHeader headingElement="h2" path={[{ text: 'Remote repositories' }]} className="mb-3" />

        <RemoteRepositoriesStep
            baseURL={`app-settings/${AppSettingURL.RemoteRepositories}`}
            description={false}
            progressBar={false}
            telemetryService={telemetryService}
            isCodyApp={true}
        />
    </div>
)
