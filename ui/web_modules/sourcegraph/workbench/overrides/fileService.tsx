import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { IFileService, IFilesConfiguration, IResolveFileOptions, IFileStat, IContent, IStreamContent, IImportResult, IResolveContentOptions, IUpdateContentOptions } from "vs/platform/files/common/files";
import * as electronService from "vs/workbench/services/files/electron-browser/fileService";
import * as nodeService from "vs/workbench/services/files/node/fileService";


class FileService {
	constructor() {
		//
	}

	resolveFile(resource: URI, options?: IResolveFileOptions): TPromise<IFileStat> {
		return TPromise.wrap({
			resource,
			name: resource.fragment,
		});
	}
}

nodeService.FileService = FileService;
electronService.FileService = FileService;
