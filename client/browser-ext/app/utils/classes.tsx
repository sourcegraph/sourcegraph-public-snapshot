export class PhabricatorInstance {
	constructor(private uriToRepoUrlMapping: { [key: string]: string; }, private phabStagingUri: string) { }

	getPhabricatorRepoFromMap(repoUri: string): string | undefined {
		repoUri = repoUri.toLowerCase();
		return this.uriToRepoUrlMapping[repoUri];
	}

	getStagingRepoUriFromRepoUrl(repoUrl: string): string | undefined {
		return repoUrl;
	}

}
