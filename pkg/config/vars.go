package config

type Structure struct {
	Port        string
	RpcUri      string
	MongoDbName string
	MongoUri    string
	RedisUri    string
}

var (
	Config *Structure
)
