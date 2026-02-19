package preview

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/fatih/color"
)

var preview7 = `
▄▖             ▘          ▗     
▌ ▛▌  ▛▛▌▀▌▛▌  ▌▛▌▛▘▛▌█▌▛▘▜▘▛▌▛▘
▙▌▙▌  ▌▌▌█▌▙▌  ▌▌▌▄▌▙▌▙▖▙▖▐▖▙▌▌ 
           ▌        ▌  by hyperocheck
 
`

var preview6 = `
               ▘          ▗     
▛▌▛▌  ▛▛▌▀▌▛▌  ▌▛▌▛▘▛▌█▌▛▘▜▘▛▌▛▘
▙▌▙▌  ▌▌▌█▌▙▌  ▌▌▌▄▌▙▌▙▖▙▖▐▖▙▌▌ 
▄▌         ▌        ▌
 
`

var preview5 = `
         ▘          ▗     
▛▛▌▀▌▛▌  ▌▛▌▛▘▛▌█▌▛▘▜▘▛▌▛▘
▌▌▌█▌▙▌  ▌▌▌▄▌▙▌▙▖▙▖▐▖▙▌▌ 
     ▌        ▌
 
`

var preview2 = `
 ▗▄▄▖ ▄▄▄      ▄▄▄▄  ▗▞▀▜▌▄▄▄▄  ▄ ▄▄▄▄   ▄▄▄ ▄▄▄▄  ▗▞▀▚▖▗▞▀▘   ■   ▄▄▄   ▄▄▄ 
▐▌   █   █     █ █ █ ▝▚▄▟▌█   █ ▄ █   █ ▀▄▄  █   █ ▐▛▀▀▘▝▚▄▖▗▄▟▙▄▖█   █ █    
▐▌▝▜▌▀▄▄▄▀     █   █      █▄▄▄▀ █ █   █ ▄▄▄▀ █▄▄▄▀ ▝▚▄▄▖      ▐▌  ▀▄▄▄▀ █    
▝▚▄▞▘                     █     █            █                ▐▌             
                          ▀                  ▀                ▐▌             
`

var preview = `
▄▄▄▄  ▗▞▀▜▌▄▄▄▄  ▄ ▄▄▄▄   ▄▄▄ ▄▄▄▄  ▗▞▀▚▖▗▞▀▘   ■   ▄▄▄   ▄▄▄     
█ █ █ ▝▚▄▟▌█   █ ▄ █   █ ▀▄▄  █   █ ▐▛▀▀▘▝▚▄▖▗▄▟▙▄▖█   █ █        
█   █      █▄▄▄▀ █ █   █ ▄▄▄▀ █▄▄▄▀ ▝▚▄▄▖      ▐▌  ▀▄▄▄▀ █        
           █     █            █                ▐▌                 
           ▀                  ▀                ▐▌                 
`

func Preview() {
	lines := strings.Split(strings.TrimRight(preview7, "\n"), "\n")

	start := randomSoftColor()
	end := randomSoftColor()

	steps := len(lines) - 1
	if steps <= 0 {
		steps = 1
	}

	for i, line := range lines {
		r := start[0] + (end[0]-start[0])*i/steps
		g := start[1] + (end[1]-start[1])*i/steps
		b := start[2] + (end[2]-start[2])*i/steps

		color.RGB(r, g, b).Println(line)
	}
}

func SimplePreview() {
	fmt.Println(preview)
}

func randomSoftColor() [3]int {
	return [3]int{
		rand.Intn(120) + 80, // R: 80–200
		rand.Intn(120) + 80, // G: 80–200
		rand.Intn(120) + 80, // B: 80–200
	}
}
