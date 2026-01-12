package cmd 

import "flag"

var Flag Flags

type Flags struct {
	From int 
	To int 
	Latency int 
	Spectator bool
	SpectatorFrom int 
	SpectatorTo int
	Evil bool 
}

func init() {
	flag.IntVar(&Flag.From, "from", 0, "for i := from; i < to; i++")
	flag.IntVar(&Flag.To, "to", 0, "for i := from; i < to; i++")
	flag.IntVar(&Flag.Latency, "latency", 500, "time.Sleep(latency)")
	flag.BoolVar(&Flag.Spectator, "spectator", false, "fun mode")
	flag.BoolVar(&Flag.Evil, "evil", false, "evil mode")

	flag.Parse()

	if Flag.Spectator || Flag.Evil {
		Flag.SpectatorFrom = Flag.From 
		Flag.SpectatorTo = Flag.To
		Flag.From = 0
		Flag.To = 0
	}
}

