var React = require("react");
var globals = require("../globals");
var $ = require("jquery");
var UserSSHKey = require("./UserSSHKey");

var UserSSHKeys = React.createClass({

	getInitialState() {
		return {
			keys: [],
			newKey: "",
			newKeyName: "",
		};
	},

	componentDidMount() {
		$.ajax({
			method: "GET",
			url: "/.ui/.user/keys",
		}).done(this.handleResponse);
	},

	onDelete(key) {
		$.ajax({
			method: "DELETE",
			url: `/.ui/.user/keys?ID=${key.ID}`,
			headers: {
				"X-CSRF-Token": globals.CsrfToken,
			},
		}).done(this.handleResponse);
	},

	handleResponse(data) {
		if (data.Results) {
			this.setState({
				keys: data.Results,
			});
		}
	},

	onSave(e) {
		e.preventDefault();

		$.ajax({
			method: "POST",
			url: "/.ui/.user/keys",
			data: JSON.stringify({
				Key: this.state.newKey,
				Name: this.state.newKeyName,
			}),
			dataType: "json",
			headers: {
				"X-CSRF-Token": globals.CsrfToken,
			},
		}).done(this.handleResponse);
	},

	onKeyUp(e) {
		var key, name;
		if (e.target.tagName === "TEXTAREA") {
			key = e.target.value;
			name = this.state.newKeyName;

			// If the key wasn't give a name, use the e-mail
			if (!name.length) {
				var keyParts = key.split(" ");

				if (keyParts.length === 3) {
					key = keyParts.slice(0, 2).join(" ");
					name = keyParts[2];
				}
			}
		} else {
			name = e.target.value;
			key = this.state.newKey;
		}

		this.setState({
			newKeyName: name,
			newKey: key,
		});
	},

	render() {
		return (
			<div className="list-group">
				{
					this.state.keys.map(function(k) {
						return <UserSSHKey key={k.ID} onDelete={this.onDelete} SSHKey={k} />;
					}.bind(this))
				}

				<br/>
				<form role="form" method="post">
					<h4>Add a new key</h4>

					<div className="list-group-item">
						<div className="form-group" style={{marginBottom: "5px"}}>
							<label htmlFor="userKeyName">Key Name</label>
							<input id="userKeyName" className="form-control" type="text" onChange={this.onKeyUp} value={this.state.newKeyName} placeholder="My Macbook Pro" />
						</div>

						<div className="form-group">
							<label htmlFor="userKeyValue">Public SSH Key</label>
							<textarea id="userKeyValue" className="form-control" cols="6" onChange={this.onKeyUp} style={{width: "100%"}} defaultValue="ssh-rsa ..." value={this.state.newKey}></textarea>
						</div>

						<div className="text-right">
							<button type="button" className="btn btn-sgblue" onClick={this.onSave}>Save</button>
						</div>
					</div>
				</form>
			</div>
		);
	},
});

module.exports = UserSSHKeys;
