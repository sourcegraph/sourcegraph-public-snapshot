import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { ObjectCampaignsList } from '../../campaigns/object/ObjectCampaignsList'
import { ChangesetAreaContext } from './ChangesetArea'
import { ChangesetHeaderEditableTitle } from './header/ChangesetHeaderEditableTitle'

interface Props extends Pick<ChangesetAreaContext, 'changeset' | 'onChangesetUpdate'>, ExtensionsControllerProps {
    className?: string
}

/**
 * The overview for a single changeset.
 */
export const ChangesetOverview: React.FunctionComponent<Props> = ({
    changeset,
    onChangesetUpdate,
    className = '',
    ...props
}) => (
    <div className={`changeset-overview ${className || ''}`}>
        <ChangesetHeaderEditableTitle
            {...props}
            changeset={changeset}
            onChangesetUpdate={onChangesetUpdate}
            className="mb-3"
        />
        <hr />
        <div className="d-flex mt-4">
            <div className="flex-1">
                {changeset.externalURL && (
                    <a href={changeset.externalURL}>
                        <ExternalLinkIcon className="icon-inline mr-1" /> View pull request
                    </a>
                )}
            </div>
            <aside style={{ width: '12rem' }} className="ml-4">
                <section>
                    <h6 className="text-muted font-size-base mb-0">Campaign</h6>
                    <ObjectCampaignsList object={changeset} />
                </section>
            </aside>
        </div>
    </div>
)
