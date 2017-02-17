import { code_font_face } from "sourcegraph/components/styles/_vars.css";
import Event, { Emitter } from "vs/base/common/event";
import { TPromise } from "vs/base/common/winjs.base";
import { IConfigurationKeys, IConfigurationService, IConfigurationServiceEvent, IConfigurationValue, getConfigurationValue } from "vs/platform/configuration/common/configuration";
import { IWorkspaceConfigurationKeys, IWorkspaceConfigurationService, IWorkspaceConfigurationValue, IWorkspaceConfigurationValues } from "vs/workbench/services/configuration/common/configuration";

import { Features } from "sourcegraph/util/features";

const _onDidUpdateConfiguration = new Emitter<IConfigurationServiceEvent>();
const onDidUpdateConfiguration = _onDidUpdateConfiguration.event;

let codeLensEnabled = false;

const config = {
	diffEditor: { renderSideBySide: false },
	workbench: {
		quickOpen: {
			closeOnFocusLost: false,
		},
		editor: {
			enablePreview: false,
		},
		activityBar: {
			visible: false,
		},
		statusBar: {
			visible: false,
		},
	},
	explorer: {
		openEditors: {
			visible: 0,
		},
	},
	editor: {
		readOnly: !Features.zap2Way.isEnabled(),
		tabSize: 4,
		automaticLayout: true,
		scrollBeyondLastLine: false,
		wrappingColumn: 0,
		fontFamily: code_font_face,
		fontSize: 15,
		lineHeight: 21,
		theme: "vs-dark",
		renderLineHighlight: "none",
		codeLens: codeLensEnabled,
		glyphMargin: false,
		hideCursorInOverviewRuler: true,
		selectionHighlight: false,
	},
	files: {
		eol: "\n",
	}
};

export function toggleCodeLens(): void {
	codeLensEnabled = !codeLensEnabled;
	config.editor.codeLens = codeLensEnabled;
	_onDidUpdateConfiguration.fire({ config } as any);
}

export function isCodeLensEnabled(): boolean {
	return codeLensEnabled;
}

export function updateConfiguration(updater: (config: any) => void): void {
	updater(config);
	_onDidUpdateConfiguration.fire({ config } as any);
}

export class ConfigurationService implements IConfigurationService {
	_serviceBrand: any;

	getConfiguration<T>(section?: string): T {
		if (!section) { return config as any; }
		return getConfigurationValue<T>(config, section);
	}

	lookup<T>(key: string): IConfigurationValue<T> {
		const value = getConfigurationValue<T>(config, key);
		return {
			value: value,
			default: value,
			user: value,
		};
	}

	keys(): IConfigurationKeys { return { default: ["zap.enable"] as string[], user: [] as string[] }; }

	reloadConfiguration<T>(section?: string): TPromise<T> { return TPromise.as({} as T); }

	onDidUpdateConfiguration: Event<IConfigurationServiceEvent> = onDidUpdateConfiguration;
}

export class WorkspaceConfigurationService extends ConfigurationService implements IWorkspaceConfigurationService {
	hasWorkspaceConfiguration(): boolean { return false; }

	lookup<T>(key: string): IWorkspaceConfigurationValue<T> {
		const value = super.lookup<T>(key);
		return {
			...value,
			workspace: undefined as any,
		};
	}

	keys(): IWorkspaceConfigurationKeys {
		const keys = super.keys();
		return {
			default: keys.default,
			user: keys.user,
			workspace: [],
		};
	}

	values(): IWorkspaceConfigurationValues { return {}; }
}
