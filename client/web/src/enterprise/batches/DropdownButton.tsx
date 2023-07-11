import React, { useCallback, useEffect, useMemo, useState, useContext } from 'react';
import { mdiChevronDown, mdiDownload } from '@mdi/js';
import { VisuallyHidden } from '@reach/visually-hidden';
import {
  ProductStatusBadge,
  Button,
  ButtonGroup,
  Menu,
  MenuButton,
  MenuList,
  Position,
  MenuItem,
  MenuDivider,
  H4,
  Text,
  Icon,
} from '@sourcegraph/wildcard';
import { useBatchChangesRolloutWindowConfig } from './backend';
import styles from './DropdownButton.module.scss';



export interface Action {
  /* The type of action. Used internally. */
  type: string;
  /* The button label for the action. */
  buttonLabel: string;
  /* Whether or not the action is disabled. */
  disabled?: boolean;
  /* The title in the dropdown menu item. */
  dropdownTitle: string;
  /* The description in the dropdown menu item. */
  dropdownDescription: string;
  /**
   * Invoked when the action is triggered. Either onDone or onCancel need to
   * be called eventually. Can return a JSX.Element to be rendered adjacent to
   * the button (i.e. a modal).
   */
  onTrigger: (onDone: () => void, onCancel: () => void) => Promise<void | JSX.Element> | void | JSX.Element;
  /** If set, displays an experimental badge next to the dropdown title. */
  experimental?: boolean;
  changeset: any;
}

export interface Props {
  actions: Action[];
  defaultAction?: number;
  disabled?: boolean;
  onLabel?: (label: string | undefined) => void;
  placeholder?: string;
  selectedChangesets: any[]
}

export const DropdownButton: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
  actions,
  defaultAction,
  disabled,
  onLabel,
  placeholder = 'Select action',
  selectedChangesets,
}) => {
  const [isDisabled, setIsDisabled] = useState(!!disabled);
  const [selected, setSelected] = useState<number | undefined>(undefined);


  const selectedAction = useMemo(() => {
    if (actions.length === 1) {
      return actions[0];
    }

    const id = selected !== undefined ? selected : defaultAction;
    if (id !== undefined && id >= 0 && id < actions.length) {
      return actions[id];
    }
    return undefined;
  }, [actions, defaultAction, selected]);

  const onSelectedTypeSelect = useCallback(
    (type: string) => {
      const index = actions.findIndex((action) => action.type === type);
      if (index >= 0) {
        setSelected(actions.findIndex((action) => action.type === type));
      } else {
        setSelected(undefined);
      }
    },
    [actions, setSelected]
  );

  const [renderedElement, setRenderedElement] = useState<JSX.Element | undefined>();

  const onTriggerAction = useCallback(async () => {
    if (selectedAction === undefined) {
      return;
    }

    // Right now, we don't handle onDone or onCancel separately, but we may
    // want to expose this at a later stage.
    setIsDisabled(true);
    const element = await Promise.resolve(
      selectedAction.onTrigger(
        () => {
          setIsDisabled(false);
          setRenderedElement(undefined);
        },
        () => {
          setIsDisabled(false);
          setRenderedElement(undefined);
        }
      )
    );
    if (element !== undefined) {
      setRenderedElement(element);
    }
  }, [selectedAction]);

  const selectedChangesetExportClick = useCallback(() => {
    if (selectedChangesets) {
      const csvData = [
        ['Title', 'State', 'CheckState', 'ReviewState', 'External URL', 'Repo Name'],
      ];
      
      selectedChangesets.forEach(node => {
        csvData.push([
          node.title,
          node.state,
          node.checkState,
          node.reviewState,
          node.externalURL?.url || '',
          node.repository.name,
        ]);
      })

        const csvString = csvData.map((row) => row.join(',')).join('\n');
        const blob = new Blob([csvString], { type: 'text/csv' });
        const url = URL.createObjectURL(blob);

        const body = selectedChangesets[0].body; // Assuming selectedChangesets exists
        const startIndex = body.indexOf("batch-changes/") + 14; // Add the length of "batch-changes/"
        const endIndex = body.indexOf(")", startIndex);
        const batchChangeName = body.substring(startIndex, endIndex);

        const a = document.createElement('a');
        a.href = url;
        a.download = `${batchChangeName}-changesets.csv`; 
        a.click();
  
        URL.revokeObjectURL(url);
        a.remove();
      }
    },[selectedChangesets]);




  
  

  const label = useMemo(() => {
    const label =
      selectedAction && selectedAction.experimental
        ? `${selectedAction.buttonLabel} (Experimental)`
        : selectedAction?.buttonLabel;

    return label ?? placeholder;
  }, [placeholder, selectedAction]);
  
 

  useEffect(() => {
    if (onLabel && selectedAction) {
      onLabel(
        selectedAction.experimental ? `${selectedAction.buttonLabel} (Experimental)` : selectedAction.buttonLabel
      );
    }
  });

  return (
    <>
      {renderedElement}
      <Menu>
        <ButtonGroup>
          <Button
            className="text-nowrap"
            onClick={onTriggerAction}
            disabled={isDisabled || actions.length === 0 || selectedAction === undefined}
            variant="primary"
          >
            {label}
          </Button>
          {actions.length > 1 && (
            <MenuButton variant="primary" className={styles.dropdownButton}>
              <Icon svgPath={mdiChevronDown} inline={false} aria-hidden={true} />
              <VisuallyHidden>Actions</VisuallyHidden>
            </MenuButton>
          )}

          {/* Add the export button */}
          <div style={{ marginLeft: '8px' }}>
          <Button className="text-nowrap" variant="primary" 
          onClick={selectedChangesetExportClick}
          >
                    <Icon aria-hidden={true} svgPath={mdiDownload} /> Export 
                </Button> 
          </div>

        </ButtonGroup>
        {actions.length > 1 && (
          <MenuList className={styles.menuList} position={Position.bottomEnd}>
            {actions.map((action, index) => (
              <React.Fragment key={action.type}>
                <DropdownItem action={action} setSelectedType={onSelectedTypeSelect} />
                {index !== actions.length - 1 && <MenuDivider />}
              </React.Fragment>
            ))}
          </MenuList>
        )}
      </Menu>
    </>
  );
};

interface DropdownItemProps {
  setSelectedType: (type: string) => void;
  action: Action;
}

const DropdownItem: React.FunctionComponent<React.PropsWithChildren<DropdownItemProps>> = ({
  action,
  setSelectedType,
}) => {
  const { rolloutWindowConfig, loading } = useBatchChangesRolloutWindowConfig();
  const onSelect = useCallback(() => {
    setSelectedType(action.type);
  }, [setSelectedType, action.type]);
  const shouldDisplayRolloutInfo = action.type === 'publish' && rolloutWindowConfig && rolloutWindowConfig.length > 0;

  return (
    <MenuItem className={styles.menuListItem} onSelect={onSelect} disabled={action.disabled}>
      <H4 className="mb-1">
        {action.dropdownTitle}
        {action.experimental && (
          <>
            {' '}
            <ProductStatusBadge status="experimental" as="small" />
          </>
        )}
      </H4>
      <Text className="text-wrap text-muted mb-0">
        <small>
          {action.dropdownDescription}
          {!loading && shouldDisplayRolloutInfo && (
            <>
              <br />
              <strong>
                Note: Rollout windows have been set up by the admin. This means that some of the selected changesets won't
                be processed until a time in the future.
              </strong>
            </>
          )}
        </small>
      </Text>
    </MenuItem>
  );
};
