import React, { useCallback, useState } from 'react'
import H from 'history'
import {
    ButtonDropdown,
    DropdownToggle,
    DropdownMenu,
    Nav,
    NavLink,
    TabContent,
    TabPane,
    DropdownItem,
} from 'reactstrap'
import { map } from 'rxjs/operators'
import { NotificationType } from '../../../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../../../backend/graphql'
import { ImportThreadByRepositoryAndNumberFromExternalServiceForm } from '../../../threads/common/importForm/ImportThreadByRepositoryAndNumberFromExternalServiceForm'
import { addThreadsToCampaign } from './AddThreadToCampaignDropdownButton'
import { useLocalStorage } from '../../../../../util/useLocalStorage'
import { ImportThreadsByQueryFromExternalServiceForm } from '../../../threads/common/importForm/ImportThreadsByQueryFromExternalServiceForm'

export const importThreadsFromExternalService = (
    input: GQL.IImportThreadsFromExternalServiceInput
): Promise<Pick<GQL.IThread, 'id'>[]> =>
    mutateGraphQL(
        gql`
            mutation ImportThreadsFromExternalService($input: ImportThreadsFromExternalServiceInput!) {
                importThreadsFromExternalService(input: $input) {
                    id
                }
            }
        `,
        { input }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.importThreadsFromExternalService)
        )
        .toPromise()

interface Props extends ExtensionsControllerNotificationProps {
    campaign: Pick<GQL.IExpCampaign, 'id'>
    onChange?: () => void

    className?: string
    location: H.Location
}

export const ImportThreadsFromExternalServiceToCampaignDropdownButton: React.FunctionComponent<Props> = ({
    campaign,
    onChange,
    className = '',
    location,
    extensionsController,
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const [isLoading, setIsLoading] = useState(false)
    const onThreadsSelect = useCallback(
        async (threads: Pick<GQL.IThread, 'id'>[]) => {
            setIsLoading(true)
            try {
                await addThreadsToCampaign({ campaign: campaign.id, threads: threads.map(({ id }) => id) })
                setIsOpen(false)
                setIsLoading(false)
                if (onChange) {
                    onChange()
                }
            } catch (err) {
                setIsLoading(false)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error adding threads to campaign: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [campaign.id, extensionsController.services.notifications.showMessages, onChange]
    )

    type TabID = 'repository-and-number' | 'query'
    const [activeTab, setActiveTab] = useLocalStorage<TabID>(
        'ImportThreadsFromExternalServiceToCampaignDropdownButton.activeTab',
        'repository-and-number'
    )

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className={className} direction="down">
            <DropdownToggle color="" className="btn btn-primary" caret={true}>
                Add threads
            </DropdownToggle>
            <DropdownMenu className="pt-0">
                <Nav pills={true}>
                    <NavLink
                        className={`flex-1 text-nowrap ${activeTab === 'repository-and-number' ? 'active' : ''}`}
                        href="#"
                        // eslint-disable-next-line react/jsx-no-bind
                        onClick={() => setActiveTab('repository-and-number')}
                    >
                        By number
                    </NavLink>
                    <NavLink
                        className={`flex-1 text-nowrap ${activeTab === 'query' ? 'active' : ''}`}
                        href="#"
                        // eslint-disable-next-line react/jsx-no-bind
                        onClick={() => setActiveTab('query')}
                    >
                        By query
                    </NavLink>
                </Nav>
                <DropdownItem divider={true} />
                <TabContent activeTab={activeTab}>
                    <TabPane tabId="repository-and-number">
                        <ImportThreadByRepositoryAndNumberFromExternalServiceForm
                            onThreadsSelect={onThreadsSelect}
                            className="p-3"
                            disabled={isLoading}
                            extensionsController={extensionsController}
                        />
                    </TabPane>
                    <TabPane tabId="query">
                        <ImportThreadsByQueryFromExternalServiceForm
                            onThreadsSelect={onThreadsSelect}
                            className="p-3"
                            disabled={isLoading}
                            extensionsController={extensionsController}
                        />
                    </TabPane>
                </TabContent>
            </DropdownMenu>
        </ButtonDropdown>
    )
}
