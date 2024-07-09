package log

type Section string
type Event string
type ExtraKey string

const (
	Internal Section = "internal"
	General  Section = "General"
	Request  Section = "Request"
	Kafka    Section = "Kafka"
	Sentry   Section = "Sentry"
	Config   Section = "Config"
	Redis    Section = "Redis"
	Firebase Section = "Firebase"
	Database Section = "database"
	Mysql    Section = "Mysql"
	Rabbitmq Section = "Rabbitmq"
)

const (
	Startup Event = "Startup"
	Select  Event = "Select"
	Update  Event = "Update"
	Insert  Event = "Insert"
	Init    Event = "Init"
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
