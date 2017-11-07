package main
import(
  "os"
  "strconv"
  "github.com/Sirupsen/logrus"
  "github.com/docker/go-plugins-helpers/volume"
)
func main() {
  debug := os.Getenv("DEBUG")
	if ok, _ := strconv.ParseBool(debug); ok {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
