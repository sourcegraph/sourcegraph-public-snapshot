export class CreateIssue {
	constructor(repo, path, commitID, startLine, endLine, title, body, callback) {
		this.repo = repo;
		this.path = path;
		this.commitID = commitID;
		this.startLine = startLine;
		this.endLine = endLine;
		this.title = title;
		this.body = body;
		this.callback = callback;
	}
}
