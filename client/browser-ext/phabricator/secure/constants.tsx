import { PhabricatorInstance } from "../../app/utils/classes";

const securePhabricatorMap = {
	"arcanist": "github.com/phacility/arcanist",
	"arc": "github.com/phacility/arcanist",
};

export const securePhabricatorInstance = new PhabricatorInstance(securePhabricatorMap, "");
