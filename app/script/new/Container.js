import FluxStoreGroup from "flux/lib/FluxStoreGroup";

import Component from "./Component";

class Container extends Component {
	componentDidMount() {
		let stores = this.stores();

		let changed = false;
		let setChanged = () => { changed = true; };
		this._containerSubscriptions = stores.map((store) => store.addListener(setChanged));

		this._containerStoreGroup = new FluxStoreGroup(stores, () => {
			if (changed) {
				this.setState({});
			}
			changed = false;
		});
	}

	componentWillUnmount() {
		this._containerStoreGroup.release();
		this._containerSubscriptions.forEach((subscription) => { subscription.remove(); });
	}
}

export default Container;
