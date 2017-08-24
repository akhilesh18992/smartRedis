package flags

import "flag"

var StatsdHostPort string
var Prefix string
var RedisHost string
var RedisPort string
var Action string
var ResolveHostname string

func Init()  {
	flag.StringVar(&StatsdHostPort, "statsd", "", "statsd host port")
	flag.StringVar(&Prefix, "prefix", "smartRedis", "statsd prefix")
	flag.StringVar(&RedisHost, "redisHost", "localhost", "redis cluster host")
	flag.StringVar(&RedisPort, "redisPort", "6379", "redis cluster port")
	flag.StringVar(&ResolveHostname, "resolveHostname", "n", "resolve hostname")
	flag.StringVar(&Action, "action", "", "action you want to perfrom (status, statsd, create-cluster)")
	flag.Parse()
}


