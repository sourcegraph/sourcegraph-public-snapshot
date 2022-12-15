import { ChangeEvent, FC, PropsWithChildren, ReactElement } from 'react'

import classNames from 'classnames'
import LinkExternalIcon from 'mdi-react/OpenInNewIcon'

import { EditorHint, QueryState, SearchPatternType } from '@sourcegraph/search'
import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import {
    Button,
    Checkbox,
    Code,
    Input,
    Label,
    InputElement,
    InputErrorMessage,
    InputDescription,
    Link,
} from '@sourcegraph/wildcard'

import { useExperimentalFeatures } from '../../../../../../../../stores'
import {
    FormGroup,
    getDefaultInputProps,
    getDefaultInputStatus,
    getDefaultInputError,
    RepositoriesField,
    useFieldAPI,
    MonacoField,
} from '../../../../../../components'
import { MonacoPreviewLink } from '../../../../../../components/form/monaco-field'
import { CreateInsightFormFields, RepoMode } from '../../types'

import styles from './InsightRepoSection.module.scss'

interface RepoSettingSectionProps {
    repositories: useFieldAPI<CreateInsightFormFields['repositories']>
    allReposMode: useFieldAPI<CreateInsightFormFields['allRepos']>
    repoQuery: useFieldAPI<CreateInsightFormFields['repoQuery']>
    repoMode: useFieldAPI<CreateInsightFormFields['repoMode']>
}

/**
 * Main entry point for the repositories insight setting section for the creation UI.
 * It contains all possible variation for the repo setting section based on feature and
 * experimental flags.
 */
export const RepoSettingSection: FC<RepoSettingSectionProps> = props => {
    const { repositories, allReposMode, repoQuery, repoMode } = props

    const repoUIVariation = useExperimentalFeatures(features => features.codeInsightsRepoUI)

    if (repoUIVariation === 'old-strict-list') {
        return <OldRepoSettingSection allReposMode={allReposMode} repositories={repositories} />
    }

    if (repoUIVariation === 'single-search-query') {
        return <SmartRepoSettingSection repoQuery={repoQuery} />
    }

    return <SearchQueryOrRepoListSection repoMode={repoMode} repoQuery={repoQuery} repositories={repositories} />
}

interface OldRepoSettingSectionProps {
    repositories: useFieldAPI<CreateInsightFormFields['repositories']>
    allReposMode: useFieldAPI<CreateInsightFormFields['allRepos']>
}

/**
 * This repo form section provides a standard UI for picking repositories URL
 * for the insight creation form. Strict list of direct repo URLS and all repositories
 * mode checkbox.
 *
 * @deprecated (Remove this section as soon as strat-scoped insight UI is merged)
 */
export const OldRepoSettingSection: FC<OldRepoSettingSectionProps> = props => {
    const { repositories, allReposMode } = props

    return (
        <FormGroup
            name="insight repositories"
            title="Targeted repositories"
            subtitle="Create a list of repositories to run your search over"
        >
            <Input
                as={RepositoriesField}
                autoFocus={true}
                required={true}
                label="Repositories"
                message="Separate repositories with commas"
                placeholder={
                    allReposMode.input.value ? 'All repositories' : 'Example: github.com/sourcegraph/sourcegraph'
                }
                className="mb-0 d-flex flex-column"
                {...getDefaultInputProps(repositories)}
            />

            <Checkbox
                {...allReposMode.input}
                type="checkbox"
                id="RunInsightsOnAllRepoCheck"
                wrapperClassName="mb-1 mt-3 font-weight-normal"
                value="all-repos-mode"
                checked={allReposMode.input.value}
                label="Run your insight over all your repositories"
            />

            <small className="w-100 mt-2 text-muted">
                This feature is actively in development. Read about the{' '}
                <Link
                    to="/help/code_insights/explanations/current_limitations_of_code_insights"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    limitations here.
                </Link>
            </small>
        </FormGroup>
    )
}

interface SmartRepoSettingSectionProps {
    repoQuery: useFieldAPI<CreateInsightFormFields['repoQuery']>
}

/**
 * Single smart search repo query field, this section is one of possible variation for the
 * strat-scoped insight repo query UI.
 */
export const SmartRepoSettingSection: FC<SmartRepoSettingSectionProps> = props => {
    const { repoQuery } = props

    return (
        <FormGroup name="insight repositories" title="Targeted repositories">
            <SmartSearchQueryRepoField repoQuery={repoQuery} />
        </FormGroup>
    )
}

interface SearchQueryOrRepoListSectionProps {
    repoQuery: useFieldAPI<CreateInsightFormFields['repoQuery']>
    repoMode: useFieldAPI<CreateInsightFormFields['repoMode']>
    repositories: useFieldAPI<CreateInsightFormFields['repositories']>
}

export const SearchQueryOrRepoListSection: FC<SearchQueryOrRepoListSectionProps> = props => {
    const { repoQuery, repoMode, repositories } = props

    return (
        <FormGroup
            name="insight repositories"
            title="Targeted repositories"
            contentClassName={styles.radioGroupSection}
        >
            <RadioGroupSection
                label="Repositories query"
                name="repoMode"
                labelId="smart-repo-search-query"
                value={RepoMode.SearchQuery.toString()}
                checked={RepoMode.SearchQuery.toString() === repoMode.input.value.toString()}
                onChange={repoMode.input.onChange}
            >
                <SmartSearchQueryRepoField
                    repoQuery={repoQuery}
                    disabled={RepoMode.SearchQuery.toString() !== repoMode.input.value.toString()}
                    aria-labelledby="smart-repo-search-query"
                />
            </RadioGroupSection>

            <RadioGroupSection
                label="Explicit list of repositories"
                name="repoMode"
                labelId="strict-list-repo"
                value={RepoMode.DirectURLList.toString()}
                checked={RepoMode.DirectURLList.toString() === repoMode.input.value.toString()}
                onChange={repoMode.input.onChange}
            >
                <Input
                    as={RepositoriesField}
                    required={true}
                    message="Use a full repo URL (github.com/...). Separate repositories with comas"
                    placeholder="Example: github.com/sourcegraph/sourcegraph"
                    aria-labelledby="strict-list-repo"
                    {...getDefaultInputProps(repositories)}
                />
            </RadioGroupSection>
        </FormGroup>
    )
}

interface RadioGroupSectionProps {
    name: string
    label: string
    value: string
    checked: boolean
    labelId: string
    className?: string
    contentClassName?: string
    onChange: (event: ChangeEvent<HTMLInputElement>) => void
}

function RadioGroupSection(props: PropsWithChildren<RadioGroupSectionProps>): ReactElement {
    const { name, label, value, checked, labelId, children, onChange } = props

    return (
        <div className={styles.radioGroup}>
            {/* eslint-disable-next-line react/forbid-elements */}
            <input
                id={labelId}
                name={name}
                type="radio"
                value={value}
                checked={checked}
                className={styles.radioGroupInput}
                onChange={onChange}
            />
            <Label htmlFor={labelId} className={styles.radioGroupLabel}>
                {label}
            </Label>
            <div
                className={classNames(styles.radioGroupContent, {
                    [styles.radioGroupContentNonActive]: !checked,
                })}
            >
                {children}
            </div>
        </div>
    )
}

interface SmartSearchQueryRepoFieldProps {
    repoQuery: useFieldAPI<CreateInsightFormFields['repoQuery']>
    label?: string
    disabled?: boolean
    'aria-labelledby'?: string
}

function SmartSearchQueryRepoField(props: SmartSearchQueryRepoFieldProps): ReactElement {
    const { repoQuery, label, disabled, 'aria-labelledby': ariaLabelledby } = props

    const { value, onChange, ...attributes } = repoQuery.input

    const handleChipSuggestions = (chip: SmartRepoQueryChip): void => {
        const nextQueryValue = `${value.query} ${chip.query}`.trimStart()
        onChange({ query: nextQueryValue, hint: EditorHint.Focus })
    }

    const handleOnChange = (queryState: QueryState): void => {
        if (queryState.query !== value.query) {
            onChange(queryState)
        }
    }

    const LabelComponent = label ? Label : 'div'

    return (
        <div>
            <LabelComponent className={styles.repoLabel} id="search-repo-query">
                {label && <span className={styles.repoLabelText}>Repositories query</span>}

                <InputElement
                    as={MonacoField}
                    queryState={value}
                    autoFocus={true}
                    status={getDefaultInputStatus(repoQuery)}
                    placeholder="Example: repo:^github\.com/sourcegraph/sourcegraph$"
                    aria-labelledby={ariaLabelledby ?? 'search-repo-query'}
                    className={styles.repoInput}
                    onChange={handleOnChange}
                    disabled={disabled}
                    {...attributes}
                />

                <MonacoPreviewLink
                    query={value.query}
                    patternType={SearchPatternType.standard}
                    className={styles.repoLabelPreviewLink}
                >
                    <LinkExternalIcon size={18} />
                </MonacoPreviewLink>
            </LabelComponent>

            <SmartRepoQueryChips onChipClick={handleChipSuggestions} />

            {getDefaultInputError(repoQuery) && <InputErrorMessage message={getDefaultInputError(repoQuery)} />}

            <InputDescription>
                <ul>
                    <li>
                        Hint: you can use regular expressions within each of the <Code weight="bold">before:</Code>{' '}
                        available filters
                    </li>
                    <li>
                        Datapoints will be automatically backfilled using the list <Code weight="bold">before:</Code> of
                        repositories resulting from today’s search. Future data points will use the list refreshed for
                        every snapshot.
                    </li>
                </ul>
            </InputDescription>
        </div>
    )
}

interface SmartRepoQueryChip {
    id: string
    query: string
}

const CHIP_QUERIES: SmartRepoQueryChip[] = [
    { id: '1', query: 'repo:' },
    { id: '2', query: '-repo:' },
    { id: '3', query: 'AND' },
    { id: '4', query: 'OR' },
    { id: '5', query: 'NOT' },
    { id: '6', query: 'select:repo' },
    { id: '7', query: 'repo:has.path()' },
    { id: '8', query: 'repo:has.file()' },
    { id: '9', query: 'repo:has.commit.after()' },
    { id: '10', query: 'repo:.*' },
]

interface SmartRepoQueryChipsProps {
    onChipClick: (chip: SmartRepoQueryChip) => void
}

function SmartRepoQueryChips(props: SmartRepoQueryChipsProps): ReactElement {
    const { onChipClick } = props

    return (
        <ul className={styles.chipsList}>
            {CHIP_QUERIES.map(chip => (
                <li key={chip.id}>
                    <Button type="button" className={styles.queryChip} onClick={() => onChipClick(chip)}>
                        <SyntaxHighlightedSearchQuery
                            query={chip.query}
                            searchPatternType={SearchPatternType.standard}
                        />
                    </Button>
                </li>
            ))}
        </ul>
    )
}
