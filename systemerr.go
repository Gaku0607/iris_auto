package iris

import (
	"fmt"
	"os"
	"time"
)

var ErrChan chan error = make(chan error)

func RecoveryPrint() {
	fmt.Println(<-ErrChan)
	time.Sleep(time.Second * 5)
	os.Exit(0)
}
