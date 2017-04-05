import { code_font_face } from "sourcegraph/components/styles/_vars.css";
import Event, { Emitter } from "vs/base/common/event";
import { TPromise } from "vs/base/common/winjs.base";
import { DefaultConfig } from "vs/editor/common/config/defaultConfig";
import { IConfigurationKeys, IConfigurationOptions, IConfigurationService, IConfigurationServiceEvent, IConfigurationValue, getConfigurationValue } from "vs/platform/configuration/common/configuration";
import { IWorkspaceConfigurationKeys, IWorkspaceConfigurationService, IWorkspaceConfigurationValue, IWorkspaceConfigurationValues } from "vs/workbench/services/configuration/common/configuration";

import { Features } from "sourcegraph/util/features";
import { OpenSearchViewletAction } from "vscode/src/vs/workbench/parts/search/browser/searchActions";

const _onDidUpdateConfiguration = new Emitter<IConfigurationServiceEvent>();
export const onDidUpdateConfiguration = _onDidUpdateConfiguration.event;

let codeLensEnabled = false;

// Exclude common vendor directories from jump-to-file, for better UX
// and perf. This is how GitHub's "t" quick file picker works as well.
//
// NOTE: If you add an exclude entry that contains a glob, then
// defaultExcludesNoGlobs can no longer be used, because it relies on
// the defaultExcludesRegExp fast path. Not using it will
// significantly slow down quickopen jump-to-file, so you'll need to
// fix that.
//
// TODO(sqs): We could make this better by using GitHub linguist's
// standard list of vendor exclusions.
const defaultExcludesNoGlobs = {
	"node_modules": true,
	"bower_components": true,
	"vendor": true,
	"dist": true,
	"out": true,
	"Godeps": true,
	"third_party": true,
};

// This is the fastest way to match strings (faster than
// Strings.prototype.indexOf). See
// https://jsperf.com/regexp-indexof-perf.
//
// Matches any path containing a path component that is a key of
// defaultExcludesNoGlobs.
export const defaultExcludesRegExp = new RegExp("(/|^)(" + Object.keys(defaultExcludesNoGlobs).join("|") + ")/");

const config = {
	diffEditor: { renderSideBySide: false },
	workbench: {
		colorTheme: "vs-dark",
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
	zap: {
		enable: true,
		overwrite: false,
	},
	editor: {
		readOnly: !Features.zap2Way.isEnabled(),
		tabSize: 4,
		automaticLayout: true,
		scrollBeyondLastLine: false,
		wordWrap: "on",
		fontFamily: code_font_face,
		fontSize: 15,
		lineHeight: 21,
		theme: "vs-dark",
		renderLineHighlight: "none",
		codeLens: codeLensEnabled,
		glyphMargin: false,
		hideCursorInOverviewRuler: true,
		selectionHighlight: true,
	},
	files: {
		eol: "\n",
		exclude: defaultExcludesNoGlobs,
	},
	search: {
		exclude: defaultExcludesNoGlobs,
	},
	window: {
		title: "",
		titleBarStyle: "custom",
	},
	zenMode: {},
};

if (!Features.textSearch.isEnabled()) {
	OpenSearchViewletAction.prototype.run = () => TPromise.wrap(void 0);
}

DefaultConfig.editor.readOnly = config.editor.readOnly;

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

	getConfiguration<C>(section?: string): C;
	getConfiguration<C>(options?: IConfigurationOptions): C;
	getConfiguration<C>(arg?: any): C {
		if (!arg) { return config as any; }
		if (typeof arg === "string") {
			return getConfigurationValue<C>(config, arg);
		}
		return getConfigurationValue<C>(config, arg.section);
	}

	lookup<T>(key: string): IConfigurationValue<T> {
		const value = getConfigurationValue<T>(config, key);
		return {
			value: value,
			default: value,
			user: value,
		};
	}

	keys(): IConfigurationKeys { return { default: ["zap.enable", "zap.overwrite", "workbench.statusBar.visible"], user: [] as string[] }; }

	reloadConfiguration<T>(section?: string): TPromise<T> { return TPromise.as({} as T); }

	onDidUpdateConfiguration: Event<IConfigurationServiceEvent> = onDidUpdateConfiguration;
}

export class WorkspaceConfigurationService extends ConfigurationService implements IWorkspaceConfigurationService {
	hasWorkspaceConfiguration(): boolean { return false; }

	getUnsupportedWorkspaceKeys(): string[] { return []; }

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

	values(): IWorkspaceConfigurationValues {
		const result: IWorkspaceConfigurationValues = Object.create(null);
		const keyset = this.keys();
		const keys = [...keyset.workspace, ...keyset.user, ...keyset.default].sort();

		let lastKey: string | undefined;
		for (const key of keys) {
			if (key !== lastKey) {
				lastKey = key;
				result[key] = this.lookup(key);
			}
		}

		return result;
	}
}
