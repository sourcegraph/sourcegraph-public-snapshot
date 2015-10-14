module.exports = {
	WantFile(repo, rev, tree) {
		this.repo = repo;
		this.rev = rev;
		this.tree = tree;
	},

	FileFetched(repo, rev, tree, file) {
		this.repo = repo;
		this.rev = rev;
		this.tree = tree;
		this.file = file;
	},
};
