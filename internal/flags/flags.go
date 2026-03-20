package flags

import "flag"

var Flag Flags

type Flags struct {
	From int
	To   int
	Port int
}

func init() {
	flag.IntVar(&Flag.From, "from", 0, "for i := from; i < to; i++")
	flag.IntVar(&Flag.To, "to", 0, "for i := from; i < to; i++")
	flag.IntVar(&Flag.Port, "port", 8080, "server port")

	flag.Parse()
}
