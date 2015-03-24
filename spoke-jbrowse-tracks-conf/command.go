package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	zmq "github.com/pebbe/zmq4"

	"gopkg.in/codegangsta/cli.v1"
	"gopkg.in/yaml.v2"
)

type TracksConf struct {
	Config []string `yaml:"config"`
}

func main() {
	app := cli.NewApp()
	app.Name = "tracks"
	app.Version = "1.0.0"
	app.Usage = "Command line app to manage jbrowse tracks configurations"
	app.Commands = []cli.Command{
		{
			Name:   "copy",
			Usage:  "Copy tracks.conf to jbrowse data directory",
			Action: CopyAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "config-file",
					Usage:  "Yaml configuration file for copying tracks.conf",
					EnvVar: "TRACKS_CONF",
					Value:  "/config/jbrowse/tracks.yml",
				},
				cli.StringFlag{
					Name:  "receive-name, rn",
					Usage: "IP address to receive the completion of name indexing",
					Value: "name:9996",
				},
			},
		},
	}
	app.Run(os.Args)
}

func CopyAction(c *cli.Context) {
	b, err := ioutil.ReadFile(c.String("config-file"))
	if err != nil {
		log.Fatal(err)
	}
	var tconf TracksConf
	err = yaml.Unmarshal(b, &tconf)
	if err != nil {
		log.Fatal(err)
	}
	// Wait for the completion before start the copying
	ReceiveMsg(c.String("receive-name"), "name:complete")
	// copy all the tracks configuration files
	for _, cs := range tconf.Config {
		cl := strings.Split(cs, "=")
		r, err := os.Open(cl[0])
		if err != nil {
			log.Fatal(err)
		}
		defer r.Close()
		w, err := os.Create(filepath.Join(cl[1], "tracks.conf"))
		if err != nil {
			log.Fatal(err)
		}
		defer w.Close()
		_, err = io.Copy(w, r)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("copied %s file to %s", cl[0], cl[1])
	}
}

func ReceiveMsg(receive string, msg string) {
	ctx, err := zmq.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	socket, _ := ctx.NewSocket(zmq.PULL)
	defer socket.Close()
	err = socket.Connect(fmt.Sprintf("tcp://%s", receive))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("waiting to receive message %s\n", msg)
	rcv, err := socket.Recv(0)
	if err != nil {
		log.Fatalf("did not receive any message %s\n", err)
	}
	if rcv == msg {
		log.Printf("recived msg %s\n", msg)
		return
	}
	log.Fatalf("did not receive import message %s\n", msg)

}
