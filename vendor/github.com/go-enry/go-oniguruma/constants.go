package rubex

const (
	ONIG_OPTION_DEFAULT = ONIG_OPTION_NONE
	/* options */
	ONIG_OPTION_NONE               = 0
	ONIG_OPTION_IGNORECASE         = 1
	ONIG_OPTION_EXTEND             = (ONIG_OPTION_IGNORECASE << 1)
	ONIG_OPTION_MULTILINE          = (ONIG_OPTION_EXTEND << 1)
	ONIG_OPTION_SINGLELINE         = (ONIG_OPTION_MULTILINE << 1)
	ONIG_OPTION_FIND_LONGEST       = (ONIG_OPTION_SINGLELINE << 1)
	ONIG_OPTION_FIND_NOT_EMPTY     = (ONIG_OPTION_FIND_LONGEST << 1)
	ONIG_OPTION_NEGATE_SINGLELINE  = (ONIG_OPTION_FIND_NOT_EMPTY << 1)
	ONIG_OPTION_DONT_CAPTURE_GROUP = (ONIG_OPTION_NEGATE_SINGLELINE << 1)
	ONIG_OPTION_CAPTURE_GROUP      = (ONIG_OPTION_DONT_CAPTURE_GROUP << 1)
	/* options (search time) */
	ONIG_OPTION_NOTBOL       = (ONIG_OPTION_CAPTURE_GROUP << 1)
	ONIG_OPTION_NOTEOL       = (ONIG_OPTION_NOTBOL << 1)
	ONIG_OPTION_POSIX_REGION = (ONIG_OPTION_NOTEOL << 1)
	ONIG_OPTION_MAXBIT       = ONIG_OPTION_POSIX_REGION /* limit */

	ONIG_NORMAL   = 0
	ONIG_MISMATCH = -1

	ONIG_MISMATCH_STR                = "mismatch"
	ONIGERR_UNDEFINED_NAME_REFERENCE = -217
)
