package main
import(
  "os"
  "strconv"
  "github.com/Sirupsen/logrus"
)
func main() {
  debug := os.Getenv("DEBUG")
	if ok, _ := strconv.ParseBool(debug); ok {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
