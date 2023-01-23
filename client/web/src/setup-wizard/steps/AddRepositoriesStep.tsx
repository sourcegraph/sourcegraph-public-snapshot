import { FC, useState } from 'react'

import { mdiBitbucket, mdiGithub, mdiGitlab } from '@mdi/js'

import { Checkbox, Icon, Input, RadioButton, RadioButtonProps } from '@sourcegraph/wildcard'

import { SetupStep, SetupStepActions } from '../components/SetupTabs'

import styles from './AddRepositoriesStep.module.scss'

enum RepositoriesPickingMode {
    All,
    OrganizationsList,
    SelectedRepositories,
}

export const AddRepositoriesStep: FC = props => {
    const [mode, setMode] = useState<RepositoriesPickingMode>(RepositoriesPickingMode.All)

    return (
        <SetupStep>
            {/* eslint-disable-next-line react/forbid-elements */}
            <form className={styles.repositoriesRoot}>
                <RadioButton
                    id="mode-1"
                    name="repositories-mode"
                    value="all-repositories"
                    label="All repositories"
                    message="We will try to resolve all reachable by your access token repositories (private repositories, all repositories from your organizations"
                    checked={mode === RepositoriesPickingMode.All}
                    onChange={() => setMode(RepositoriesPickingMode.All)}
                />

                <RadioButtonWithContent
                    id="mode-2"
                    name="repositories-mode"
                    value="org-repositories"
                    label="Sync all repositories from selected organizations"
                    message="We will try to resolve all repositories within selected organizations"
                    checked={mode === RepositoriesPickingMode.OrganizationsList}
                    onChange={() => setMode(RepositoriesPickingMode.OrganizationsList)}
                >
                    <OrganizationsList />
                </RadioButtonWithContent>

                <RadioButtonWithContent
                    id="mode-3"
                    name="repositories-mode"
                    value="selected-repositories"
                    label="Sync selected repositories"
                    message="Pick a specific set of repositories"
                    checked={mode === RepositoriesPickingMode.SelectedRepositories}
                    onChange={() => setMode(RepositoriesPickingMode.SelectedRepositories)}
                >
                    <RepositoriesList />
                </RadioButtonWithContent>
            </form>

            <SetupStepActions nextAvailable={true} />
        </SetupStep>
    )
}

const RadioButtonWithContent: FC<RadioButtonProps> = props => {
    const { checked, className, children, ...attributes } = props

    return (
        <fieldset className={className}>
            <RadioButton {...attributes} checked={checked} />
            {checked && <div className="ml-4 mt-2">{children}</div>}
        </fieldset>
    )
}

const OrganizationsList: FC = () => (
    <section>
        <Input
            label="Search organizations"
            placeholder="sourcegraph/"
            autoFocus={true}
            variant="small"
            className="mb-2"
        />

        <ul className={styles.list}>
            <li className={styles.item}>
                <Checkbox className={styles.checkbox} /> <Icon svgPath={mdiGithub} className={styles.icon} />{' '}
                Sourcegraph
            </li>
            <li className={styles.item}>
                <Checkbox className={styles.checkbox} /> <Icon svgPath={mdiGithub} className={styles.icon} /> My
                Personal organization
            </li>
            <li className={styles.item}>
                <Checkbox className={styles.checkbox} /> <Icon svgPath={mdiGitlab} className={styles.icon} /> Gitlab org
            </li>
            <li className={styles.item}>
                <Checkbox className={styles.checkbox} /> <Icon svgPath={mdiBitbucket} className={styles.icon} />{' '}
                BitBucket org
            </li>
            <li className={styles.item}>
                <Checkbox className={styles.checkbox} /> <Icon svgPath={mdiBitbucket} className={styles.icon} />{' '}
                Experiment project org
            </li>
        </ul>
    </section>
)

const RepositoriesList: FC = () => (
    <section>
        <Input
            label="Search repositories"
            placeholder="sourcegraph/sourcegraph"
            autoFocus={true}
            variant="small"
            className="mb-2"
        />

        <ul className={styles.list}>
            <li className={styles.item}>
                <Checkbox className={styles.checkbox} /> <Icon svgPath={mdiGithub} className={styles.icon} />{' '}
                sourcegraph/sourcegraph
            </li>
            <li className={styles.item}>
                <Checkbox className={styles.checkbox} /> <Icon svgPath={mdiGithub} className={styles.icon} />{' '}
                sourcegraph/sourcegraph
            </li>
            <li className={styles.item}>
                <Checkbox className={styles.checkbox} /> <Icon svgPath={mdiGitlab} className={styles.icon} />{' '}
                vovakulikov/looper
            </li>
            <li className={styles.item}>
                <Checkbox className={styles.checkbox} /> <Icon svgPath={mdiBitbucket} className={styles.icon} />{' '}
                vovakulikov/tokio
            </li>
            <li className={styles.item}>
                <Checkbox className={styles.checkbox} /> <Icon svgPath={mdiBitbucket} className={styles.icon} />{' '}
                vovakulikov/go-private
            </li>
            <li className={styles.item}>
                <Checkbox className={styles.checkbox} /> <Icon svgPath={mdiGithub} className={styles.icon} />{' '}
                sourcegraph/sourcegraph
            </li>
            <li className={styles.item}>
                <Checkbox className={styles.checkbox} /> <Icon svgPath={mdiGitlab} className={styles.icon} />{' '}
                vovakulikov/looper
            </li>
            <li className={styles.item}>
                <Checkbox className={styles.checkbox} /> <Icon svgPath={mdiGithub} className={styles.icon} />{' '}
                sourcegraph/sourcegraph
            </li>
            <li className={styles.item}>
                <Checkbox className={styles.checkbox} /> <Icon svgPath={mdiGitlab} className={styles.icon} />{' '}
                vovakulikov/looper
            </li>
            <li className={styles.item}>
                <Checkbox className={styles.checkbox} /> <Icon svgPath={mdiGithub} className={styles.icon} />{' '}
                sourcegraph/sourcegraph
            </li>
            <li className={styles.item}>
                <Checkbox className={styles.checkbox} /> <Icon svgPath={mdiGitlab} className={styles.icon} />{' '}
                vovakulikov/looper
            </li>
        </ul>
    </section>
)
