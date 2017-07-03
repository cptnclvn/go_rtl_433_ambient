package main

//Imported Packages
import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/influxdata/influxdb/client/v2"
)

//Global Variables
var (
	protocol = flag.String("protocol", "20", "Protocol to enable")
	cmdPath  = flag.String("cmdPath", "rtl_433", "full path for rtl_433")
	debug    = flag.Bool("debug", false, "set debug")

	namedOnly = flag.Bool("namedOnly", false, "Only insert named sensors. See named nameFields")

	influxUsername = flag.String("influxUsername", "admin", "influxDB Username")
	influxPassword = flag.String("influxPassword", "admin", "influxDB Password")
	influxURL      = flag.String("influxURL", "http://influxdb:8086", "influxDB URL, disabled if empty")
	influxDatabase = flag.String("influxDatabase", "temperature_db", "influx Database name")
)

func init() {
}

func main() {
	// Named Device
	var nameFieldsFlag fieldFlag

	//Map(Hash Table) of the "namedOnly" arguement
	nameFields := make(map[int64]string)
	flag.Var(&nameFieldsFlag, "nameFields", "List of id=name pairs (comma separated) to  be injected as a name label eg 1251=kitchen")

	flag.Parse()

	// parsing args for naming devices
	for _, field := range nameFieldsFlag.Fields {
		if len(strings.Split(field, "=")) != 2 {
			fmt.Println("Invalid forceField", field)
			flag.PrintDefaults()
			os.Exit(2)
		}
		split := strings.Split(field, "=")
		deviceID, err := strconv.ParseInt(split[0], 10, 64)
		if err != nil {
			fmt.Println("Invalid forceField shoud be an int", field)
			flag.PrintDefaults()
			os.Exit(2)
		}
		nameFields[deviceID] = split[1]
	}

	if *namedOnly && len(nameFields) == 0 {
		fmt.Println("namedOnly is filtering all the sensors")
		flag.PrintDefaults()
		os.Exit(2)
	}

	var influxClient client.Client
	if *influxURL != "" {
		var err error
		influxClient, err = client.NewHTTPClient(client.HTTPConfig{
			Addr:     *influxURL,
			Username: *influxUsername,
			Password: *influxPassword,
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	cmd := exec.Command(*cmdPath, "-R", *protocol, "-F", "json", "-q")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// read command's stdout line by line
	in := bufio.NewScanner(stdout)

	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  *influxDatabase,
		Precision: "s",
	})

	for in.Scan() {
		var msg DeviceMessage
		if err := json.Unmarshal([]byte(in.Text()), &msg); err != nil {
			log.Println(err)
			continue
		}
		// add names labels
		if name, ok := nameFields[int64(msg.ID)]; ok {
			msg.Name = name
		} else {
			if *namedOnly {
				if *debug {
					log.Println("Skipped sensors because of namedOnly", msg.ID)
				}
				continue
			}
		}

		if influxClient != nil {
			bp.AddPoint(msg.ToInfluxPoint())
			if err := influxClient.Write(bp); err != nil {
				log.Println(err)
			}
		}
		if *debug {
			log.Println(msg)
		}
	}
	if err := in.Err(); err != nil {
		log.Printf("error: %s", err)
	}

}
