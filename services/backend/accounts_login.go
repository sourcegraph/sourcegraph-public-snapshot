package backend

import (
	"regexp"
	"strings"
)

var validLoginRE = regexp.MustCompile(`^[a-zA-Z0-9_]{3,50}$`)

// isValidUsername returns whether name is a valid login that can
// be used for an account on Sourcegraph.
func isValidLogin(login string) bool {
	if !validLoginRE.MatchString(login) {
		return false
	}

	login = strings.ToLower(login)
	for _, u := range reservedLogins {
		if u == login {
			return false
		}
	}
	return true
}

// reservedLogins is a reasonable list of logins that nobody should be
// able to register an account under. These logins might be used to
// mislead potential visitors, if they see a URL like
// https://administrator.sourcegraph.com or
// https://sourcegraph.com/groups (if in the future we make those URLs
// go to the profile pages of those users).
//
// These should all be lowercase (for comparison in isValidLogin).
//
// Taken from https://gist.github.com/caseyohara/1453705.
var reservedLogins = []string{
	"supportdetails",
	"support-details",

	// Generic reserved words
	"about",
	"account",
	"activate",
	"add",
	"admin",
	"administrator",
	"api",
	"app",
	"apps",
	"archive",
	"archives",
	"auth",
	"blog",
	"cache",
	"cancel",
	"careers",
	"cart",
	"changelog",
	"checkout",
	"codereview",
	"compare",
	"config",
	"configuration",
	"connect",
	"contact",
	"create",
	"codex",
	"delete",
	"direct_messages",
	"doc",
	"docs",
	"documentation",
	"download",
	"downloads",
	"edit",
	"email",
	"employment",
	"enterprise",
	"event",
	"events",
	"faq",
	"favorites",
	"feed",
	"feedback",
	"feeds",
	"fleet",
	"fleets",
	"follow",
	"followers",
	"following",
	"friend",
	"friends",
	"gist",
	"group",
	"groups",
	"help",
	"home",
	"hosting",
	"hostmaster",
	"html",
	"idea",
	"ideas",
	"index",
	"info",
	"invitations",
	"invite",
	"is",
	"it",
	"job",
	"jobs",
	"json",
	"language",
	"languages",
	"lists",
	"login",
	"logout",
	"logs",
	"mail",
	"map",
	"maps",
	"market",
	"marketplace",
	"mine",
	"mis",
	"news",
	"oauth",
	"oauth_clients",
	"offers",
	"openid",
	"order",
	"orders",
	"organizations",
	"plans",
	"platform",
	"popular",
	"post",
	"postmaster",
	"pricing",
	"privacy",
	"projects",
	"put",
	"recruitment",
	"register",
	"remove",
	"replies",
	"root",
	"rss",
	"sales",
	"save",
	"search",
	"security",
	"sessions",
	"settings",
	"shop",
	"signup",
	"sitemap",
	"ssl",
	"ssladmin",
	"ssladministrator",
	"sslwebmaster",
	"status",
	"stories",
	"styleguide",
	"subscribe",
	"subscriptions",
	"support",
	"sysadmin",
	"sysadministrator",
	"terms",
	"tour",
	"translations",
	"trends",
	"unfollow",
	"unsubscribe",
	"update",
	"url",
	"user",
	"weather",
	"webmaster",
	"widget",
	"widgets",
	"wiki",
	"ww",
	"www",
	"wwww",
	"xfn",
	"xml",
	"xmpp",
	"yaml",
	"yml",

	// Top 50 languages by speaking population
	"chinese",
	"mandarin",
	"spanish",
	"english",
	"bengali",
	"hindi",
	"portuguese",
	"russian",
	"japanese",
	"german",
	"wu",
	"javanese",
	"korean",
	"french",
	"vietnamese",
	"telugu",
	"marathi",
	"tamil",
	"turkish",
	"urdu",
	"min-nan",
	"jinyu",
	"gujarati",
	"polish",
	"arabic",
	"ukrainian",
	"italian",
	"xiang",
	"malayalam",
	"hakka",
	"kannada",
	"oriya",
	"panjabi",
	"sunda",
	"romanian",
	"bhojpuri",
	"azerbaijani",
	"farsi",
	"maithili",
	"hausa",
	"burmese",
	"serbo-croatian",
	"gan",
	"awadhi",
	"thai",
	"dutch",
	"yoruba",
	"sindhi",

	// Country TLDs
	"ac", // Ascension Island
	"ad", // Andorra
	"ae", // United Arab Emirates
	"af", // Afghanistan
	"ag", // Antigua and Barbuda
	"ai", // Anguilla
	"al", // Albania
	"am", // Armenia
	"an", // Netherlands Antilles
	"ao", // Angola
	"aq", // Antarctica
	"ar", // Argentina
	"as", // American Samoa
	"at", // Austria
	"au", // Australia
	"aw", // Aruba
	"ax", // and
	"az", // Azerbaijan
	"ba", // Bosnia and Herzegovina
	"bb", // Barbados
	"bd", // Bangladesh
	"be", // Belgium
	"bf", // Burkina Faso
	"bg", // Bulgaria
	"bh", // Bahrain
	"bi", // Burundi
	"bj", // Benin
	"bm", // Bermuda
	"bn", // Brunei Darussalam
	"bo", // Bolivia
	"br", // Brazil
	"bs", // Bahamas
	"bt", // Bhutan
	"bv", // Bouvet Island
	"bw", // Botswana
	"by", // Belarus
	"bz", // Belize
	"ca", // Canada
	"cc", // Cocos (Keeling) Islands
	"cd", // Democratic Republic of the Congo
	"cf", // Central African Republic
	"cg", // Republic of the Congo
	"ch", // Switzerland
	"ci", // CÃ´te d"Ivoire
	"ck", // Cook Islands
	"cl", // Chile
	"cm", // Cameroon
	"cn", // People"s Republic of China
	"co", // Colombia
	"cr", // Costa Rica
	"cs", // Czechoslovakia
	"cu", // Cuba
	"cv", // Cape Verde
	"cx", // Christmas Island
	"cy", // Cyprus
	"cz", // Czech Republic
	"dd", // East Germany
	"de", // Germany
	"dj", // Djibouti
	"dk", // Denmark
	"dm", // Dominica
	"do", // Dominican Republic
	"dz", // Algeria
	"ec", // Ecuador
	"ee", // Estonia
	"eg", // Egypt
	"eh", // Western Sahara
	"er", // Eritrea
	"es", // Spain
	"et", // Ethiopia
	"eu", // European Union
	"fi", // Finland
	"fj", // Fiji
	"fk", // Falkland Islands
	"fm", // Federated States of Micronesia
	"fo", // Faroe Islands
	"fr", // France
	"ga", // Gabon
	"gb", // United Kingdom
	"gd", // Grenada
	"ge", // Georgia
	"gf", // French Guiana
	"gg", // Guernsey
	"gh", // Ghana
	"gi", // Gibraltar
	"gl", // Greenland
	"gm", // The Gambia
	"gn", // Guinea
	"gp", // Guadeloupe
	"gq", // Equatorial Guinea
	"gr", // Greece
	"gs", // South Georgia and the South Sandwich Islands
	"gt", // Guatemala
	"gu", // Guam
	"gw", // Guinea-Bissau
	"gy", // Guyana
	"hk", // Hong Kong
	"hm", // Heard Island and McDonald Islands
	"hn", // Honduras
	"hr", // Croatia
	"ht", // Haiti
	"hu", // Hungary
	"id", // Indonesia
	"ie", // Republic of Ireland  Northern Ireland
	"il", // Israel
	"im", // Isle of Man
	"in", // India
	"io", // British Indian Ocean Territory
	"iq", // Iraq
	"ir", // Iran
	"is", // Iceland
	"it", // Italy
	"je", // Jersey
	"jm", // Jamaica
	"jo", // Jordan
	"jp", // Japan
	"ke", // Kenya
	"kg", // Kyrgyzstan
	"kh", // Cambodia
	"ki", // Kiribati
	"km", // Comoros
	"kn", // Saint Kitts and Nevis
	"kp", // Democratic People"s Republic of Korea
	"kr", // Republic of Korea
	"kw", // Kuwait
	"ky", // Cayman Islands
	"kz", // Kazakhstan
	"la", // Laos
	"lb", // Lebanon
	"lc", // Saint Lucia
	"li", // Liechtenstein
	"lk", // Sri Lanka
	"lr", // Liberia
	"ls", // Lesotho
	"lt", // Lithuania
	"lu", // Luxembourg
	"lv", // Latvia
	"ly", // Libya
	"ma", // Morocco
	"mc", // Monaco
	"md", // Moldova
	"me", // Montenegro
	"mg", // Madagascar
	"mh", // Marshall Islands
	"mk", // Republic of Macedonia
	"ml", // Mali
	"mm", // Myanmar
	"mn", // Mongolia
	"mo", // Macau
	"mp", // Northern Mariana Islands
	"mq", // Martinique
	"mr", // Mauritania
	"ms", // Montserrat
	"mt", // Malta
	"mu", // Mauritius
	"mv", // Maldives
	"mw", // Malawi
	"mx", // Mexico
	"my", // Malaysia
	"mz", // Mozambique
	"na", // Namibia
	"nc", // New Caledonia
	"ne", // Niger
	"nf", // Norfolk Island
	"ng", // Nigeria
	"ni", // Nicaragua
	"nl", // Netherlands
	"no", // Norway
	"np", // Nepal
	"nr", // Nauru
	"nu", // Niue
	"nz", // New Zealand
	"om", // Oman
	"pa", // Panama
	"pe", // Peru
	"pf", // French Polynesia
	"pg", // Papua New Guinea
	"ph", // Philippines
	"pk", // Pakistan
	"pl", // Poland
	"pm", // Saint-Pierre and Miquelon
	"pn", // Pitcairn Islands
	"pr", // Puerto Rico
	"ps", // Palestinian territories
	"pt", // Portugal
	"pw", // Palau
	"py", // Paraguay
	"qa", // Qatar
	"re", // RÃ©union
	"ro", // Romania
	"rs", // Serbia
	"ru", // Russia
	"rw", // Rwanda
	"sa", // Saudi Arabia
	"sb", // Solomon Islands
	"sc", // Seychelles
	"sd", // Sudan
	"se", // Sweden
	"sg", // Singapore
	"sh", // Saint Helena
	"si", // Slovenia
	"sj", // Svalbard and  Jan Mayen Islands
	"sk", // Slovakia
	"sl", // Sierra Leone
	"sm", // San Marino
	"sn", // Senegal
	"so", // Somalia
	"sr", // Suriname
	"ss", // South Sudan
	"st", // SÃ£o TomÃ© and PrÃ­ncipe
	"su", // Soviet Union
	"sv", // El Salvador
	"sy", // Syria
	"sz", // Swaziland
	"tc", // Turks and Caicos Islands
	"td", // Chad
	"tf", // French Southern and Antarctic Lands
	"tg", // Togo
	"th", // Thailand
	"tj", // Tajikistan
	"tk", // Tokelau
	"tl", // East Timor
	"tm", // Turkmenistan
	"tn", // Tunisia
	"to", // Tonga
	"tp", // East Timor
	"tr", // Turkey
	"tt", // Trinidad and Tobago
	"tv", // Tuvalu
	"tw", // Republic of China (Taiwan)
	"tz", // Tanzania
	"ua", // Ukraine
	"ug", // Uganda
	"uk", // United Kingdom
	"us", // United States of America
	"uy", // Uruguay
	"uz", // Uzbekistan
	"va", // Vatican City
	"vc", // Saint Vincent and the Grenadines
	"ve", // Venezuela
	"vg", // British Virgin Islands
	"vi", // United States Virgin Islands
	"vn", // Vietnam
	"vu", // Vanuatu
	"wf", // Wallis and Futuna
	"ws", // Samoa
	"ye", // Yemen
	"yt", // Mayotte
	"yu", // Yugoslavia
	"za", // South Africa
	"zm", // Zambia
	"zw", // Zimbabwe
}
