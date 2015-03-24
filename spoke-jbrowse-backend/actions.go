package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	zmq "github.com/pebbe/zmq4"
	"gopkg.in/codegangsta/cli.v1"
	"gopkg.in/yaml.v1"
)

type TrackListConfig struct {
	DatasetId     string `json:"dataset_id"`
	FormatVersion int64  `json:"formatVersion"`
	Names         struct {
		Type string `json:"type"`
		Url  string `json:"url"`
	} `json:"names"`
	Tracks []struct {
		Autocomplete string   `json:"autocomplete"`
		Category     string   `json:"category"`
		ChunkSize    int64    `json:"chunkSize"`
		Compress     int64    `json:"compress"`
		Description  int64    `json:"description"`
		Feature      []string `json:"feature"`
		Key          string   `json:"key"`
		Label        string   `json:"label"`
		StoreClass   string   `json:"storeClass"`
		Style        struct {
			ClassName string `json:"className"`
		} `json:"style"`
		Subfeatures bool   `json:"subfeatures"`
		Track       string `json:"track"`
		Type        string `json:"type"`
		UrlTemplate string `json:"urlTemplate"`
	} `json:"tracks"`
}

type JBrowseConfig struct {
	DbAdaptor string `json:"db_adaptor"`
	DbArgs    struct {
		Adaptor string `json:"-adaptor"`
		Dsn     string `json:"-dsn"`
		Pass    string `json:"-pass"`
		User    string `json:"-user"`
	} `json:"db_args"`
	Description   string `json:"description"`
	TRACKDEFAULTS struct {
		Autocomplete string `json:"autocomplete"`
		Class        string `json:"class"`
	} `json:"TRACK DEFAULTS"`
	Tracks []struct {
		ArrowheadClass    string   `json:"arrowheadClass"`
		Category          string   `json:"category"`
		Class             string   `json:"class"`
		Feature           []string `json:"feature"`
		Key               string   `json:"key"`
		SubfeatureClasses struct {
			Exon string `json:"exon"`
		} `json:"subfeature_classes"`
		Subfeatures bool   `json:"subfeatures"`
		Track       string `json:"track"`
	} `json:"tracks"`
}

type ExportConfig struct {
	RefSeq  []string `yaml:"refseq"`
	BioDB   []string `yaml:"biodb"`
	Dataset []string `yaml:"dataset"`
}

func ExportAction(c *cli.Context) {
	b, err := ioutil.ReadFile(c.String("export-config"))
	if err != nil {
		log.Fatalf("Unable to read file %s\n", err)
	}
	var exconf ExportConfig
	err = yaml.Unmarshal(b, &exconf)
	if err != nil {
		log.Fatal("Error in reading yaml, %s\n", err)
	}
	var cmd string
	var cfiles []string
	var broadcast string
	var msg string
	if c.Bool("feature") {
		cmd = "bin/biodb-to-json.pl"
		cfiles = exconf.BioDB
		broadcast = c.String("broadcast-feature")
		msg = "feature:complete"
		ReceiveMsg(c.String("receive-refseq"), "refseq:complete")
	} else {
		cmd = "bin/prepare-refseqs.pl"
		cfiles = exconf.RefSeq
		broadcast = c.String("broadcast-refseq")
		msg = "refseq:complete"
		ReceiveMsg(c.String("receive-import"), "import:complete")
	}
	wg := new(sync.WaitGroup)
	//	 Export reference features for each of the config file
	for _, cs := range cfiles {
		refs := strings.Split(cs, "=")
		wg.Add(1)
		go RunExportCmd(cmd, filepath.Join(c.String("config-folder"), refs[0]), refs[1], wg, c.String("script-folder"))
	}
	wg.Wait()
	//if c.Bool("refseq") {
	//for _, dstr := range exconf.Dataset {
	//sl := strings.Split(dstr, "=")
	//wg.Add(1)
	//go AddDatasetID(sl[0], sl[1], wg)
	//}
	//}
	//wg.Wait()
	PublishMsg(broadcast, msg)
}

func GenerateAction(c *cli.Context) {
	b, err := ioutil.ReadFile(c.String("export-config"))
	if err != nil {
		log.Fatalf("Unable to read file %s\n", err)
	}
	var exconf ExportConfig
	err = yaml.Unmarshal(b, &exconf)
	if err != nil {
		log.Fatal("Error in reading yaml, %s\n", err)
	}
	cmd := "bin/generate-names.pl"
	cfiles := exconf.RefSeq
	msg := "feature:complete"
	ReceiveMsg(c.String("receive-feature"), msg)
	wg := new(sync.WaitGroup)
	//	 Export reference features for each of the config file
	for _, cs := range cfiles {
		refs := strings.Split(cs, "=")
		wg.Add(1)
		go RunGenCmd(cmd, refs[1], wg, c.String("script-folder"))
	}
	wg.Wait()
	PublishMsg(c.String("broadcast-name"), "name:complete")
}

func ImportAction(c *cli.Context) {
	ValidateImportArgs(c)
	time.Sleep(3000 * time.Millisecond)
	b, err := ioutil.ReadFile(c.String("config"))
	if err != nil {
		log.Fatal(err)
	}
	var jc JBrowseConfig
	err = json.Unmarshal(b, &jc)
	if err != nil {
		log.Fatal(err)
	}
	gcf := GetCanonicalGFF3(c.String("gff-folder"))
	if len(gcf) == 0 {
		log.Fatalf("No gff3 file(s) found in %s\n", c.String("gff-folder"))
	}
	opt := map[string]string{
		"fast":      "empty",
		"create":    "empty",
		"nosummary": "empty",
		"adaptor":   jc.DbArgs.Adaptor,
		"user":      jc.DbArgs.User,
		"password":  jc.DbArgs.Pass,
		"dsn":       jc.DbArgs.Dsn,
	}
	RunBioSeqFeatureCmd(opt, gcf)
	if c.Bool("publish") {
		PublishMsg(c.String("broadcast-import"), "import:complete")
	}
}

func FetchAction(c *cli.Context) {
	if len(c.Args()) == 0 {
		log.Fatal("Need an http url to download file")
	}
	url := c.Args()[0]
	out := filepath.Join(c.String("folder"), c.String("output"))
	w, err := os.Create(out)
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error in getting %s url: %s\n", url, err)
	}
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Downloaded file from %s url\n", url)

	if c.Bool("decompress") {
		zr, err := zip.OpenReader(out)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range zr.File {
			zipped, err := f.Open()
			if err != nil {
				log.Fatal(err)
			}
			defer zipped.Close()
			dpath := filepath.Join(c.String("folder"), f.Name)
			if f.FileInfo().IsDir() {
				os.MkdirAll(dpath, f.Mode())
			} else {
				w, err := os.OpenFile(dpath, os.O_WRONLY|os.O_CREATE, f.Mode())
				if err != nil {
					log.Fatal(err)
				}
				defer w.Close()
				if _, err := io.Copy(w, zipped); err != nil {
					log.Fatal(err)
				}
			}
		}
		log.Printf("Decompressed the file %s\n", out)
		if c.Bool("remove-after") {
			if err := os.Remove(out); err != nil {
				log.Fatal(err)
			}
		}
	}

}

// Get all gff3 files whose names start with canonical
// The file that starts with the name canonical_gff3 becomes the first element
// and become the first file to get loaded
func GetCanonicalGFF3(path string) []string {
	ginfo, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	files := make([]string, 0)
	cidx := 0
	midx := 1
	for i, g := range ginfo {
		if strings.HasPrefix(g.Name(), "canonical") && strings.HasSuffix(g.Name(), "gff3") {
			if strings.HasPrefix(g.Name(), "canonical_core") {
				cidx = i
			}
			if strings.Contains(g.Name(), "mitochondrial") {
				midx = i
			}
			files = append(files, filepath.Join(path, g.Name()))
		}
	}
	if len(files) > 1 && cidx != 0 { // need to have more than one file
		// swap the core gff3 file to be the first element
		ctemp := files[cidx]
		files[cidx] = files[0]
		files[0] = ctemp
	}
	if len(files) > 2 && midx != 1 { // need to have at least more than two files
		// swap the mito gff3 file to be the second element
		mtemp := files[midx]
		files[midx] = files[1]
		files[1] = mtemp
	}

	return files
}

func ValidateImportArgs(c *cli.Context) {
	if !c.IsSet("config") {
		log.Fatal("Need a config file, option --config is not set")
	}
	if !c.IsSet("gff-folder") {
		log.Fatal("Need location of gff-folder, option --gff-folder is not set")
	}
}

func RunBioSeqFeatureCmd(opt map[string]string, files []string) {
	p := make([]string, 0)
	for k, v := range opt {
		if v == "empty" {
			p = append(p, fmt.Sprint("--", k))
		} else {
			p = append(p, fmt.Sprint("--", k), v)
		}
	}
	p = append(p, files...)
	cmdline := strings.Join(p, " ")
	log.Printf("going to run %s\n", cmdline)
	b, err := exec.Command("bp_seqfeature_load.pl", p...).CombinedOutput()
	if err != nil {
		log.Printf("Status %s message %s\n", err.Error(), string(b))
	} else {
		log.Printf("finished running the command %s\n", cmdline)
	}
}

func RunExportCmd(cmd string, config string, outfolder string, wg *sync.WaitGroup, sf string) {
	defer wg.Done()
	err := os.Chdir(sf)
	if err != nil {
		log.Printf("Error in changing directory: %s\n", err)
		return
	}
	p := []string{cmd, "--conf", config, "--out", outfolder}
	cmdline := strings.Join(p, " ")
	log.Printf("going to run %s\n", cmdline)
	b, err := exec.Command("perl", p...).CombinedOutput()
	if err != nil {
		log.Printf("Status %s message %s\n", err.Error(), string(b))
	} else {
		log.Printf("finished running the command %s\n", cmdline)
	}
}

func RunGenCmd(cmd string, outfolder string, wg *sync.WaitGroup, sf string) {
	defer wg.Done()
	err := os.Chdir(sf)
	if err != nil {
		log.Printf("Error in changing directory: %s\n", err)
		return
	}
	p := []string{cmd, "--mem", "5000000", "--out", outfolder}
	cmdline := strings.Join(p, " ")
	log.Printf("going to run %s\n", cmdline)
	b, err := exec.Command("perl", p...).CombinedOutput()
	if err != nil {
		log.Printf("Status %s message %s\n", err.Error(), string(b))
	} else {
		log.Printf("finished running the command %s\n", cmdline)
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

func PublishMsg(broadcast string, msg string) {
	ctx, err := zmq.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	socket, _ := ctx.NewSocket(zmq.PUSH)
	defer socket.Close()
	socket.Bind(fmt.Sprintf("tcp://%s", broadcast))
	_, err = socket.Send(msg, 0)
	if err != nil {
		log.Fatalf("Error in sending message %s\n", msg)
	}
	log.Printf("Send message *%s* successfully\n", msg)
	// This timer is needed for the receiver to work, not sure why it is needed though
	time.Sleep(100 * time.Millisecond)
}

func AddDatasetID(id string, folder string, wg *sync.WaitGroup) {
	defer wg.Done()
	tf := filepath.Join(folder, "trackList.json")
	tbr, err := ioutil.ReadFile(tf)
	if err != nil {
		log.Fatal(err)
	}
	var tr TrackListConfig
	err = json.Unmarshal(tbr, &tr)
	if err != nil {
		log.Fatal(err)
	}
	tr.DatasetId = id
	tbw, err := json.Marshal(tr)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(tf, tbw, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
