package main

import (
	"os"

	"gopkg.in/codegangsta/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Name = "Command line wrapper to export import data for jbrowse backend"
	app.Version = "1.0.0"
	app.Commands = []cli.Command{
		{
			Name:   "import",
			Usage:  "Import GFF3 into bioseqfeature postgresql database",
			Action: ImportAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "config, c",
					Usage:  "Name of jbrowse config file[required]",
					EnvVar: "CONFIG",
				},
				cli.StringFlag{
					Name:   "gff-folder, gf",
					Usage:  "Name of the folder from where all gff3 files will be picked up[requied]",
					EnvVar: "GFF_FOLDER",
				},
				cli.BoolFlag{
					Name:  "publish, p",
					Usage: "Flag to publish/notify a completion message through zeromq",
				},
				cli.StringFlag{
					Name:  "broadcast-import, bi",
					Usage: "IP address to publish the message, default is *:9999",
					Value: "*:9999",
				},
			},
		},
		{
			Name:   "export",
			Usage:  "Export features from bioseqfeature postgresql database to a jbrowse filesystem format",
			Action: ExportAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "config-folder, cf",
					Usage:  "Name of jbrowse config folder, all file with extension json will be picked up",
					EnvVar: "CONFIG_FOLDER",
					Value:  "/config/jbrowse",
				},
				cli.StringFlag{
					Name:   "export-config",
					Usage:  "Name of the export yaml config file that specifies folder for export",
					EnvVar: "EXPORT_CONFIG",
					Value:  "/config/jbrowse/export.yml",
				},
				cli.StringFlag{
					Name:  "broadcast-refseq, br",
					Usage: "IP address to publish the completion of reference feature export on zmq, default is *:9998",
					Value: "*:9998",
				},
				cli.StringFlag{
					Name:  "broadcast-feature, bf",
					Usage: "IP address to publish the completion of non-reference feature export on zmq, default is *:9997",
					Value: "*:9997",
				},
				cli.StringFlag{
					Name:  "receive-import, ri",
					Usage: "IP address to receive the completion of import GFF3 import, default is importer:9999",
					Value: "importer:9999",
				},
				cli.StringFlag{
					Name:  "receive-refseq, rr",
					Usage: "IP address to receive the completion of reference feature export, default is refseq:9998",
					Value: "refseq:9998",
				},
				cli.BoolFlag{
					Name:  "feature, f",
					Usage: "Export non-reference feature, default is off",
				},
				cli.BoolFlag{
					Name:  "refseq, r",
					Usage: "Export reference features, default is on",
				},
				cli.StringFlag{
					Name:  "script-folder, sf",
					Usage: "jbrowse backend perl scripts folder",
					Value: "/usr/src/jbrowse",
				},
			},
		},
		{
			Name:   "generate",
			Usage:  "Build a index of feature name in the jbrowse export data folers",
			Action: GenerateAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "export-config",
					Usage:  "Name of the export yaml config file that specifies various data folders",
					EnvVar: "EXPORT_CONFIG",
					Value:  "/config/jbrowse/export.yml",
				},
				cli.StringFlag{
					Name:  "receive-feature, rf",
					Usage: "IP address to receive the completion of feature export, default is feature:9998",
					Value: "feature:9997",
				},
				cli.StringFlag{
					Name:  "script-folder, sf",
					Usage: "jbrowse backend perl scripts folder",
					Value: "/usr/src/jbrowse",
				},
				cli.StringFlag{
					Name:  "broadcast-name, bn",
					Usage: "IP address to publish the completion of name indexing on zmq",
					Value: "*:9996",
				},
			},
		},
		{
			Name:   "fetch",
			Usage:  "Fetches file from HTTP endpoint, mostly meant for getting gff3 file for import process",
			Action: FetchAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "folder, f",
					Usage: "Directory where the file will be downloaded",
					Value: "/data",
				},
				cli.StringFlag{
					Name:  "output, o",
					Usage: "Name of the output file",
					Value: "downloaded",
				},
				cli.BoolFlag{
					Name:  "decompress, d",
					Usage: "Decompress the file in the folder after downloading, only zip file is allowed, default is false",
				},
				cli.BoolFlag{
					Name:  "remove-after, rf",
					Usage: "Remove the downloaded file after decompression, works only if decompress is true",
				},
			},
		},
	}
	app.Run(os.Args)
}
