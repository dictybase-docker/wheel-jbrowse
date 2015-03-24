FROM dictybase/postgres:9.3
MAINTAINER Siddhartha Basu<siddhartha-basu@northwestern.edu>

RUN mkdir -p /docker-entrypoint-initdb.d && mkdir -p /config
COPY *.conf /config/
COPY create_genome_user.sh /docker-entrypoint-initdb.d/ 
