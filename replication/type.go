package replication

import (
	"math"
)

// ValueType of redis type
type ValueType byte

// type value
const (
	TypeString  ValueType = 0
	TypeList    ValueType = 1
	TypeSet     ValueType = 2
	TypeZSet    ValueType = 3
	TypeHash    ValueType = 4
	TypeZSet2   ValueType = 5
	TypeModule  ValueType = 6
	TypeModule2 ValueType = 7

	TypeHashZipmap      ValueType = 9
	TypeListZiplist     ValueType = 10
	TypeSetIntset       ValueType = 11
	TypeZSetZiplist     ValueType = 12
	TypeHashZiplist     ValueType = 13
	TypeListQuicklist   ValueType = 14
	TypeStreamListPacks ValueType = 15
)

const (
	rdbVersion  = 9
	rdb6bitLen  = 0
	rdb14bitLen = 1
	rdb32bitLen = 0x80
	rdb64bitLen = 0x81
	rdbEncVal   = 3
	rdbLenErr   = math.MaxUint64

	rdbOpCodeModuleAux = 247
	rdbOpCodeIdle      = 248
	rdbOpCodeFreq      = 249
	rdbOpCodeAux       = 250
	rdbOpCodeResizeDB  = 251
	rdbOpCodeExpiryMS  = 252
	rdbOpCodeExpiry    = 253
	rdbOpCodeSelectDB  = 254
	rdbOpCodeEOF       = 255

	rdbModuleOpCodeEOF    = 0
	rdbModuleOpCodeSint   = 1
	rdbModuleOpCodeUint   = 2
	rdbModuleOpCodeFloat  = 3
	rdbModuleOpCodeDouble = 4
	rdbModuleOpCodeString = 5

	rdbLoadNone  = 0
	rdbLoadEnc   = (1 << 0)
	rdbLoadPlain = (1 << 1)
	rdbLoadSds   = (1 << 2)

	rdbSaveNode        = 0
	rdbSaveAofPreamble = (1 << 0)

	rdbEncInt8  = 0
	rdbEncInt16 = 1
	rdbEncInt32 = 2
	rdbEncLZF   = 3

	rdbZiplist6bitlenString  = 0
	rdbZiplist14bitlenString = 1
	rdbZiplist32bitlenString = 2

	rdbZiplistInt16 = 0xc0
	rdbZiplistInt32 = 0xd0
	rdbZiplistInt64 = 0xe0
	rdbZiplistInt24 = 0xf0
	rdbZiplistInt8  = 0xfe
	rdbZiplistInt4  = 15

	rdbLpHdrSize           = 6
	rdbLpHdrNumeleUnknown  = math.MaxUint16
	rdbLpMaxIntEncodingLen = 0
	rdbLpMaxBacklenSize    = 5
	rdbLpMaxEntryBacklen   = 34359738367
	rdbLpEncodingInt       = 0
	rdbLpEncodingString    = 1

	rdbLpEncoding7BitUint     = 0
	rdbLpEncoding7BitUintMask = 0x80

	rdbLpEncoding6BitStr     = 0x80
	rdbLpEncoding6BitStrMask = 0xC0

	rdbLpEncoding13BitInt     = 0xC0
	rdbLpEncoding13BitIntMask = 0xE0

	rdbLpEncoding12BitStr     = 0xE0
	rdbLpEncoding12BitStrMask = 0xF0

	rdbLpEncoding16BitInt     = 0xF1
	rdbLpEncoding16BitIntMask = 0xFF

	rdbLpEncoding24BitInt     = 0xF2
	rdbLpEncoding24BitIntMask = 0xFF

	rdbLpEncoding32BitInt     = 0xF3
	rdbLpEncoding32BitIntMask = 0xFF

	rdbLpEncoding64BitInt     = 0xF4
	rdbLpEncoding64BitIntMask = 0xFF

	rdbLpEncoding32BitStr     = 0xF0
	rdbLpEncoding32BitStrMask = 0xFF

	rdbLpEOF                     = 0xFF
	rdbStreamItemFlagNone        = 0        /* No special flags. */
	rdbStreamItemFlagDeleted     = (1 << 0) /* Entry was deleted. Skip it. */
	rdbStreamItemFlangSameFields = (1 << 1) /* Same fields as master entry. */
)

type CommandType string

const (
	RDB CommandType = "rdb"

	Ping CommandType = "ping"
	Pong CommandType = "pong"
	Ok   CommandType = "ok"
	Err  CommandType = "err"

	Empty     CommandType = "empty"
	Undefined CommandType = "undefined"

	Select    CommandType = "select"
	Zadd      CommandType = "zadd"
	Sadd      CommandType = "sadd"
	Zrem      CommandType = "zrem"
	Delete    CommandType = "delete"
	Lpush     CommandType = "lpush"
	LpushX    CommandType = "lpushx"
	Rpush     CommandType = "rpush"
	RpushX    CommandType = "rpushx"
	Rpop      CommandType = "rpop"
	RpopLpush CommandType = "rpoplpush"
	Lpop      CommandType = "lpop"

	ZremRangeByLex   CommandType = "zremrangebylex"
	ZremRangeByRank  CommandType = "zremrangebyrank"
	ZremRangeByScore CommandType = "zremrangebyscore"
	ZunionStore      CommandType = "zunionstore"
	ZincrBy          CommandType = "zincrby"
	SdiffStore       CommandType = "sdiffstore"
	SinterStore      CommandType = "sinterstore"
	Smove            CommandType = "smove"
	SunionStore      CommandType = "sunionstore"
	Srem             CommandType = "srem"
	Set              CommandType = "set"
	SetBit           CommandType = "setbit"

	Append       CommandType = "append"
	BitField     CommandType = "bitfield"
	BitOp        CommandType = "bitop"
	Decr         CommandType = "decr"
	DecrBy       CommandType = "decrby"
	Incr         CommandType = "incr"
	IncrBy       CommandType = "incrby"
	IncrByFloat  CommandType = "incrbyfloat"
	GetSet       CommandType = "getset"
	Mset         CommandType = "mset"
	MsetNX       CommandType = "msetnx"
	SetEX        CommandType = "setex"
	SetNX        CommandType = "setnx"
	SetRange     CommandType = "setrange"
	BlPop        CommandType = "blpop"
	BrPop        CommandType = "brpop"
	BrPopLpush   CommandType = "brpoplpush"
	Linsert      CommandType = "linsert"
	Lrem         CommandType = "lrem"
	Lset         CommandType = "lset"
	Ltrim        CommandType = "ltrim"
	Expire       CommandType = "expire"
	ExpireAt     CommandType = "expireat"
	Pexpire      CommandType = "pexpire"
	PexpireAt    CommandType = "pexpireat"
	Move         CommandType = "move"
	Persist      CommandType = "persist"
	Rename       CommandType = "rename"
	Restore      CommandType = "restore"
	Hset         CommandType = "hset"
	HsetNx       CommandType = "hsetnx"
	HmSet        CommandType = "hmset"
	HincrBy      CommandType = "hincrby"
	HincrByFloat CommandType = "hincrbyfloat"
	PfAdd        CommandType = "pfadd"
	PfMerge      CommandType = "pfmerge"
	PsetX        CommandType = "psetx"
)

var CommandTypeMap = map[string]CommandType{
	"select":           Select,
	"zadd":             Zadd,
	"sadd":             Sadd,
	"zrem":             Zrem,
	"delete":           Delete,
	"lpush":            Lpush,
	"lpushx":           LpushX,
	"rpush":            Rpush,
	"rpushx":           RpushX,
	"rpop":             Rpop,
	"rpoplpush":        RpopLpush,
	"lpop":             Lpop,
	"zremrangebylex":   ZremRangeByLex,
	"zremrangebyrank":  ZremRangeByRank,
	"zremrangebyscore": ZremRangeByScore,
	"zunionstore":      ZunionStore,
	"zincrby":          ZincrBy,
	"sdiffstore":       SdiffStore,
	"sinterstore":      SinterStore,
	"smove":            Smove,
	"sunionstore":      SunionStore,
	"srem":             Srem,
	"set":              Set,
	"setbit":           SetBit,
	"append":           Append,
	"bitfield":         BitField,
	"bitop":            BitOp,
	"decr":             Decr,
	"decrby":           DecrBy,
	"incr":             Incr,
	"incrby":           IncrBy,
	"incrbyfloat":      IncrByFloat,
	"getset":           GetSet,
	"mset":             Mset,
	"msetnx":           MsetNX,
	"setex":            SetEX,
	"setnx":            SetNX,
	"setrange":         SetRange,
	"blpop":            BlPop,
	"brpop":            BrPop,
	"brpoplpush":       BrPopLpush,
	"linsert":          Linsert,
	"lrem":             Lrem,
	"lset":             Lset,
	"ltrim":            Ltrim,
	"expire":           Expire,
	"expireat":         ExpireAt,
	"pexpire":          Pexpire,
	"pexpireat":        PexpireAt,
	"move":             Move,
	"persist":          Persist,
	"rename":           Rename,
	"restore":          Restore,
	"hset":             Hset,
	"hsetnx":           HsetNx,
	"hmset":            HmSet,
	"hincrby":          HincrBy,
	"hincrbyfloat":     HincrByFloat,
	"pfadd":            PfAdd,
	"pfmerge":          PfMerge,
	"psetx":            PsetX,
}
