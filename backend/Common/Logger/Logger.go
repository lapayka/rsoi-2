package Logger

import "log"

func GetLogger() *log.Logger {
	logger := log.Default()
	logger.SetFlags(log.Lshortfile)

	return logger
}
