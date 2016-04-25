package sphinx

const DefaultIndex = "*"

// Sent to / received from server to ensure same version.
const MAJOR_PROTOCOL_VERSION = 1

// Command versions from sphinxclient.c
const (
	VER_COMMAND_EXCERPT  = 0x103
	VER_COMMAND_UPDATE   = 0x102
	VER_COMMAND_KEYWORDS = 0x100
	VER_COMMAND_STATUS   = 0x100
	// Version of the sphinx client as of 2.0.8 release
	VER_COMMAND_SEARCH = 0x119
)

type MatchMode int

// Matching modes from sphinxclient.h
const (
	SPH_MATCH_ALL MatchMode = iota
	SPH_MATCH_ANY
	SPH_MATCH_PHRASE
	SPH_MATCH_BOOLEAN
	SPH_MATCH_EXTENDED
	SPH_MATCH_FULLSCAN
	SPH_MATCH_EXTENDED2
)

type RankMode int

// Ranking modes from sphinxclient.h
const (
	SPH_RANK_PROXIMITY_BM25 RankMode = iota // Default mode, phrase proximity major factor and BM25 minor one
	SPH_RANK_BM25
	SPH_RANK_NONE
	SPH_RANK_WORDCOUNT
	SPH_RANK_PROXIMITY
	SPH_RANK_MATCHANY
	SPH_RANK_FIELDMASK
	SPH_RANK_SPH04
	_ // Don't support SPH_RANK_EXPR since PHP driver does not
	SPH_RANK_TOTAL

	SPH_RANK_DEFAULT = SPH_RANK_PROXIMITY_BM25
)

type SortMode int

// Sorting modes, also from sphinxclient.h
const (
	SPH_SORT_RELEVANCE SortMode = iota
	SPH_SORT_ATTR_DESC
	SPH_SORT_ATTR_ASC
	SPH_SORT_TIME_SEGMENTS
	SPH_SORT_EXTENDED
	SPH_SORT_EXPR // Deprecated, never use it.
)

// Grouping functions from sphinxclient.h
const (
	SPH_GROUPBY_DAY = iota
	SPH_GROUPBY_WEEK
	SPH_GROUPBY_MONTH
	SPH_GROUPBY_YEAR
	SPH_GROUPBY_ATTR
	SPH_GROUPBY_ATTRPAIR

	// Default Groupby option and function
	SPH_GROUPBY_FUNC_DEFAULT = SPH_GROUPBY_ATTR
	SPH_GROUPBY_DEFAULT      = ""
	SPH_GROUPBY_SORT_DEFAULT = "@groupby desc"
)

// Searchd status codes from sphinxclient.h
const (
	SEARCHD_OK = iota
	SEARCHD_ERROR
	SEARCHD_RETRY
	SEARCHD_WARNING
)

// Attribute types from sphinxclient.h
const (
	_ = iota // Starts at 1, so ignore 0
	SPH_ATTR_INTEGER
	SPH_ATTR_TIMESTAMP
	SPH_ATTR_ORDINAL
	SPH_ATTR_BOOL
	SPH_ATTR_FLOAT
	SPH_ATTR_BIGINT
	SPH_ATTR_STRING

	SPH_ATTR_MULTI   = 0x40000001
	SPH_ATTR_MULTI64 = 0x40000002
)

type Command int

// Searchd commands from sphinxclient.c
const (
	SEARCHD_COMMAND_SEARCH Command = iota
	SEARCHD_COMMAND_EXCERPT
	SEARCHD_COMMAND_UPDATE
	SEARCHD_COMMAND_KEYWORDS
	SEARCHD_COMMAND_PERSIST
	SEARCHD_COMMAND_STATUS
)

type Filter int

// Filter values from sphinxclient.h
const (
	SPH_FILTER_VALUES Filter = iota
	SPH_FILTER_RANGE
	SPH_FILTER_FLOATRANGE
)

// Define true/false values from sphinxclient.h
const (
	SPH_FALSE = iota
	SPH_TRUE
)
