package utils

import (
	"fmt"
	"strings"
)

func Log(msg interface{}, prefix ...string) {
	finalPrefix := strings.Join(prefix, "")
	fmt.Printf("%s %s\n", finalPrefix, msg)
}
