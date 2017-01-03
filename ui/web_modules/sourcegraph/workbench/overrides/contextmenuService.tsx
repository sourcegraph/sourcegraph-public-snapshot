import { ContextMenuService as VSContextMenuService } from "vs/platform/contextview/browser/contextMenuService";
import { IContextViewService } from "vs/platform/contextview/browser/contextView";
import { IMessageService } from "vs/platform/message/common/message";
import { ITelemetryService } from "vs/platform/telemetry/common/telemetry";

export class ContextMenuService extends VSContextMenuService {
	constructor(
		@IMessageService messageService: IMessageService,
		@ITelemetryService telemetryService: ITelemetryService,
		@IContextViewService contextViewService: IContextViewService
	) {
		const element = document.getElementsByName("body")[0];
		super(element, telemetryService, messageService, contextViewService);
	}
}
