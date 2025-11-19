package log

type Section string
type Event string
type ExtraKey string

const (
	Internal   Section = "internal"
	General    Section = "General"
	Request    Section = "Request"
	Kafka      Section = "Kafka"
	Sentry     Section = "Sentry"
	Config     Section = "Config"
	Mongodb    Section = "Mongodb"
	Redis      Section = "Redis"
	Firebase   Section = "Firebase"
	Database   Section = "database"
	Mysql      Section = "Mysql"
	Rabbitmq   Section = "Rabbitmq"
	Validation Section = "Validation"
)

const (
	Startup         Event = "Startup"
	Shutdown        Event = "Shutdown"
	Select          Event = "Select"
	Update          Event = "Update"
	Insert          Event = "Insert"
	Init            Event = "Init"
	RequestResponse Event = "RequestResponse"
)

// extra keys

const (
	AppName      ExtraKey = "AppName"
	LoggerName   ExtraKey = "LoggerName"
	ClientIp     ExtraKey = "ClientIp"
	HostIp       ExtraKey = "HostIp"
	Method       ExtraKey = "Method"
	StatusCode   ExtraKey = "StatusCode"
	BodySize     ExtraKey = "BodySize"
	Path         ExtraKey = "Path"
	Latency      ExtraKey = "Latency"
	Body         ExtraKey = "Body"
	RequestBody  ExtraKey = "RequestBody"
	ResponseBody ExtraKey = "ResponseBody"
	ErrorMessage ExtraKey = "ErrorMessage"
)
