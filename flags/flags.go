package flags

import "flag"

var StatsdHostPort string
var StatsdPushInterval int
var Prefix string
var RedisHost string
var RedisPort string
var Action string
var ResolveHostname string
var RepeatInterval int

func Init() {
	flag.StringVar(&Action, "action", "", "action you want to perfrom (status, statsd, create-cluster)")
	flag.StringVar(&StatsdHostPort, "statsd", "", "statsd host port")
	flag.StringVar(&Prefix, "prefix", "smartRedis", "statsd prefix")
	flag.StringVar(&RedisHost, "redisHost", "localhost", "redis cluster host")
	flag.StringVar(&RedisPort, "redisPort", "6379", "redis cluster port")
	flag.StringVar(&ResolveHostname, "resolveHostname", "n", "resolve hostname")
	flag.IntVar(&StatsdPushInterval, "statsdPushInterval", 60, "statsd metrics push interval in secs")
	flag.IntVar(&RepeatInterval, "repeatInterval", 0, "repeat status fetch interval in secs")
	flag.Parse()
}
