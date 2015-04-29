# JBrowse for dictyBase
Source repository for docker image to run jbrowse at [dictyBase](http://dictybase.org).

## Features
* Uses the flatfile backend of JBrowse.
* Comes in two flavors __end to end__(developmental) and __data only__(production deploy).
* Comes with its own static file server that runs at port 9595.
* Dynamic update of JBrowse config files.
* Based on design principle of [radial](https://github.com/radial/docs) topology. 
* JBrowse config, log and data are managed using [data containers](http://docs.docker.com/userguide/dockervolumes/).

## Installation
* Install [docker](https://docs.docker.com/installation/#installation).
* Install [docker-compose](http://docs.docker.com/compose/install/).

## Quickstart
* Clone this repository

```
git clone https://github.com/dictybase-docker/wheel-jbrowse.git
```

* Download JBrowse data inside a ``data``` subfolder.

```
mkdir -p $PWD/data && curl -L -o data/jbrowse-data.tar.gz https://northwestern.box.com/shared/static/xw7sfh72mnreb4nwwkjbj8lvzo3il9dq.gz
```

* Start JBrowse application

```
docker-compose -f jbrowse_full.yml up -d
```

* Open browser at http://localhost:9595

## Detail guide
In order to understand the containerization strategy for JBrowse,
understanding the structure and concept of JBrowse application is
important. 

### JBrowse concepts
* Source code contains both the perl scripts(for flat file backend) and client
  side files(for display).
* It is recommended to have the flat file data inside the source code checkout
  folder. Symlinking is an option for mapping data that are outside of source
  code folder tree.
* There is one global config file(jbrowse.conf) in the root of the source
  tree. There can multiple config files for every genome(tracks.conf) to
  configure the display of tracks. Each config file could also have a _json_
  counterpart(jbrowse.json and trackList.json).
* Needs a static file server for display.

The gory details of most of the concepts are described in the
JBrowse [guide](http://gmod.org/wiki/JBrowse_Configuration_Guide).


### Strategy for containerization
#### Data containers
Data container concepts from radial topology is borrowed to managed various
parts of JBrowse application. Here are the list of data container volumes and
their application

* `/log` : Static web server log from from frontend container.
* `/config`: Contain the JBrowse configuration files in ```/config/jbrowse```
  folder. It will be mapped the `config` subfolder of this repository in the
  host.
* `/data`: Contains the JBrowse JSON formatted data for the flat file backend.
* `/ngs`: Contains the data files(bam,bigwig etc) from NGS experiments. It
  will map to ```/mnt/ngs``` folder of the host.

#### Flat file backend
This backend of JBrowse needs bunch of perl scripts to prepare JBrowse
compatible JSON files from various biological data sources(GFF3, Fasta).
The [backend
container](https://github.com/dictybase-docker/wheel-jbrowse/tree/master/spoke-jbrowse-backend) 
handles this transformation. This transformation needs a database backend
which in this case is served by a custom [postgresql
container](https://github.com/dictybase-docker/wheel-jbrowse/tree/master/spoke-jbrowse-postgresql).

#### JBrowse application
The application is handled by the
[frontend](https://github.com/dictybase-docker/wheel-jbrowse/tree/master/spoke-jbrowse-frontend)
container. The image contains the application in ```/usr/src/jbrowse``` folder. The data folder ```/usr/src/jbrowse/data```
folder is symlinked to ```/data/jbrowse``` where all JSON formatted files are
kept. The frontend container runs a static file server (port 9595) to run the
JBrowse application.

#### Configuration files
JBrowse has a local(jbrowse.conf) and genome specific
configuration(tracks.conf) files both of which are kept in the
[config](https://github.com/dictybase-docker/wheel-jbrowse/tree/master/config/jbrowse)
subfolder of this repository. This folder, through docker volume mapping gets
exposed to ```/config```  folder inside the [frontend
container](https://github.com/dictybase-docker/wheel-jbrowse/tree/master/spoke-jbrowse-frontend).
The [config
manager](https://github.com/dictybase-docker/wheel-jbrowse/tree/master/spoke-jbrowse-tracks-conf)
container copies all files from ```/config``` folder to the jbrowse
application folder. It also runs a file watcher that copies any of the any
updated configuration files from ```/config``` folder. The ```jbrowse.conf```
gets copied to jbrowse source folder ```/usr/src/jbrowse```. The track config
files gets copied to the data folder( ```/data/jbrowse```) of the respective
genomes. The location of genome subfolders mapping is kept in the __dataset__
key of a [yml
configuration](https://github.com/dictybase-docker/wheel-jbrowse/blob/master/config/jbrowse/export.yml)
file.

#### NGS data
The NGS(next generation sequence) data is expected to be present in
```/mnt/ngs``` folder of host OS. The ```/mnt/ngs``` folder is made available
through `/ngs` volume through a docker data container. The NGS data is also
expected to follow a folder structure like this ...

```
    rnaseq
    ├── PRJNA118577
    │   ├── bam
    │   └── bw
    └── PRJNA143419
        ├── bam
            └── bw
```

The __RNA-Seq__ data goes inside a top level `rnaseq` folder. Each dataset
will be inside a folder named after its _study accession_ number. Bam and
bigwig files will be inside their respective folders. To make it available in
JBrowse data folder, a symlink is created between `/data/jbrowse/rnaseq` and
`/ngs/rnaseq` subfolders. In effect the NGS data gets available inside JBrowse
data folder.

### Starting containers
The containers could be started
using two different methods, one is __end to end__ and the other is __data only__.

#### End to End
In this setup data is generated from GFF3 files through a temporary postgresql
database. The data generation is done through a set of perl scripts that
shipped with JBrowse. 

```
docker-compose -f jbrowse_full.yml up -d
```

#### Data only
Here the data generation process is skipped and instead a copy of the generated
data is used directly for running jbrowse. 

```
docker-compose -f jbrowse_data_only.yml up -d
```




