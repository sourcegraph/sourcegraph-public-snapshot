/**
 * This file was automatically generated by json-schema-to-typescript.
 * DO NOT MODIFY IT BY HAND. Instead, modify the source JSONSchema file,
 * and run json-schema-to-typescript to regenerate this file.
 */

/**
 * The JSON Schema for the configuration settings used by this extension. This schema is merged with the Sourcegraph settings schema. The final schema for settings is the union of Sourcegraph settings and all added extensions' settings.
 */
export type CoreSchemaMetaSchema =
  | {
      [k: string]: any;
    }
  | boolean;

/**
 * The Sourcegraph extension manifest describes the extension and the features it provides.
 */
export interface SourcegraphExtensionManifest {
  /**
   * The title of the extension. If not specified, the extension ID is used.
   */
  title?: string;
  /**
   * The extension's description, which summarizes the extension's purpose and features. It should not exceed a few sentences.
   */
  description?: string;
  /**
   * The extension icon in data URI format (must begin with data:image/png).
   */
  icon?: string;
  /**
   * The extension's README, which should describe (in detail) the extension's purpose, features, and usage instructions. Markdown formatting is supported.
   */
  readme?: string;
  /**
   * A URL to a file containing the bundled JavaScript source code of this extension.
   */
  url: string;
  repository?: ExtensionRepository;
  /**
   * A list of events that cause this extension to be activated. '*' means that it will always be activated.
   */
  activationEvents: string[];
  /**
   * Arguments provided to the extension upon initialization (in the `initialize` message's `initializationOptions` field).
   */
  args?: {
    [k: string]: any;
  };
  contributes?: Contributions;
}
/**
 * The location of the version control repository for this extension.
 */
export interface ExtensionRepository {
  /**
   * The version control system (e.g. git).
   */
  type?: string;
  /**
   * A URL to the source code for this extension.
   */
  url: string;
}
/**
 * Features contributed by this extension. Extensions may also register certain types of contributions dynamically.
 */
export interface Contributions {
  configuration?: CoreSchemaMetaSchema;
  /**
   * The actions that this extension supports.
   */
  actions?: Action[];
  /**
   * Describes where to place actions in menus.
   */
  menus?: {
    /**
     * The file header.
     */
    "editor/title"?: MenuItem[];
    /**
     * The command palette (usually in the upper right).
     */
    commandPalette?: MenuItem[];
    /**
     * The help menu.
     */
    help?: MenuItem[];
    [k: string]: any;
  };
}
export interface Action {
  /**
   * The unique ID for this action.
   */
  id?: string;
  /**
   * The command to execute when this action is taken.
   */
  command?: string;
  /**
   * The arguments to the command.
   */
  commandArguments?: any[];
  /**
   * The templated text that is shown in the UI.
   */
  title?: string;
  /**
   * The templated prefix for the title (e.g. a category of `Codecov` shows up as `Codecov: ...` in the command palette).
   */
  category?: string;
  /**
   * The templated icon that is shown in the UI (usually a data URI).
   */
  iconURL?: string;
  /**
   * The action item.
   */
  actionItem?: {
    /**
     * The templated text that is shown on the action item the UI.
     */
    label?: string;
    /**
     * The templated tooltip text for an action item that is shown in the UI.
     */
    description?: string;
    /**
     * The templated icon that is shown in the UI (usually a data URI).
     */
    iconURL?: string;
    [k: string]: any;
  };
  [k: string]: any;
}
export interface MenuItem {
  /**
   * The ID of the action to take when this menu item is clicked.
   */
  action?: string;
  /**
   * The tooltip text to show when hovering over this menu item.
   */
  alt?: string;
  /**
   * An expression that determines whether or not to show this menu item.
   */
  when?: string;
  [k: string]: any;
}
