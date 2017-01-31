import { code_font_face } from "sourcegraph/components/styles/_vars.css";
import Event, { Emitter } from "vs/base/common/event";
import { IConfigurationServiceEvent, IConfigurationValue, getConfigurationValue } from "vs/platform/configuration/common/configuration";

const _onDidUpdateConfiguration = new Emitter<IConfigurationServiceEvent>();
const onDidUpdateConfiguration = _onDidUpdateConfiguration.event;

let codeLensEnabled = false;

const config = {
	workbench: {
		quickOpen: {
			closeOnFocusLost: false,
		},
		editor: {
			enablePreview: false,
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
		readOnly: true,
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

export class ConfigurationService {
	onDidUpdateConfiguration: Event<IConfigurationServiceEvent> = onDidUpdateConfiguration;
	getConfiguration(key: string): any {
		return config;
	}

	lookup(key: string): IConfigurationValue<any> {
		return {
			value: getConfigurationValue(config, key),
			default: getConfigurationValue(config, key),
			user: getConfigurationValue(config, key),
		};
	}

}
