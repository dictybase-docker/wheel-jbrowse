# JBrowse for dictyBase
Source repository for docker image to run jbrowse at [dictyBase](http://dictybase.org)
The docker container setup is based on [radial](https://github.com/radial/docs)
topology. 

## Quick start(for the impatient)
* Install [docker](http://docker.io) and
[docker-compose](http://docs.docker.com/compose/install/)

* Run JBrowse from GFF3 files

```
docker-compose -f jbrowse_full.yml up -d
```

Open browser at http://localhost:9595


## Prerequisite
Install [docker](http://docker.io) and
[docker-compose](http://docs.docker.com/compose/install/). docker-compose is
used to manage the orchestration of multiple containers.

## Starting containers
The container setup will run JBrowse using the flatfile backend where the data
is served from custom formatted json files. The containers could be started
using two different methods, one is __end to end__ and the other is __data only__.

### End to End
In this setup data is generated from GFF3 files through a temporary postgresql
database. The data generation is done through a set of perl scripts that
shipped with JBrowse. 

```
docker-compose -f jbrowse_full.yml up -d
```

### Data only
Here the data generation process is skipped and instead a copy of the generated
data is used directly for running jbrowse. 

```
docker-compose -f jbrowse_data_only.yml up -d
```




