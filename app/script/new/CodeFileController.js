var React = require("react");

var Dispatcher = require("./Dispatcher");
var CodeActions = require("./CodeActions");
var CodeStore = require("./CodeStore");
var CodeListing = require("./CodeListing");
require("./CodeBackend");

var CodeFileController = React.createClass({
	propTypes: {
		repo: React.PropTypes.string,
		rev: React.PropTypes.string,
		tree: React.PropTypes.string,
		startline: React.PropTypes.number,
		endline: React.PropTypes.number,
		token: React.PropTypes.number,
		unitType: React.PropTypes.string,
		unit: React.PropTypes.string,
		def: React.PropTypes.string,
		example: React.PropTypes.number,
	},

	getInitialState() {
		return {
			files: CodeStore.files,
		};
	},

	componentWillMount() {
		Dispatcher.dispatch(new CodeActions.WantFile(this.props.repo, this.props.rev, this.props.tree));
	},

	componentDidMount() {
		CodeStore.addListener(this._onStoreChange);
	},

	componentWillUnmount() {
		CodeStore.removeListener(this._onStoreChange);
	},

	_onStoreChange() {
		this.setState({
			files: CodeStore.files,
		});
	},

	render() {
		var file = this.state.files.get(this.props.repo, this.props.rev, this.props.tree);
		if (file === undefined) {
			return null;
		}
		return (
			<CodeListing lines={file.Entry.SourceCode.Lines} />
		);
	},
});

module.exports = CodeFileController;
