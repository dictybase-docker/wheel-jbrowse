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
	"gopkg.in/fsnotify.v1"
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
				cli.BoolFlag{
					Name:  "nowait",
					Usage: "Do not wait to receive message for indexing",
				},
			},
		},
		{
			Name:   "remove",
			Usage:  "Remove trackList.json file from jbrowse data directory",
			Action: RemoveAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "config-file",
					Usage:  "Yaml configuration file for removing trackList.json",
					EnvVar: "TRACKS_CONF",
					Value:  "/config/jbrowse/tracks.yml",
				},
			},
		},
		{
			Name:   "watchcopy",
			Usage:  "Watches and copy config files from jbrowse folder",
			Action: WatchCopyAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "config-file",
					Usage:  "Yaml configuration file for copying tracks.conf",
					EnvVar: "TRACKS_CONF",
					Value:  "/config/jbrowse/tracks.yml",
				},
				cli.StringFlag{
					Name:   "config-folder",
					Usage:  "Config folder to watch",
					Value:  "/config/jbrowse",
					EnvVar: "TRACKS_FOLDER",
				},
			},
		},
	}
	app.Run(os.Args)
}

func RemoveAction(c *cli.Context) {
	b, err := ioutil.ReadFile(c.String("config-file"))
	if err != nil {
		log.Fatal(err)
	}
	var tconf TracksConf
	err = yaml.Unmarshal(b, &tconf)
	if err != nil {
		log.Fatal(err)
	}
	// Go through the list and remove tracklist.json
	for _, cs := range tconf.Config {
		cl := strings.Split(cs, "=")
		err := os.Remove(filepath.Join(cl[1], "trackList.json"))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Removed trackList.json from %s", cl[1])
	}
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
	if !c.Bool("nowait") {
		// Wait for the completion before start the copying
		ReceiveMsg(c.String("receive-name"), "name:complete")
	}
	// copy all the tracks configuration files
	CopyTracksConfig(tconf)
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

func WatchCopyAction(c *cli.Context) {
	b, err := ioutil.ReadFile(c.String("config-file"))
	if err != nil {
		log.Fatal(err)
	}
	var tconf TracksConf
	err = yaml.Unmarshal(b, &tconf)
	if err != nil {
		log.Fatal(err)
	}
	cmap := make(map[string]string)
	// create map of file name
	for _, cs := range tconf.Config {
		cl := strings.Split(cs, "=")
		cmap[cl[0]] = cl[1]
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Printf("Modified file %s", event.Name)
					if v, ok := cmap[event.Name]; ok {
						r, err := os.Open(event.Name)
						if err != nil {
							log.Fatal(err)
						}
						defer r.Close()
						w, err := os.Create(filepath.Join(v, "tracks.conf"))
						if err != nil {
							log.Fatal(err)
						}
						defer w.Close()
						_, err = io.Copy(w, r)
						if err != nil {
							log.Fatal(err)
						}
						log.Printf("copied %s file to %s", event.Name, v)
					}
				}
			case err := <-watcher.Errors:
				log.Println(err)
			}
		}

	}()
	log.Printf("Going to watch %s", c.String("config-folder"))
	err = watcher.Add(c.String("config-folder"))
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func CopyTracksConfig(tconf TracksConf) {
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
