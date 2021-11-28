package setup

#dependency: {
	name:                    string
	onlyEmployees?:          bool
	instructions:            string
	comment?:                string
	requiresSGSetupRestart?: bool
	check:                   #check
}

#check: {
	name: string
	args: [...string]
	result: bool
}

[
	#dependency & {
		name: "git"
		check: {name: "checkInPath", args: ["git"]}
		instructions: "brew install git"
	},
	#dependency & {
		name: "docker"
		check: {name: "checkInPath", args: ["docker"]}
		instructions: "brew install --cask docker"
	},
	#dependency & {
		name: "gnu-sed"
		check: {name: "checkInPath", args: ["gsed"]}
		instructions: "brew install gnu-sed"
	},
	#dependency & {
		name: "comby"
		check: {name: "checkInPath", args: ["comby"]}
		instructions: "brew install comby"
	},
	#dependency & {
		name: "pcre"
		check: {name: "checkInPath", args: ["pcregrep"]}
		instructions: "brew install pcre"
	},
	#dependency & {
		name: "sqlite"
		check: {name: "checkInPath", args: ["sqlite3"]}
		instructions: "brew install sqlite"
	},
	#dependency & {
		name: "jq"
		check: {name: "checkInPath", args: ["jq"]}
		instructions: "brew install jq"
	},
	#dependency & {
		name: "bash"
		check: {name: "checkCommandOutputContains", args: ["bash --version", "version 5"]}
		instructions: "brew install bash"
	},
]
